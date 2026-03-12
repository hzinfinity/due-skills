package main

import (
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/registry/consul/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/codes"
	"github.com/dobyte/due/v2/log"
	"github.com/dobyte/due/v2/message"
)

const (
	LoginRoute  = 1 // 玩家登录
	ChatRoute   = 2 // 聊天消息
	LogoutRoute = 3 // 玩家登出
)

func main() {
	// 创建容器
	container := due.NewContainer()

	// 创建 Redis 定位器，用于玩家位置定位
	locator := redis.NewLocator(
		redis.WithAddr("127.0.0.1:6379"),
	)

	// 创建 Consul 注册中心，用于服务发现
	registry := consul.NewRegistry(
		consul.WithAddr("127.0.0.1:8500"),
		consul.WithID("node-001"),
		consul.WithName("node"),
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

func initListen(proxy *node.Proxy) {
	// 注册登录路由
	proxy.Router().AddRouteHandler(LoginRoute, false, loginHandler)
	// 注册聊天路由
	proxy.Router().AddRouteHandler(ChatRoute, false, chatHandler)
	// 注册登出路由
	proxy.Router().AddRouteHandler(LogoutRoute, false, logoutHandler)
}

// 登录请求/响应
type LoginRequest struct {
	UID  int64  `json:"uid"`
	Name string `json:"name"`
}

type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	UID     int64  `json:"uid"`
	Name    string `json:"name"`
}

func loginHandler(ctx node.Context) {
	req := &LoginRequest{}
	res := &LoginResponse{}

	defer func() {
		if err := ctx.Response(res); err != nil {
			log.Errorf("send response failed: %v", err)
		}
	}()

	if err := ctx.Parse(req); err != nil {
		log.Errorf("parse request failed: %v", err)
		res.Code = codes.InternalError.Code()
		res.Message = "parse request failed"
		return
	}

	if req.UID <= 0 {
		res.Code = codes.InvalidArgument.Code()
		res.Message = "invalid uid"
		return
	}

	if req.Name == "" {
		res.Code = codes.InvalidArgument.Code()
		res.Message = "invalid name"
		return
	}

	// 登录成功，返回玩家信息
	res.Code = codes.OK.Code()
	res.Message = "login success"
	res.UID = req.UID
	res.Name = req.Name

	log.Infof("玩家登录：uid=%d, name=%s", req.UID, req.Name)

	// 广播玩家登录消息给其他玩家
	broadcastLogin(req.UID, req.Name)
}

// 聊天请求/响应
type ChatRequest struct {
	UID     int64  `json:"uid"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func chatHandler(ctx node.Context) {
	req := &ChatRequest{}
	res := &ChatResponse{}

	defer func() {
		if err := ctx.Response(res); err != nil {
			log.Errorf("send response failed: %v", err)
		}
	}()

	if err := ctx.Parse(req); err != nil {
		log.Errorf("parse request failed: %v", err)
		res.Code = codes.InternalError.Code()
		res.Message = "parse request failed"
		return
	}

	if req.UID <= 0 {
		res.Code = codes.InvalidArgument.Code()
		res.Message = "invalid uid"
		return
	}

	if req.Content == "" {
		res.Code = codes.InvalidArgument.Code()
		res.Message = "empty content"
		return
	}

	// 获取玩家名称（实际项目中可从数据库或缓存获取）
	name := getPlayerName(req.UID)

	// 广播聊天消息给所有在线玩家
	broadcastChat(req.UID, name, req.Content)

	res.Code = codes.OK.Code()
	res.Message = "send success"

	log.Infof("玩家聊天：uid=%d, name=%s, content=%s", req.UID, name, req.Content)
}

// 登出请求/响应
type LogoutRequest struct {
	UID int64 `json:"uid"`
}

type LogoutResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func logoutHandler(ctx node.Context) {
	req := &LogoutRequest{}
	res := &LogoutResponse{}

	defer func() {
		if err := ctx.Response(res); err != nil {
			log.Errorf("send response failed: %v", err)
		}
	}()

	if err := ctx.Parse(req); err != nil {
		log.Errorf("parse request failed: %v", err)
		res.Code = codes.InternalError.Code()
		res.Message = "parse request failed"
		return
	}

	if req.UID <= 0 {
		res.Code = codes.InvalidArgument.Code()
		res.Message = "invalid uid"
		return
	}

	// 获取玩家名称
	name := getPlayerName(req.UID)

	// 广播玩家登出消息
	broadcastLogout(req.UID, name)

	res.Code = codes.OK.Code()
	res.Message = "logout success"

	log.Infof("玩家登出：uid=%d, name=%s", req.UID, name)
}

// broadcastLogin 广播玩家登录消息
func broadcastLogin(uid int64, name string) {
	// 广播给所有在线玩家（实际项目中可获取在线玩家列表）
	message.Broadcast([]int64{}, ChatRoute, map[string]interface{}{
		"type":    "login",
		"uid":     uid,
		"name":    name,
		"content": name + " 已加入聊天室",
	})
}

// broadcastChat 广播聊天消息
func broadcastChat(uid int64, name string, content string) {
	// 广播给所有在线玩家
	message.Broadcast([]int64{}, ChatRoute, map[string]interface{}{
		"type":    "chat",
		"uid":     uid,
		"name":    name,
		"content": content,
	})
}

// broadcastLogout 广播玩家登出消息
func broadcastLogout(uid int64, name string) {
	// 广播给所有在线玩家
	message.Broadcast([]int64{}, ChatRoute, map[string]interface{}{
		"type":    "logout",
		"uid":     uid,
		"name":    name,
		"content": name + " 已离开聊天室",
	})
}

// getPlayerName 获取玩家名称（简单实现，实际项目中应从数据库或缓存获取）
func getPlayerName(uid int64) string {
	return "Player_" + string(rune(uid))
}
