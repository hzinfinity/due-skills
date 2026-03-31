package main

import (
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/network/kcp/v2"
	"github.com/dobyte/due/registry/consul/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/gate"
	"github.com/dobyte/due/v2/log"
)

func main() {
	// 创建容器
	container := due.NewContainer()

	// 创建 KCP 服务器 - 使用正常模式
	// 模式说明:
	//   0: 快速模式 (RTT ~20ms) - 适合竞技游戏、音游
	//   1: 正常模式 (RTT ~40ms) - 适合 MOBA、FPS
	//   2: 流畅模式 (RTT ~60ms) - 适合 MMO、休闲游戏
	server := kcp.NewServer(
		kcp.WithPort(10000),
		kcp.WithMode(1), // 正常模式
		kcp.WithMaxConnNum(5000),
		kcp.WithMsgSize(8192),
		kcp.WithSendChanSize(2048),
	)

	// 创建定位器
	locator := redis.NewLocator(
		redis.WithAddr("127.0.0.1:6379"),
	)

	// 创建注册中心
	registry := consul.NewRegistry(
		consul.WithAddr("127.0.0.1:8500"),
	)

	// 创建 Gate 组件
	component := gate.NewGate(
		gate.WithID("gate-kcp-001"),
		gate.WithName("gate"),
		gate.WithServer(server),
		gate.WithLocator(locator),
		gate.WithRegistry(registry),
	)

	// 添加组件到容器
	container.Add(component)

	log.Info("Gate KCP 服务启动中，端口: 10000")

	// 启动服务
	container.Serve()
}
