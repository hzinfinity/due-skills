# due 网关开发模式 (v2.5.2)

本文档详细介绍 due v2.5.2 框架中 Gate（网关）服务的开发模式。

## Gate 概述

Gate 服务是客户端与游戏服务器之间的入口，负责：
- 客户端连接管理（TCP/KCP/WebSocket）
- 消息协议编解码
- 会话管理
- 消息路由到 Node

## v2.5.2 Gate 完整示例

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
    server := ws.NewServer(
        ws.WithPort(8800),
    )

    // 创建定位器
    locator := redis.NewLocator(
        redis.WithAddr("127.0.0.1:6379"),
    )

    // 创建服务注册
    registry := consul.NewRegistry(
        consul.WithAddr("127.0.0.1:8500"),
    )

    // 创建 Gate 组件
    component := gate.NewGate(
        gate.WithID("gate-001"),
        gate.WithName("gate"),
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

## 协议支持 (v2.5.2)

due v2.5.2 支持三种主流协议：

### WebSocket 协议

适用于 H5 游戏、即时通讯等场景：

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
    container := due.NewContainer()

    // WebSocket 服务器
    server := ws.NewServer(
        ws.WithPort(8800),
        ws.WithMaxConnNum(10000),
    )

    // 定位器
    locator := redis.NewLocator()

    // 注册中心
    registry := consul.NewRegistry()

    // Gate 组件
    component := gate.NewGate(
        gate.WithServer(server),
        gate.WithLocator(locator),
        gate.WithRegistry(registry),
    )

    container.Add(component)
    container.Serve()
}
```

### TCP 协议

适用于对实时性要求高的游戏：

```go
package main

import (
    "github.com/dobyte/due/locate/redis/v2"
    "github.com/dobyte/due/network/tcp/v2"
    "github.com/dobyte/due/registry/consul/v2"
    "github.com/dobyte/due/v2"
    "github.com/dobyte/due/v2/cluster/gate"
)

func main() {
    container := due.NewContainer()

    // TCP 服务器
    server := tcp.NewServer(
        tcp.WithPort(9000),
        tcp.WithMaxConnNum(10000),
    )

    locator := redis.NewLocator()
    registry := consul.NewRegistry()

    component := gate.NewGate(
        gate.WithServer(server),
        gate.WithLocator(locator),
        gate.WithRegistry(registry),
    )

    container.Add(component)
    container.Serve()
}
```

### KCP 协议

适用于弱网络环境下的实时游戏：

```go
package main

import (
    "github.com/dobyte/due/locate/redis/v2"
    "github.com/dobyte/due/network/kcp/v2"
    "github.com/dobyte/due/registry/consul/v2"
    "github.com/dobyte/due/v2"
    "github.com/dobyte/due/v2/cluster/gate"
)

func main() {
    container := due.NewContainer()

    // KCP 服务器
    server := kcp.NewServer(
        kcp.WithPort(10000),
        kcp.WithMode(0),  // 0: 快速，1: 正常，2: 流畅
    )

    locator := redis.NewLocator()
    registry := consul.NewRegistry()

    component := gate.NewGate(
        gate.WithServer(server),
        gate.WithLocator(locator),
        gate.WithRegistry(registry),
    )

    container.Add(component)
    container.Serve()
}
```

## 消息协议 (v2.5.2)

### 默认数据包格式

```
┌────────┬────────┬──────┬─────┬─────────┐
│  size  │ header │ route│ seq │ message │
│ 2 bytes│1 byte  │2 bytes│2 bytes│ 可变   │
└────────┴────────┴──────┴─────┴─────────┘
```

- **size**: 数据包总长度（2 字节）
- **header**: 消息头标识（1 字节）
- **route**: 消息路由/类型（2 字节）
- **seq**: 序列号（2 字节，用于请求响应匹配）
- **message**: 消息体（可变长度）

### 心跳包格式

```
┌────────┬────────┬─────────┬──────────────┐
│  size  │ header │ extcode │ heartbeat_time│
│ 2 bytes│1 byte  │1 byte   │ 4 bytes      │
└────────┴────────┴─────────┴──────────────┘
```

## 会话管理 (v2.5.2)

在 due v2.5.2 中，Session 由框架自动管理，通过 `ctx.Session()` 获取：

```go
func handler(ctx gate.Context) {
    // 获取 Session
    session := ctx.Session()

    // 获取连接 ID
    cid := session.Cid()

    // 获取绑定的 UID
    uid := session.Uid()

    // 设置 Session 数据
    session.Set("key", "value")

    // 获取 Session 数据
    value := session.Get("key")
}
```

## 消息路由 (v2.5.2)

在 due v2.5.2 中，Gate 自动处理消息路由到 Node，无需手动配置 Match 函数。

消息自动根据 route 转发到对应的 Node Actor：

```
Client → Gate(自动路由) → Node(Proxy.Router 处理)
```

## 配置选项 (v2.5.2)

### Gate 组件配置

```go
component := gate.NewGate(
    gate.WithID("gate-001"),        // 服务 ID
    gate.WithName("gate"),          // 服务名称
    gate.WithServer(server),        // 网络服务器
    gate.WithLocator(locator),      // 定位器
    gate.WithRegistry(registry),    // 注册中心
)
```

### WebSocket 服务器配置

```go
server := ws.NewServer(
    ws.WithPort(8800),              // 端口
    ws.WithMaxConnNum(10000),       // 最大连接数
    ws.WithMsgSize(4096),           // 消息大小限制
    ws.WithHeartbeatInterval(30),   // 心跳间隔（秒）
)
```

### 定位器配置

```go
locator := redis.NewLocator(
    redis.WithAddr("127.0.0.1:6379"),
    redis.WithPassword("password"),
    redis.WithDB(0),
)
```

### 注册中心配置

```go
registry := consul.NewRegistry(
    consul.WithAddr("127.0.0.1:8500"),
    consul.WithID("gate-001"),
    consul.WithName("gate"),
)
```

## 完整示例 (v2.5.2)

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

    // WebSocket 服务器
    server := ws.NewServer(
        ws.WithPort(8800),
        ws.WithMaxConnNum(10000),
        ws.WithMsgSize(4096),
        ws.WithHeartbeatInterval(30),
    )

    // 定位器
    locator := redis.NewLocator(
        redis.WithAddr("127.0.0.1:6379"),
    )

    // 注册中心
    registry := consul.NewRegistry(
        consul.WithAddr("127.0.0.1:8500"),
        consul.WithID("gate-001"),
        consul.WithName("gate"),
    )

    // Gate 组件
    component := gate.NewGate(
        gate.WithID("gate-001"),
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

## v2.5.2 变化说明

**重要**: due v2.5.2 调整了组件使用方式：

1. **使用 Container 统一管理**：所有组件通过 `due.NewContainer()` 管理
2. **Gate 作为组件**：使用 `gate.NewGate()` 创建组件，而非直接创建 ws.NewGate()
3. **自动路由**：消息自动路由到 Node，无需手动 Match 函数
4. **Session 管理**：通过 `ctx.Session()` 获取会话

## 最佳实践

### ✅ 推荐做法

- 使用 Container 统一管理组件
- 配置服务注册实现服务发现
- 配置定位器用于消息路由
- 合理设置消息大小限制
- 配置心跳检测机制

### ❌ 避免做法

- 在 Gate 层执行业务逻辑
- 不使用 Container 直接调用 Serve()
- 忽略配置服务注册和定位器
- 不限制消息大小
