package main

import (
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/registry/consul/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"node/handler"
)

func main() {
	container := due.NewContainer()

	locator := redis.NewLocator(
		redis.WithAddr("127.0.0.1:6379"),
	)

	registry := consul.NewRegistry(
		consul.WithAddr("127.0.0.1:8500"),
	)

	component := node.NewNode(
		node.WithID("node-chat-001"),
		node.WithName("node"),
		node.WithLocator(locator),
		node.WithRegistry(registry),
	)

	proxy := component.Proxy()
	proxy.Router().AddRouteHandler(handler.RouteLogin, false, handler.LoginHandler)
	proxy.Router().AddRouteHandler(handler.RouteChat, false, handler.ChatHandler)
	proxy.Router().AddRouteHandler(handler.RouteLogout, false, handler.LogoutHandler)

	container.Add(component)

	log.Info("Chat Room Node 启动中")
	container.Serve()
}
