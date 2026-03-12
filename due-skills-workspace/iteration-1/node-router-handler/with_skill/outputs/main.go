package main

import (
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/registry/consul/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/codes"
	"github.com/dobyte/due/v2/log"
)

const RouteLogin = 1

func main() {
	// 创建容器
	container := due.NewContainer()

	// 创建定位器
	locator := redis.NewLocator(
		redis.WithAddr("127.0.0.1:6379"),
	)

	// 创建服务注册中心
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

	// 添加组件到容器
	container.Add(component)

	// 启动服务
	container.Serve()
}

// initListen 初始化监听路由
func initListen(proxy *node.Proxy) {
	// 注册登录路由处理器：routeID=1, isSync=false, handlerFunc=loginHandler
	proxy.Router().AddRouteHandler(RouteLogin, false, loginHandler)
}

// loginReq 登录请求
type loginReq struct {
	UID   int64  `json:"uid"`
	Token string `json:"token"`
}

// loginRes 登录响应
type loginRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// loginHandler 玩家登录处理器（Actor 模型）
func loginHandler(ctx node.Context) {
	req := &loginReq{}
	res := &loginRes{}

	// 确保响应一定会发送
	defer func() {
		if err := ctx.Response(res); err != nil {
			log.Errorf("response message failed: %v", err)
		}
	}()

	// 解析请求
	if err := ctx.Parse(req); err != nil {
		log.Errorf("parse request message failed: %v", err)
		res.Code = codes.InternalError.Code()
		res.Message = "解析请求失败"
		return
	}

	// 验证参数
	if req.UID <= 0 {
		res.Code = codes.InvalidArgument.Code()
		res.Message = "UID 不能为空"
		return
	}

	if req.Token == "" {
		res.Code = codes.InvalidArgument.Code()
		res.Message = "Token 不能为空"
		return
	}

	// TODO: 这里可以添加实际的 token 验证逻辑
	// 例如：验证 token 是否有效，检查用户是否被封禁等

	// 登录成功
	res.Code = codes.OK.Code()
	res.Message = "登录成功"

	log.Infof("玩家登录成功：uid=%d, token=%s", req.UID, req.Token)
}
