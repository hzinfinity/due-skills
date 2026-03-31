package main

import (
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/registry/consul/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"node-basic/handler"
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
		node.WithID("node-basic-001"),
		node.WithName("node"),
		node.WithLocator(locator),
		node.WithRegistry(registry),
	)

	initListen(component.Proxy())

	container.Add(component)

	log.Info("Node 服务启动中")
	container.Serve()
}

func initListen(proxy *node.Proxy) {
	proxy.Router().AddRouteHandler(handler.GreetRoute, false, handler.GreetHandler)
}
