# due Node 开发模式 (v2.5.2)

本文档详细介绍 due v2.5.2 框架中 Node 服务的开发模式，重点介绍 Actor 模型的使用。

## Node 概述

Node 服务是游戏服务器的核心，负责：
- 使用 Actor 模型处理有状态游戏逻辑
- 接收 Gate 转发的客户端消息
- 数据持久化
- 与其他服务通信

## v2.5.2 Node 完整示例

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

## Actor 模型基础

### 什么是 Actor

Actor 是并发编程的基本单元，具有以下特性：
- **独立状态**：每个 Actor 拥有独立的状态
- **消息驱动**：通过消息进行通信
- **顺序处理**：消息按顺序处理
- **位置透明**：可以在任何 Node 上运行

### Actor 生命周期

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│  Init   │ ──▶ │ Running │ ──▶ │Stopping │ ──▶ │ Destroy │
└─────────┘     └─────────┘     └─────────┘     └─────────┘
```

## 路由处理器 (v2.5.2)

在 due v2.5.2 中，Actor 逻辑通过路由处理器实现：

### 基础路由处理器

```go
const (
    LoginRoute = 1
    MoveRoute  = 2
    ChatRoute  = 3
)

func initListen(proxy *node.Proxy) {
    // 注册登录路由
    proxy.Router().AddRouteHandler(LoginRoute, false, loginHandler)
    // 注册移动路由
    proxy.Router().AddRouteHandler(MoveRoute, false, moveHandler)
    // 注册聊天路由
    proxy.Router().AddRouteHandler(ChatRoute, false, chatHandler)
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

    // 处理登录逻辑
    res.Code = codes.OK.Code()
    res.UID = req.UID
}

func moveHandler(ctx node.Context) {
    req := &MoveRequest{}
    res := &MoveResponse{}

    defer func() {
        ctx.Response(res)
    }()

    if err := ctx.Parse(req); err != nil {
        res.Code = codes.InternalError.Code()
        return
    }

    // 处理移动逻辑
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

    // 处理聊天逻辑
}
```

### 同步与异步处理

```go
// 异步处理（推荐）- isSync = false
proxy.Router().AddRouteHandler(routeID, false, handler)

// 同步处理 - isSync = true
// 同步模式下，消息按顺序处理，适用于需要严格顺序的场景
proxy.Router().AddRouteHandler(routeID, true, handler)
```

## 创建 Node 服务 (v2.5.2)

### 基础 Node

```go
package main

import (
    "github.com/dobyte/due/locate/redis/v2"
    "github.com/dobyte/due/registry/consul/v2"
    "github.com/dobyte/due/v2"
    "github.com/dobyte/due/v2/cluster/node"
)

func main() {
    container := due.NewContainer()

    locator := redis.NewLocator()
    registry := consul.NewRegistry()

    component := node.NewNode(
        node.WithLocator(locator),
        node.WithRegistry(registry),
    )

    initListen(component.Proxy())

    container.Add(component)
    container.Serve()
}
```

### 完整 Node 配置

```go
component := node.NewNode(
    node.WithID("node-001"),
    node.WithName("node"),
    node.WithLocator(locator),
    node.WithRegistry(registry),
    node.WithWorkerSize(32),         // Worker 数量
)
```

## 消息处理 (v2.5.2)

### 消息结构

```go
// 请求消息
type LoginRequest struct {
    UID      int64  `json:"uid"`
    Token    string `json:"token"`
}

// 响应消息
type LoginResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    UID     int64  `json:"uid"`
}
```

### Context 接口

```go
// node.Context 提供的方法
ctx.Route()      // 获取路由 ID
ctx.Session()    // 获取会话
ctx.Parse(req)   // 解析请求数据
ctx.Response(res)// 发送响应
ctx.Uid()        // 获取用户 ID
ctx.Cid()        // 获取连接 ID
```

## Actor 间通信 (v2.5.2)

### 发送消息

```go
import "github.com/dobyte/due/v2/message"

// 发送消息给特定用户
func sendMessage(uid int64, route int64, data interface{}) {
    message.Send(uid, route, data)
}

// 广播消息给多个用户
func broadcastMessage(uids []int64, route int64, data interface{}) {
    message.Broadcast(uids, route, data)
}
```

### 推送消息

```go
// 通过 Proxy 推送消息
func pushMessage(proxy *node.Proxy, uid int64, route int64, data interface{}) {
    proxy.Push(uid, route, data)
}
```

## 数据持久化

### 使用 Redis 缓存

```go
import (
    "github.com/dobyte/due/redis/v2"
)

type PlayerData struct {
    UID   int64  `json:"uid"`
    Name  string `json:"name"`
    Level int    `json:"level"`
}

func savePlayerData(client *redis.Client, data *PlayerData) error {
    key := fmt.Sprintf("player:%d", data.UID)
    return client.Set(key, data).Err()
}

