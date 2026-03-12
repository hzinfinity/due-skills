package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hertz-contrib/registry/consul"
	"go-redis/redis/v9"
	"go.uber.org/zap"

	due "github.com/cloudwego/due"
	"github.com/cloudwego/due/cluster"
	"github.com/cloudwego/due/lock/redislock"
	"github.com/cloudwego/due/log/hlog"
	"github.com/cloudwego/due/pkg/setting"
	"github.com/cloudwego/due/rpc"
)

// Config 配置结构
type Config struct {
	ServerPort   int    `json:"server_port"`
	ServiceName  string `json:"service_name"`
	ServiceAddr  string `json:"service_addr"`
	ConsulAddr   string `json:"consul_addr"`
	RedisAddr    string `json:"redis_addr"`
	RedisPassword string `json:"redis_password"`
	RedisDB      int    `json:"redis_db"`
	MaxConn      int    `json:"max_conn"`
}

// WebSocketGateServer WebSocket 网关服务
type WebSocketGateServer struct {
	config       *Config
	server       *due.Server
	redisClient  *redis.Client
	consulRegistry *consul.ConsulRegistry
	upgrader     websocket.Upgrader
	connections  sync.Map // 存储所有 WebSocket 连接
	connCount    int64    // 当前连接数
	logger       *zap.Logger
	hub          *Hub // 连接管理中心
}

// Hub 连接管理中心
type Hub struct {
	connections map[string]*WebSocketConn // connID -> Connection
	mu          sync.RWMutex
	broadcast   chan []byte
	register    chan *WebSocketConn
	unregister  chan *WebSocketConn
}

// WebSocketConn WebSocket 连接封装
type WebSocketConn struct {
	connID    string
	conn      *websocket.Conn
	serverID  string
	userID    string
	sendChan  chan []byte
	hbTime    time.Time
	isOnline  bool
}

// Message 消息结构
type Message struct {
	Type      string      `json:"type"`
	ConnID    string      `json:"conn_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	ServerID  string      `json:"server_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
}

const (
	// 写入超时
	writeWait = 10 * time.Second

	//  pong 等待时间
	pongWait = 60 * time.Second

	// ping 周期
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512 * 1024
)

var (
	newline = []byte{'\n'}
)

// NewHub 创建连接管理中心
func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*WebSocketConn),
		broadcast:   make(chan []byte, 1000),
		register:    make(chan *WebSocketConn),
		unregister:  make(chan *WebSocketConn),
	}
}

// Run 运行 Hub
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.connections[conn.connID] = conn
			h.mu.Unlock()
			log.Printf("连接注册成功，connID: %s, 当前连接数：%d", conn.connID, len(h.connections))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn.connID]; ok {
				delete(h.connections, conn.connID)
				close(conn.sendChan)
			}
			h.mu.Unlock()
			log.Printf("连接注销成功，connID: %s, 当前连接数：%d", conn.connID, len(h.connections))

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, conn := range h.connections {
				select {
				case conn.sendChan <- message:
				default:
					// 发送失败，连接可能已关闭
				}
			}
			h.mu.RUnlock()
		}
	}
}

// NewConfig 创建默认配置
func NewConfig() *Config {
	return &Config{
		ServerPort:    8080,
		ServiceName:   "websocket-gate-server",
		ServiceAddr:   "",
		ConsulAddr:    "127.0.0.1:8500",
		RedisAddr:     "127.0.0.1:6379",
		RedisPassword: "",
		RedisDB:       0,
		MaxConn:       10000,
	}
}

// LoadFromFile 从文件加载配置
func (c *Config) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}

// NewWebSocketGateServer 创建 WebSocket 网关服务
func NewWebSocketGateServer(config *Config) (*WebSocketGateServer, error) {
	logger, _ := zap.NewProduction()

	server := &WebSocketGateServer{
		config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源
			},
		},
		connections: sync.Map{},
		logger:      logger,
		hub:         NewHub(),
	}

	return server, nil
}

// InitRedis 初始化 Redis 客户端
func (s *WebSocketGateServer) InitRedis() error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     s.config.RedisAddr,
		Password: s.config.RedisPassword,
		DB:       s.config.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis 连接失败：%w", err)
	}

	s.redisClient = rdb
	log.Println("Redis 连接成功")
	return nil
}

// InitConsul 初始化 Consul 注册中心
func (s *WebSocketGateServer) InitConsul() error {
	registry, err := consul.NewConsulRegistry(
		consul.WithConsulAddr(s.config.ConsulAddr),
		consul.WithTTL(10),
	)
	if err != nil {
		return fmt.Errorf("初始化 Consul 失败：%w", err)
	}

	s.consulRegistry = registry
	log.Println("Consul 注册中心初始化成功")
	return nil
}

