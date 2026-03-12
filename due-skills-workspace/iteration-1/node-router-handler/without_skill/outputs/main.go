package main

import (
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/registry/consul/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/codes"
	"github.com/dobyte/due/v2/log"
)

const (
	// LoginRoute 玩家登录路由 ID
	LoginRoute = 1
)

func main() {
	// 创建容器
	container := due.NewContainer()

	// 创建 Redis 定位器，用于服务发现
	locator := redis.NewLocator(
		redis.WithAddr("127.0.0.1:6379"),
	)

	// 创建 Consul 注册中心，用于服务注册
	registry := consul.NewRegistry(
		consul.WithAddr("127.0.0.1:8500"),
	)

	// 创建 Node 组件
	component := node.NewNode(
		node.WithID("node-001"),
		node.WithName("node"),
		node.WithLocator(locator),
		node.WithRegistry(registry),
	)

	// 注册路由处理器
	initListen(component.Proxy())

	// 添加组件到容器并启动服务
	container.Add(component)
	container.Serve()
}

// initListen 注册所有路由处理器
func initListen(proxy *node.Proxy) {
	// 注册玩家登录路由处理器
	// 参数：routeID=1, isSync=false(异步处理), handlerFunc
	proxy.Router().AddRouteHandler(LoginRoute, false, loginHandler)
}

// LoginRequest 玩家登录请求
type LoginRequest struct {
	UID   int64  `json:"uid"`    // 用户 ID
	Token string `json:"token"`  // 认证令牌
}

// LoginResponse 玩家登录响应
type LoginResponse struct {
	Code    int    `json:"code"`    // 响应码
	Message string `json:"message"` // 响应消息
}

// loginHandler 玩家登录路由处理器
// 使用 Actor 模型处理消息，每个玩家的消息顺序处理
func loginHandler(ctx node.Context) {
	req := &LoginRequest{}
	res := &LoginResponse{}

	// 确保响应发送
	defer func() {
		if err := ctx.Response(res); err != nil {
			log.Errorf("send response failed: %v", err)
		}
	}()

	// 解析请求数据
	if err := ctx.Parse(req); err != nil {
		log.Errorf("parse request failed: %v", err)
		res.Code = codes.InternalError.Code()
		res.Message = "Internal error"
		return
	}

	// 验证 Token
	if !validateToken(req.Token) {
		log.Errorf("invalid token: uid=%d", req.UID)
		res.Code = codes.Unauthorized.Code()
		res.Message = "Invalid token"
		return
	}

	// 加载玩家数据（示例：从数据库或缓存加载）
	playerData, err := loadPlayerData(req.UID)
	if err != nil {
		log.Errorf("load player data failed: uid=%d, err=%v", req.UID, err)
		res.Code = codes.InternalError.Code()
		res.Message = "Failed to load player data"
		return
	}

	// 更新玩家在线状态
	if err := setPlayerOnline(req.UID, true); err != nil {
		log.Errorf("set player online failed: uid=%d, err=%v", req.UID, err)
	}

	// 构建响应
	res.Code = codes.OK.Code()
	res.Message = "Login successful"

	log.Infof("player login: uid=%d, name=%s", req.UID, playerData.Name)
}

// validateToken 验证 Token 有效性
func validateToken(token string) bool {
	// TODO: 实现实际的 Token 验证逻辑
	// 可以调用认证服务或验证 JWT
	if token == "" {
		return false
	}
	return true
}

// PlayerData 玩家数据结构
type PlayerData struct {
	UID   int64  `json:"uid"`
	Name  string `json:"name"`
	Level int    `json:"level"`
}

// loadPlayerData 加载玩家数据
func loadPlayerData(uid int64) (*PlayerData, error) {
	// TODO: 从数据库或缓存加载实际玩家数据
	// 示例：返回模拟数据
	return &PlayerData{
		UID:   uid,
		Name:  "Player",
		Level: 1,
	}, nil
}

// setPlayerOnline 设置玩家在线状态
func setPlayerOnline(uid int64, online bool) error {
	// TODO: 更新玩家在线状态到缓存或数据库
	log.Debugf("set player online: uid=%d, online=%v", uid, online)
	return nil
}
