package main

import (
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/registry/consul/v2"
	"github.com/dobyte/due/transport/rpcx/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/mesh"
	"github.com/dobyte/due/v2/log"

	"mesh/service"
)

func main() {
	container := due.NewContainer()

	locator := redis.NewLocator(
		redis.WithAddr("127.0.0.1:6379"),
	)

	registry := consul.NewRegistry(
		consul.WithAddr("127.0.0.1:8500"),
	)

	transporter := rpcx.NewTransporter()

	component := mesh.NewMesh(
		mesh.WithID("mesh-user-001"),
		mesh.WithName("mesh-user"),
		mesh.WithLocator(locator),
		mesh.WithRegistry(registry),
		mesh.WithTransporter(transporter),
	)

	userService := &service.UserService{}
	component.AddServiceProvider("user.service", "UserService", userService)

	container.Add(component)

	log.Info("Mesh User 服务启动中，RPCX 传输")
	container.Serve()
}