// InitDueServer 初始化 due 服务器
func (s *WebSocketGateServer) InitDueServer() error {
	// 创建 Redis Locator
	redisLocator := cluster.NewRedisLocator(cluster.RedisLocatorConfig{
		Client: s.redisClient,
		Prefix: "due:websocket:locator",
	})

	// 创建 Redis Lock
	redisLock := redislock.NewLock(s.redisClient)

	// 创建 Noder
	noder, err := due.NewNoder(
		due.WithNoderID(s.config.ServiceAddr),
		due.WithLocator(redisLocator),
		due.WithLock(redisLock),
	)
	if err != nil {
		return fmt.Errorf("创建 Noder 失败：%w", err)
	}

	// 创建 RPC 服务器配置
	rpcConfig := &setting.RPC{
		Network:   "tcp",
		Addr:      fmt.Sprintf(":%d", s.config.ServerPort+1000), // RPC 端口
		Timeout:   setting.Duration{Duration: 5 * time.Second},
	}

	// 创建 due Server
	server, err := due.NewServer(
		noder,
		due.WithRPC(rpcConfig),
		due.WithLogger(hlog.NewLogger(hlog.WithLevel(hlog.LevelInfo))),
	)
	if err != nil {
		return fmt.Errorf("创建 due Server 失败：%w", err)
	}

	s.server = server
	log.Println("due Server 初始化成功")
	return nil
}

// RegisterService 注册服务到 Consul
func (s *WebSocketGateServer) RegisterService() error {
	if s.consulRegistry == nil {
		return fmt.Errorf("Consul 注册中心未初始化")
	}

	serviceInfo := &registry.Info{
		ServiceName: s.config.ServiceName,
		Addr:        s.config.ServiceAddr,
		Weight:      100,
		Tags:        []string{"websocket", "gate", "v2.5.2"},
	}

	err := s.consulRegistry.Register(serviceInfo)
	if err != nil {
		return fmt.Errorf("服务注册失败：%w", err)
	}

	log.Printf("服务注册成功：%s @ %s", s.config.ServiceName, s.config.ServiceAddr)
	return nil
}

// HandleWebSocket 处理 WebSocket 连接
func (s *WebSocketGateServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 检查连接数是否超过限制
	currentConn := s.connCount
	if currentConn >= int64(s.config.MaxConn) {
		http.Error(w, "连接数已达上限", http.StatusServiceUnavailable)
		return
	}

	// 升级 WebSocket 连接
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败：%v", err)
		return
	}

	// 生成连接 ID
	connID := fmt.Sprintf("conn_%d_%d", time.Now().UnixNano(), currentConn)

	// 创建 WebSocket 连接对象
	wsConn := &WebSocketConn{
		connID:   connID,
		conn:     conn,
		serverID: s.config.ServiceAddr,
		sendChan: make(chan []byte, 256),
		hbTime:   time.Now(),
		isOnline: true,
	}

	// 增加连接计数
	s.connCount++

	// 注册到 Hub
	s.hub.register <- wsConn

	// 将连接信息存储到 Redis
	ctx := context.Background()
	connInfo := map[string]interface{}{
		"conn_id":   connID,
		"server_id": s.config.ServiceAddr,
		"user_id":   "",
		"online":    true,
		"hb_time":   time.Now().Unix(),
	}
	connInfoJSON, _ := json.Marshal(connInfo)
	s.redisClient.HSet(ctx, fmt.Sprintf("due:websocket:connections:%s", connID), "info", connInfoJSON)
	s.redisClient.SAdd(ctx, "due:websocket:online:servers", s.config.ServiceAddr)

	// 发送欢迎消息
	welcomeMsg := Message{
		Type:      "welcome",
		ConnID:    connID,
		ServerID:  s.config.ServiceAddr,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"message":   "欢迎连接到 WebSocket 网关",
			"max_conn":  s.config.MaxConn,
			"curr_conn": s.connCount,
		},
	}
	welcomeData, _ := json.Marshal(welcomeMsg)
	wsConn.sendChan <- welcomeData

	log.Printf("新连接：connID=%s, 当前连接数=%d/%d", connID, s.connCount, s.config.MaxConn)

	// 启动连接处理器
	go s.handleConn(wsConn)
	go s.writePump(wsConn)
}

// handleConn 处理连接消息
func (s *WebSocketGateServer) handleConn(wsConn *WebSocketConn) {
	defer func() {
		s.closeConn(wsConn)
	}()

	wsConn.conn.SetReadLimit(maxMessageSize)
	wsConn.conn.SetReadDeadline(time.Now().Add(pongWait))
	wsConn.conn.SetPongHandler(func(string) error {
		wsConn.hbTime = time.Now()
		wsConn.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := wsConn.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 错误：%v", err)
			}
			break
		}

		// 处理消息
		s.processMessage(wsConn, message)
	}
}

