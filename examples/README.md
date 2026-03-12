# due 示例代码 (v2.5.2)

本目录包含 due v2.5.2 框架的完整示例代码。

## 示例项目结构

```
examples/
├── gate-ws/           # WebSocket 网关示例 (v2.5.2)
├── gate-tcp/          # TCP 网关示例 (v2.5.2)
├── node-basic/        # 基础 Node 示例 (v2.5.2)
├── chat-room/         # 聊天室完整示例 (v2.5.2)
└── mesh/              # Mesh 微服务示例 (v2.5.2)
```

## Gate 服务示例 (v2.5.2)

```go
// examples/gate-ws/main.go
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

    // WebSocket 服务器
    server := ws.NewServer(
        ws.WithPort(8800),
        ws.WithMaxConnNum(10000),
    )

    // 定位器
    locator := redis.NewLocator(
        redis.WithAddr("127.0.0.1:6379"),
    )

    // 注册中心
    registry := consul.NewRegistry(
        consul.WithAddr("127.0.0.1:8500"),
    )

    // Gate 组件
    component := gate.NewGate(
        gate.WithID("gate-ws-001"),
        gate.WithName("gate"),
        gate.WithServer(server),
        gate.WithLocator(locator),
        gate.WithRegistry(registry),
    )

    // 添加并启动
    container.Add(component)
    container.Serve()
}
```

## Node 服务示例 (v2.5.2)

```go
// examples/node-basic/main.go
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

    // 定位器
    locator := redis.NewLocator(
        redis.WithAddr("127.0.0.1:6379"),
    )

    // 注册中心
    registry := consul.NewRegistry(
        consul.WithAddr("127.0.0.1:8500"),
    )

    // Node 组件
    component := node.NewNode(
        node.WithID("node-001"),
        node.WithName("node"),
        node.WithLocator(locator),
        node.WithRegistry(registry),
    )

    // 注册路由处理器
    initListen(component.Proxy())

    container.Add(component)
    container.Serve()
}

func initListen(proxy *node.Proxy) {
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

## 完整聊天室示例 (v2.5.2)

```go
// examples/chat-room/main.go
package main

import (
    "github.com/dobyte/due/locate/redis/v2"
    "github.com/dobyte/due/network/ws/v2"
    "github.com/dobyte/due/registry/consul/v2"
    "github.com/dobyte/due/v2"
    "github.com/dobyte/due/v2/cluster/gate"
    "github.com/dobyte/due/v2/cluster/node"
    "github.com/dobyte/due/v2/codes"
    "github.com/dobyte/due/v2/log"
)

const (
    RouteLogin  = 1
    RouteChat   = 2
    RouteLogout = 3
)

func main() {
    container := due.NewContainer()

    // 定位器
    locator := redis.NewLocator(
        redis.WithAddr("127.0.0.1:6379"),
    )

    // 注册中心
    registry := consul.NewRegistry(
        consul.WithAddr("127.0.0.1:8500"),
    )

    // Gate 组件
    gateComponent := gate.NewGate(
        gate.WithID("gate-chat-001"),
        gate.WithName("gate"),
        gate.WithServer(ws.NewServer(ws.WithPort(8800))),
        gate.WithLocator(locator),
        gate.WithRegistry(registry),
    )

    // Node 组件
    nodeComponent := node.NewNode(
        node.WithID("node-chat-001"),
        node.WithName("node"),
        node.WithLocator(locator),
        node.WithRegistry(registry),
    )

    // 注册路由
    initListen(nodeComponent.Proxy())

    container.Add(gateComponent)
    container.Add(nodeComponent)
    container.Serve()
}

func initListen(proxy *node.Proxy) {
    proxy.Router().AddRouteHandler(RouteLogin, false, loginHandler)
    proxy.Router().AddRouteHandler(RouteChat, false, chatHandler)
    proxy.Router().AddRouteHandler(RouteLogout, false, logoutHandler)
}

// 登录
type LoginRequest struct {
    UID  int64  `json:"uid"`
    Name string `json:"name"`
}

type LoginResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    RoomID  string `json:"room_id"`
}

func loginHandler(ctx node.Context) {
    req := &LoginRequest{}
    res := &LoginResponse{}

    defer func() {
        ctx.Response(res)
    }()

    if err := ctx.Parse(req); err != nil {
        res.Code = codes.InternalError.Code()
        return
    }

    // 加入聊天室（使用 UID 作为 Actor ID）
    res.Code = codes.OK.Code()
    res.Message = "登录成功"
    res.RoomID = "room_001"

    log.Infof("玩家登录：uid=%d, name=%s", req.UID, req.Name)
}

