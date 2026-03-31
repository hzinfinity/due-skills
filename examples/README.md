# due 示例代码 (v2.5.2)

本目录包含 due v2.5.2 框架的完整可运行示例代码。

## 目录结构

```
examples/
├── gate-ws/           # WebSocket 网关
│   ├── main.go
│   └── go.mod
├── gate-tcp/          # TCP 网关
│   ├── main.go
│   └── go.mod
├── gate-kcp/          # KCP 网关
│   ├── main.go
│   └── go.mod
├── node-basic/        # 基础 Node 服务
│   ├── main.go
│   ├── handler/
│   │   └── greet.go
│   └── go.mod
├── chat-room/         # 完整聊天室示例
│   ├── docker-compose.yml
│   ├── gate/
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   └── node/
│       ├── main.go
│       ├── handler/
│       │   ├── login.go
│       │   ├── chat.go
│       │   └── logout.go
│       ├── go.mod
│       └── Dockerfile
└── mesh/              # Mesh 微服务
    ├── main.go
    ├── service/
    │   └── user.go
    └── go.mod
```

## 快速开始

### 1. gate-ws (WebSocket 网关)

适用于 H5 游戏、即时通讯等场景：

```bash
cd examples/gate-ws
go mod tidy
go run main.go
```

### 2. gate-tcp (TCP 网关)

适用于对实时性要求高的游戏，如动作游戏、射击游戏：

```bash
cd examples/gate-tcp
go mod tidy
go run main.go
```

### 3. gate-kcp (KCP 网关)

适用于弱网络环境下的实时游戏，如 MOBA、吃鸡：

```bash
cd examples/gate-kcp
go mod tidy
go run main.go
```

### 4. node-basic (Node 服务)

演示 Actor 路由处理器模式：

```bash
cd examples/node-basic
go mod tidy
go run main.go
```

### 5. chat-room (完整聊天室)

Gate + Node 组合，包含登录、聊天、登出功能：

```bash
cd examples/chat-room
docker-compose up -d

# 或不使用 Docker:
# 先启动 Redis 和 Consul
cd gate && go mod tidy && go run main.go &
cd node && go mod tidy && go run main.go
```

### 6. mesh (微服务)

演示 RPCX 传输和服务注册：

```bash
cd examples/mesh
go mod tidy
go run main.go
```

## 运行依赖

| 服务 | 端口 | 说明 |
|------|------|------|
| Redis | 6379 | 定位器 |
| Consul | 8500 | 服务注册 |

快速启动依赖服务：

```bash
docker run -d --name redis -p 6379:6379 redis:latest
docker run -d --name consul -p 8500:8500 consul:latest
```

## v2.5.2 关键模式

1. **Container 模式**: `due.NewContainer()` 统一管理组件
2. **组件模式**: Gate/Node/Mesh 都作为组件创建
3. **路由处理器**: Node 使用 `proxy.Router().AddRouteHandler()` 注册
4. **模块路径**: 使用 `/v2` 路径导入