// processMessage 处理接收到的消息
func (s *WebSocketGateServer) processMessage(wsConn *WebSocketConn, message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("消息解析失败：%v", err)
		return
	}

	ctx := context.Background()

	switch msg.Type {
	case "ping":
		// 响应 ping 消息
		pongMsg := Message{
			Type:      "pong",
			ConnID:    wsConn.connID,
			Timestamp: time.Now().Unix(),
		}
		pongData, _ := json.Marshal(pongMsg)
		wsConn.sendChan <- pongData

	case "bind_user":
		// 绑定用户
		if msg.UserID != "" {
			wsConn.userID = msg.UserID
			// 更新 Redis 中的用户映射
			s.redisClient.HSet(ctx, fmt.Sprintf("due:websocket:connections:%s", wsConn.connID), "user_id", msg.UserID)
			s.redisClient.Set(ctx, fmt.Sprintf("due:websocket:user:%s", msg.UserID), wsConn.connID, 0)

			bindResp := Message{
				Type:      "bind_user_ack",
				ConnID:    wsConn.connID,
				UserID:    msg.UserID,
				Timestamp: time.Now().Unix(),
				Data:      map[string]string{"status": "success"},
			}
			bindData, _ := json.Marshal(bindResp)
			wsConn.sendChan <- bindData
			log.Printf("用户绑定成功：userID=%s, connID=%s", msg.UserID, wsConn.connID)
		}

	case "message":
		// 转发消息
		s.forwardMessage(wsConn, msg.Data)

	default:
		log.Printf("未知消息类型：%s", msg.Type)
	}
}

// forwardMessage 转发消息
func (s *WebSocketGateServer) forwardMessage(wsConn *WebSocketConn, data interface{}) {
	// 这里可以实现消息的路由和转发逻辑
	// 例如：将消息转发给其他服务器上的用户
	log.Printf("收到消息：connID=%s, data=%v", wsConn.connID, data)
}

// writePump 写入泵
func (s *WebSocketGateServer) writePump(wsConn *WebSocketConn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.closeConn(wsConn)
	}()

	for {
		select {
		case message, ok := <-wsConn.sendChan:
			wsConn.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了通道
				wsConn.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := wsConn.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将队列中的其他消息也发送出去
			n := len(wsConn.sendChan)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-wsConn.sendChan)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			wsConn.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wsConn.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// closeConn 关闭连接
func (s *WebSocketGateServer) closeConn(wsConn *WebSocketConn) {
	if !wsConn.isOnline {
		return
	}

	wsConn.isOnline = false
	s.hub.unregister <- wsConn
	s.connCount--

	// 从 Redis 中删除连接信息
	ctx := context.Background()
	s.redisClient.Del(ctx, fmt.Sprintf("due:websocket:connections:%s", wsConn.connID))
	if wsConn.userID != "" {
		s.redisClient.Del(ctx, fmt.Sprintf("due:websocket:user:%s", wsConn.userID))
	}

	// 检查是否还有其他在线连接
	count, _ := s.redisClient.SCard(ctx, "due:websocket:online:servers").Result()
	if count == 0 {
		s.redisClient.Del(ctx, "due:websocket:online:servers")
	}

	wsConn.conn.Close()
	log.Printf("连接关闭：connID=%s, 当前连接数=%d", wsConn.connID, s.connCount)
}

// HandleRPCMessage 处理 RPC 消息（来自其他服务器）
func (s *WebSocketGateServer) HandleRPCMessage(ctx context.Context, msg *rpc.Message) error {
	// 处理来自其他 due 节点的消息
	var message Message
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		return err
	}

	// 查找目标连接并发送消息
	s.hub.mu.RLock()
	targetConn, ok := s.hub.connections[message.ConnID]
	s.hub.mu.RUnlock()

	if ok && targetConn.isOnline {
		data, _ := json.Marshal(message)
		targetConn.sendChan <- data
	}

	return nil
}