// 聊天
type ChatRequest struct {
    Content string `json:"content"`
}

type ChatResponse struct {
    Code int `json:"code"`
}

func chatHandler(ctx node.Context) {
    req := &ChatRequest{}
    res := &ChatResponse{}

    defer func() {
        ctx.Response(res)
    }()

    if err := ctx.Parse(req); err != nil {
        res.Code = codes.InternalError.Code()
        return
    }

    // 广播聊天消息
    log.Infof("玩家聊天：uid=%d, content=%s", ctx.Uid(), req.Content)

    res.Code = codes.OK.Code()
}

// 登出
type LogoutRequest struct{}
type LogoutResponse struct {
    Code int `json:"code"`
}

func logoutHandler(ctx node.Context) {
    res := &LogoutResponse{}
    defer func() {
        ctx.Response(res)
    }()

    log.Infof("玩家登出：uid=%d", ctx.Uid())
    res.Code = codes.OK.Code()
}
```

## Mesh 微服务示例 (v2.5.2)

```go
// examples/mesh/main.go
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
    container := due.NewContainer()

    // 定位器
    locator := redis.NewLocator(
        redis.WithAddr("127.0.0.1:6379"),
    )

    // 注册中心
    registry := consul.NewRegistry(
        consul.WithAddr("127.0.0.1:8500"),
    )

    // 传输器
    transporter := rpcx.NewTransporter()

    // Mesh 组件
    component := mesh.NewMesh(
        mesh.WithID("mesh-001"),
        mesh.WithName("mesh"),
        mesh.WithLocator(locator),
        mesh.WithRegistry(registry),
        mesh.WithTransporter(transporter),
    )

    // 注册服务提供者
    component.AddServiceProvider("user.service", "UserService", &UserService{})

    container.Add(component)
    container.Serve()
}

// 用户服务
type UserService struct{}

type GetUserRequest struct {
    UID int64 `json:"uid"`
}

type GetUserResponse struct {
    Code int    `json:"code"`
    UID  int64  `json:"uid"`
    Name string `json:"name"`
}

func (s *UserService) GetUser(ctx context.Context, req *GetUserRequest, res *GetUserResponse) error {
    res.Code = 0
    res.UID = req.UID
    res.Name = "Alice"
    return nil
}
```

## Docker Compose 配置 (v2.5.2)

```yaml
# examples/docker-compose.yml
version: '3'
services:
  consul:
    image: consul:latest
    ports:
      - "8500:8500"
    command: agent -server -ui -bootstrap-expect=1 -client=0.0.0.0

  redis:
    image: redis:latest
    ports:
      - "6379:6379"

  gate:
    build:
      context: .
      dockerfile: Dockerfile.gate
    ports:
      - "8800:8800"
    environment:
      - CONSUL_ADDR=consul:8500
      - REDIS_ADDR=redis:6379
    depends_on:
      - consul
      - redis

  node:
    build:
      context: .
      dockerfile: Dockerfile.node
    environment:
      - CONSUL_ADDR=consul:8500
      - REDIS_ADDR=redis:6379
    depends_on:
      - consul
      - redis
```

## Dockerfile

```dockerfile
# Dockerfile.gate
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gate ./gate

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/gate .
COPY config/ ./config/

EXPOSE 8800
CMD ["./gate"]
```

```dockerfile
# Dockerfile.node
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o node ./node

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/node .
COPY config/ ./config/

CMD ["./node"]
```

## 运行示例

```bash
# 安装依赖
go get -u github.com/dobyte/due/v2@latest
go get -u github.com/dobyte/due/locate/redis/v2@latest
go get -u github.com/dobyte/due/network/ws/v2@latest
go get -u github.com/dobyte/due/registry/consul/v2@latest
go get -u github.com/dobyte/due/transport/rpcx/v2@latest

# 启动依赖服务
cd examples
docker-compose up -d consul redis

# 运行 Gate 服务
go run gate-ws/main.go

# 运行 Node 服务
go run node-basic/main.go

# 运行 Mesh 服务
go run mesh/main.go
```

## v2.5.2 注意事项

1. **使用 Container**: 所有组件通过 `due.NewContainer()` 管理
2. **组件模式**: Gate/Node/Mesh 都作为组件创建
3. **路由处理器**: Node 使用 `proxy.Router().AddRouteHandler()` 注册处理器
4. **模块路径**: 使用 `/v2` 路径导入模块
