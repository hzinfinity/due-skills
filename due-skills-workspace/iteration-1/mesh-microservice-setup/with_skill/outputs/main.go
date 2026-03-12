package main

import (
	"context"

	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/registry/consul/v2"
	"github.com/dobyte/due/transport/rpcx/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/mesh"
)

func main() {
	// 创建容器
	container := due.NewContainer()

	// 创建定位器 - 用于服务发现
	locator := redis.NewLocator(
		redis.WithAddr("127.0.0.1:6379"),
	)

	// 创建服务注册中心 - 使用 Consul
	registry := consul.NewRegistry(
		consul.WithAddr("127.0.0.1:8500"),
	)

	// 创建传输器 - 使用 RPCX 作为传输协议
	transporter := rpcx.NewTransporter()

	// 创建 Mesh 组件
	component := mesh.NewMesh(
		mesh.WithID("mesh-user-001"),
		mesh.WithName("mesh-user"),
		mesh.WithLocator(locator),
		mesh.WithRegistry(registry),
		mesh.WithTransporter(transporter),
	)

	// 注册服务提供者
	userService := &UserService{}
	component.AddServiceProvider("user.service", "UserService", userService)

	// 添加组件到容器
	container.Add(component)

	// 启动服务
	container.Serve()
}

// UserService 用户服务结构体
type UserService struct{}

// GetUserRequest 获取用户信息请求
type GetUserRequest struct {
	UID int64 `json:"uid"`
}

// GetUserResponse 获取用户信息响应
type GetUserResponse struct {
	Code   int    `json:"code"`
	UID    int64  `json:"uid"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}

// GetUser 获取用户信息接口
// 该方法通过 RPCX 协议对外提供服务
func (s *UserService) GetUser(ctx context.Context, req *GetUserRequest, res *GetUserResponse) error {
	// 查询用户信息逻辑
	// 这里可以根据实际需求对接数据库或其他数据源
	res.Code = 0
	res.UID = req.UID
	res.Name = "User_" + string(rune(req.UID))
	res.Email = "user@example.com"
	res.Avatar = "https://example.com/avatar/default.png"
	return nil
}