// Start 启动服务
func (s *WebSocketGateServer) Start() error {
	// 启动 Hub
	go s.hub.Run()

	// 创建 HTTP 服务器
	http.HandleFunc("/ws", s.HandleWebSocket)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "healthy",
			"conn_count":  s.connCount,
			"max_conn":    s.config.MaxConn,
			"server_id":   s.config.ServiceAddr,
			"timestamp":   time.Now().Unix(),
		})
	})
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_connections": s.connCount,
			"max_connections":   s.config.MaxConn,
			"server_id":         s.config.ServiceAddr,
			"service_name":      s.config.ServiceName,
			"timestamp":         time.Now().Unix(),
		})
	})

	// 启动 HTTP 服务器
	httpAddr := fmt.Sprintf(":%d", s.config.ServerPort)
	log.Printf("HTTP 服务器启动：%s", httpAddr)
	log.Printf("WebSocket 端点：ws://localhost:%d/ws", s.config.ServerPort)
	log.Printf("健康检查端点：http://localhost:%d/health", s.config.ServerPort)
	log.Printf("最大并发连接数：%d", s.config.MaxConn)

	go func() {
		if err := http.ListenAndServe(httpAddr, nil); err != nil {
			log.Fatalf("HTTP 服务器启动失败：%v", err)
		}
	}()

	// 启动 due RPC 服务器
	go func() {
		if err := s.server.Run(); err != nil {
			log.Fatalf("due RPC 服务器启动失败：%v", err)
		}
	}()

	log.Println("WebSocket 网关服务启动成功")
	return nil
}

// Stop 停止服务
func (s *WebSocketGateServer) Stop() error {
	log.Println("正在停止服务...")

	// 关闭所有连接
	s.hub.mu.RLock()
	for _, conn := range s.hub.connections {
		s.closeConn(conn)
	}
	s.hub.mu.RUnlock()

	// 从 Consul 注销服务
	if s.consulRegistry != nil {
		serviceInfo := &registry.Info{
			ServiceName: s.config.ServiceName,
			Addr:        s.config.ServiceAddr,
		}
		s.consulRegistry.Deregister(serviceInfo)
		log.Println("已从 Consul 注销服务")
	}

	// 关闭 Redis 连接
	if s.redisClient != nil {
		s.redisClient.Close()
		log.Println("Redis 连接已关闭")
	}

	// 关闭 due 服务器
	if s.server != nil {
		s.server.Stop()
		log.Println("due RPC 服务器已停止")
	}

	log.Println("服务已停止")
	return nil
}

func main() {
	// 解析命令行参数
	configFile := flag.String("config", "", "配置文件路径")
	serverPort := flag.Int("port", 8080, "HTTP 服务器端口")
	consulAddr := flag.String("consul", "127.0.0.1:8500", "Consul 地址")
	redisAddr := flag.String("redis", "127.0.0.1:6379", "Redis 地址")
	serviceAddr := flag.String("addr", "", "服务地址（IP:Port）")
	maxConn := flag.Int("max-conn", 10000, "最大并发连接数")
	flag.Parse()

	// 创建配置
	config := NewConfig()

	// 如果提供了配置文件，从文件加载
	if *configFile != "" {
		if err := config.LoadFromFile(*configFile); err != nil {
			log.Printf("加载配置文件失败：%v，使用默认配置", err)
		}
	}

	// 命令行参数覆盖配置文件
	if *serverPort != 8080 {
		config.ServerPort = *serverPort
	}
	if *consulAddr != "127.0.0.1:8500" {
		config.ConsulAddr = *consulAddr
	}
	if *redisAddr != "127.0.0.1:6379" {
		config.RedisAddr = *redisAddr
	}
	if *serviceAddr != "" {
		config.ServiceAddr = *serviceAddr
	}
	if *maxConn != 10000 {
		config.MaxConn = *maxConn
	}

	// 如果未指定服务地址，使用默认值
	if config.ServiceAddr == "" {
		config.ServiceAddr = fmt.Sprintf("127.0.0.1:%d", config.ServerPort)
	}

	// 创建 WebSocket 网关服务
	server, err := NewWebSocketGateServer(config)
	if err != nil {
		log.Fatalf("创建 WebSocket 网关服务失败：%v", err)
	}

	// 初始化 Redis
	if err := server.InitRedis(); err != nil {
		log.Fatalf("初始化 Redis 失败：%v", err)
	}

	// 初始化 Consul
	if err := server.InitConsul(); err != nil {
		log.Fatalf("初始化 Consul 失败：%v", err)
	}

	// 初始化 due Server
	if err := server.InitDueServer(); err != nil {
		log.Fatalf("初始化 due Server 失败：%v", err)
	}

	// 注册服务到 Consul
	if err := server.RegisterService(); err != nil {
		log.Fatalf("服务注册失败：%v", err)
	}

	// 启动服务
	if err := server.Start(); err != nil {
		log.Fatalf("启动服务失败：%v", err)
	}

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("收到退出信号...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Stop(); err != nil {
		log.Printf("停止服务时出错：%v", err)
	}

	log.Println("服务已退出")
}
