# due 快速开始指南 (v2.5.2)

本指南帮助你快速开始使用 due v2.5.2 游戏服务器框架。

## 版本说明

**当前目标版本**: due v2.5.2

**v2.5.2 更新内容**:
- 调整 http 组件路由注册方法
- 修复 node 组件 Push 方法无序的 BUG
- 修复 node 组件 WaitGroup 计数异常 BUG
- 重构集群内部 RPC 通信传输模块
- 改进集群调试模式引发的 RPC 通信中断问题
- 小幅提升性能

## 安装 due v2.5.2

在你的 Go 项目中添加 due 依赖：

```bash
# 主框架
go get -u github.com/dobyte/due/v2@latest

# 定位器（推荐使用 Redis）
go get -u github.com/dobyte/due/locate/redis/v2@latest

# WebSocket 网络组件
go get -u github.com/dobyte/due/network/ws/v2@latest

# Consul 服务注册
go get -u github.com/dobyte/due/registry/consul/v2@latest

# RPCX 传输组件
go get -u github.com/dobyte/due/transport/rpcx/v2@latest
```

## 运行示例

due 框架提供了完整的示例，可以通过 docker-compose 快速启动：

```bash
# 克隆 due 仓库
git clone https://github.com/dobyte/due.git
cd due

# 启动 docker-compose
docker-compose up -d
```

## 创建第一个 Gate 服务 (v2.5.2)

Gate 服务负责管理客户端连接和消息路由：

```go
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

    // 创建 WebSocket 服务器
    server := ws.NewServer()

    // 创建定位器
    locator := redis.NewLocator()

    // 创建服务注册
    registry := consul.NewRegistry()

    // 创建 Gate 组件
    component := gate.NewGate(
        gate.WithServer(server),
        gate.WithLocator(locator),
        gate.WithRegistry(registry),
    )

    // 添加组件到容器
    container.Add(component)

    // 启动服务
    container.Serve()
}
```

## 创建第一个 Node 服务 (v2.5.2)

Node 服务使用 Actor 模型处理游戏逻辑：

```go
package main

import (
    "fmt"
    "github.com/dobyte/due/locate/redis/v2"
    "github.com/dobyte/due/registry/consul/v2"
    "github.com/dobyte/due/v2"
    "github.com/dobyte/due/v2/cluster/node"
    "github.com/dobyte/due/v2/codes"
    "github.com/dobyte/due/v2/log"
    "github.com/dobyte/due/v2/utils/xtime"
)

const greetRoute = 1

func main() {
    // 创建容器
    container := due.NewContainer()

    // 创建定位器
    locator := redis.NewLocator()

    // 创建服务注册
    registry := consul.NewRegistry()

    // 创建 Node 组件
    component := node.NewNode(
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

func initListen(proxy *node.Proxy) {
    // 添加路由处理器：routeID, isSync, handlerFunc
    proxy.Router().AddRouteHandler(greetRoute, false, greetHandler)
}

type greetReq struct {
    Message string `json:"message"`
}

type greetRes struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func greetHandler(ctx node.Context) {
    req := &greetReq{}
    res := &greetRes{}

    defer func() {
        if err := ctx.Response(res); err != nil {
            log.Errorf("response message failed: %v", err)
        }
    }()

    if err := ctx.Parse(req); err != nil {
        log.Errorf("parse request message failed: %v", err)
        res.Code = codes.InternalError.Code()
        return
    }

    log.Info(req.Message)
    res.Code = codes.OK.Code()
    res.Message = fmt.Sprintf("I'm server, and the current time is: %s",
        xtime.Now().Format(xtime.DateTime))
}
```

## 完整项目示例

### Gate + Node 完整示例

```go
// cmd/gate/main.go - Gate 服务入口
package main

import (
    "github.com/dobyte/due/locate/redis/v2"
    "github.com/dobyte/due/network/ws/v2"
    "github.com/dobyte/due/registry/consul/v2"
    "github.com/dobyte/due/v2"
    "github.com/dobyte/due/v2/cluster/gate"
)

func main() {
    container := due.NewContainer()

    server := ws.NewServer(
        ws.WithPort(8800),  // WebSocket 端口
    )

    locator := redis.NewLocator(
        redis.WithAddr("127.0.0.1:6379"),
    )

    registry := consul.NewRegistry(
        consul.WithAddr("127.0.0.1:8500"),
    )

    component := gate.NewGate(
        gate.WithID("gate-001"),
        gate.WithName("gate"),
        gate.WithServer(server),
        gate.WithLocator(locator),
        gate.WithRegistry(registry),
    )

    container.Add(component)
    container.Serve()
}
```

```go
// cmd/node/main.go - Node 服务入口
package main

import (
    "github.com/dobyte/due/locate/redis/v2"
    "github.com/dobyte/due/registry/consul/v2"
    "github.com/dobyte/due/v2"
    "github.com/dobyte/due/v2/cluster/node"
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
        node.WithID("node-001"),
        node.WithName("node"),
        node.WithLocator(locator),
        node.WithRegistry(registry),
    )

    // 注册路由
    initListen(component.Proxy())

    container.Add(component)
    container.Serve()
}

func initListen(proxy *node.Proxy) {
    proxy.Router().AddRouteHandler(routeID, isSync, handlerFunc)
}
```

## 项目结构

一个典型的 due v2.5.2 项目结构：

```
my-game/
├── cmd/
│   ├── gate/           # 网关服务入口
│   │   └── main.go
│   └── node/           # 节点服务入口
│       └── main.go
├── internal/
│   ├── actor/          # Actor 实现
│   │   └── player.go
│   ├── handler/        # 消息处理器
│   │   └── login.go
│   └── logic/          # 业务逻辑
│       └── user.go
├── config/             # 配置文件
│   └── config.yaml
├── go.mod
└── go.sum
```

## 核心概念

### Container（容器）
due v2 使用 Container 统一管理组件生命周期：
- `due.NewContainer()` 创建容器
- `container.Add(component)` 添加组件
- `container.Serve()` 启动所有组件

### Gate（网关）
- 管理客户端连接（TCP/KCP/WebSocket）
- 消息路由和转发
- 会话管理

### Node（节点）
- 使用 Actor 模型处理有状态游戏逻辑
- 通过 Router 注册消息处理器
- 使用 Context 进行请求响应

### Locator（定位器）
- 用于服务发现和消息路由
- 支持 Redis、Etcd 等后端

### Registry（注册中心）
- 服务注册与发现
- 支持 Consul、Etcd、Nacos 等

## 下一步

- 阅读 [架构设计文档](../references/architecture-patterns.md) 了解 Gate/Node/Mesh 架构
- 阅读 [网关开发模式](../references/gate-patterns.md) 学习如何实现网关
- 阅读 [Node 开发模式](../references/node-patterns.md) 学习 Actor 模型
- 阅读 [组件使用指南](../references/component-patterns.md) 学习各组件使用
