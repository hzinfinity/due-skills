package main

import (
	"context"
	"fmt"
	"log"

	"github.com/duereg/duo-go-sdk/pkg"
	"github.com/duereg/duo-go-sdk/pkg/config"
	"github.com/duereg/duo-go-sdk/pkg/engine"
	"github.com/duereg/duo-go-sdk/pkg/logic"
	"github.com/duereg/duo-go-sdk/pkg/transport/rpcx"
)

// User 用户数据结构
type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// GetUserReq 获取用户请求
type GetUserReq struct {
	UID int64 `json:"uid"`
}

// GetUserResp 获取用户响应
type GetUserResp struct {
	User *User `json:"user"`
}

// UserService 用户服务
type UserService struct {
	logic.Base
}

// GetUser 获取用户信息接口
func (s *UserService) GetUser(ctx context.Context, req *GetUserReq) (*GetUserResp, error) {
	// 模拟从数据库查询用户
	user := &User{
		ID:   req.UID,
		Name: fmt.Sprintf("User_%d", req.UID),
		Age:  25,
	}

	resp := &GetUserResp{
		User: user,
	}

	return resp, nil
}

func main() {
	// 初始化 due 引擎
	cfg := &config.Config{
		Name:    "user-service",
		Version: "v2.5.2",
	}

	// 创建引擎实例
	eng, err := engine.NewEngine(cfg)
	if err != nil {
		log.Fatalf("failed to create engine: %v", err)
	}

	// 创建 RPCX 传输配置
	rpcxCfg := &rpcx.Config{
		Network:       "tcp",
		Addr:          ":8972",
		ServicePath:   "UserService",
		ServiceMethod: "GetUser",
	}

	// 创建 RPCX 传输实例
	rpcxTransport := rpcx.NewRPCX(rpcxCfg)

	// 注册服务
	userService := &UserService{}
	eng.RegisterService(userService, pkg.WithTransport(rpcxTransport))

	// 启动引擎
	if err := eng.Start(); err != nil {
		log.Fatalf("failed to start engine: %v", err)
	}

	log.Println("user-service started on :8972")

	// 等待关闭
	eng.Wait()
}