func loadPlayerData(client *redis.Client, uid int64) (*PlayerData, error) {
    key := fmt.Sprintf("player:%d", uid)
    data := &PlayerData{}
    err := client.Get(key).Scan(data)
    return data, err
}
```

## 完整示例

### 玩家系统完整实现

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
)

const (
    LoginRoute  = 1
    LogoutRoute = 2
    MoveRoute   = 3
    ChatRoute   = 4
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

    initListen(component.Proxy())

    container.Add(component)
    container.Serve()
}

func initListen(proxy *node.Proxy) {
    proxy.Router().AddRouteHandler(LoginRoute, false, loginHandler)
    proxy.Router().AddRouteHandler(LogoutRoute, false, logoutHandler)
    proxy.Router().AddRouteHandler(MoveRoute, false, moveHandler)
    proxy.Router().AddRouteHandler(ChatRoute, false, chatHandler)
}

// 登录请求/响应
type LoginRequest struct {
    UID   int64  `json:"uid"`
    Token string `json:"token"`
}

type LoginResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    UID     int64  `json:"uid"`
    Name    string `json:"name"`
}

func loginHandler(ctx node.Context) {
    req := &LoginRequest{}
    res := &LoginResponse{}

    defer func() {
        ctx.Response(res)
    }()

    if err := ctx.Parse(req); err != nil {
        log.Errorf("parse request failed: %v", err)
        res.Code = codes.InternalError.Code()
        return
    }

    // 验证 Token
    if !validateToken(req.Token) {
        res.Code = codes.Unauthorized.Code()
        res.Message = "Invalid token"
        return
    }

    // 加载玩家数据
    name := loadPlayerName(req.UID)

    res.Code = codes.OK.Code()
    res.Message = "Success"
    res.UID = req.UID
    res.Name = name

    log.Infof("玩家登录：uid=%d", req.UID)
}

// 登出
type LogoutRequest struct {
    UID int64 `json:"uid"`
}

type LogoutResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func logoutHandler(ctx node.Context) {
    req := &LogoutRequest{}
    res := &LogoutResponse{}

    defer func() {
        ctx.Response(res)
    }()

    if err := ctx.Parse(req); err != nil {
        res.Code = codes.InternalError.Code()
        return
    }

    // 保存玩家数据
    savePlayerData(req.UID)

    res.Code = codes.OK.Code()
    res.Message = "Success"
}

// 移动
type MoveRequest struct {
    UID int64   `json:"uid"`
    X   float64 `json:"x"`
    Y   float64 `json:"y"`
    Z   float64 `json:"z"`
}

type MoveResponse struct {
    Code int `json:"code"`
}

func moveHandler(ctx node.Context) {
    req := &MoveRequest{}
    res := &MoveResponse{}

    defer func() {
        ctx.Response(res)
    }()

    if err := ctx.Parse(req); err != nil {
        res.Code = codes.InternalError.Code()
        return
    }

    // 更新位置
    updatePlayerPosition(req.UID, req.X, req.Y, req.Z)

    // 广播位置给附近玩家
    nearbyUIDs := getNearbyPlayers(req.UID)
    message.Broadcast(nearbyUIDs, MoveRoute, map[string]interface{}{
        "uid": req.UID,
        "x":   req.X,
        "y":   req.Y,
        "z":   req.Z,
    })

    res.Code = codes.OK.Code()
}

// 聊天
type ChatRequest struct {
    UID     int64  `json:"uid"`
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

    // 获取玩家名称
    name := loadPlayerName(req.UID)

    // 广播聊天消息给附近玩家
    nearbyUIDs := getNearbyPlayers(req.UID)
    message.Broadcast(nearbyUIDs, ChatRoute, map[string]interface{}{
        "uid":     req.UID,
        "name":    name,
        "content": req.Content,
    })

    res.Code = codes.OK.Code()
}

// 辅助函数
func validateToken(token string) bool {
    // 验证 Token
    return true
}

func loadPlayerName(uid int64) string {
    // 从数据库或缓存加载玩家名称
    return fmt.Sprintf("Player_%d", uid)
}

func savePlayerData(uid int64) {
    // 保存玩家数据
}

func updatePlayerPosition(uid int64, x, y, z float64) {
    // 更新玩家位置
}

func getNearbyPlayers(uid int64) []int64 {
    // 获取附近玩家 UID 列表
    return []int64{}
}
```

## 最佳实践

### ✅ 推荐做法

- 使用路由处理器处理消息
- 使用 `ctx.Parse()` 解析请求
- 使用 `ctx.Response()` 发送响应
- 使用 Container 统一管理组件
- 实现错误码规范

### ❌ 避免做法

- 在 Node 层执行业务逻辑（应在 Actor/Handler 中）
- 忽略错误处理
- 不使用 Container 直接调用 Serve()
- 硬编码路由 ID（使用常量定义）
