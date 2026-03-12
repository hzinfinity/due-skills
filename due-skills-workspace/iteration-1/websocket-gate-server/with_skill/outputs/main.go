package main

import (
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/network/ws/v2"
	"github.com/dobyte/due/registry/consul/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/gate"
)

func main() {
	// 创建容器，用于管理组件生命周期
	container := due.NewContainer()

	// 创建 WebSocket 服务器，支持 10000 个并发连接
	server := ws.NewServer(
		ws.WithPort(8800),           // WebSocket 监听端口
		ws.WithMaxConnNum(10000),    // 最大连接数 10000
		ws.WithPendingNum(10000),    // 待处理消息队列大小
		ws.WithReadBufferSize(2048), // 读缓冲区大小 (KB)
		ws.WithWriteBufferSize(2048),// 写缓冲区大小 (KB)
	)

	// 创建 Redis 定位器，用于服务发现和消息路由
	locator := redis.NewLocator(
		redis.WithAddr("127.0.0.1:6379"), // Redis 地址
		redis.WithDB(0),                   // Redis 数据库
		redis.WithPassword(""),            // Redis 密码（如有）
	)

	// 创建 Consul 注册中心，用于服务注册与发现
	registry := consul.NewRegistry(
		consul.WithAddr("127.0.0.1:8500"), // Consul 地址
		consul.WithScheme("http"),         // Consul 协议
	)

	// 创建 Gate 组件
	component := gate.NewGate(
		gate.WithID("gate-ws-001"),       // Gate 服务唯一 ID
		gate.WithName("gate"),            // Gate 服务名称
		gate.WithServer(server),          // 绑定 WebSocket 服务器
		gate.WithLocator(locator),        // 绑定定位器
		gate.WithRegistry(registry),      // 绑定注册中心
		gate.WithPort(8801),              // Gate RPC 端口
	)

	// 添加组件到容器
	container.Add(component)

	// 启动服务
	container.Serve()
}
