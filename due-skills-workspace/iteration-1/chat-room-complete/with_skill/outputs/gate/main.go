package main

import (
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/network/ws/v2"
	"github.com/dobyte/due/registry/consul/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/gate"
)

func main() {
	// 创建容器
	container := due.NewContainer()

	// 创建 WebSocket 服务器，监听端口 8800
	server := ws.NewServer(
		ws.WithPort(8800),
		ws.WithMaxConnNum(10000),
		ws.WithMsgSize(4096),
		ws.WithHeartbeatInterval(30),
	)

	// 创建 Redis 定位器，用于玩家位置定位
	locator := redis.NewLocator(
		redis.WithAddr("127.0.0.1:6379"),
	)

	// 创建 Consul 注册中心，用于服务发现
	registry := consul.NewRegistry(
		consul.WithAddr("127.0.0.1:8500"),
		consul.WithID("gate-001"),
		consul.WithName("gate"),
	)

	// 创建 Gate 组件
	component := gate.NewGate(
		gate.WithID("gate-001"),
		gate.WithName("gate"),
		gate.WithServer(server),
		gate.WithLocator(locator),
		gate.WithRegistry(registry),
	)

	// 添加组件到容器并启动服务
	container.Add(component)
	container.Serve()
}
