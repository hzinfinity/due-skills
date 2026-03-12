# due v2.5.2 Redis EventBus 配置指南

## 概述

Redis EventBus 是基于 Redis Pub/Sub 实现的事件总线，用于在 due v2.5.2 中解耦服务间通信。

## 安装依赖

```bash
go get -u github.com/dobyte/due/eventbus/redis/v2@latest
go get -u github.com/dobyte/due/v2/cache/redis@latest
go get -u github.com/redis/go-redis/v9@latest
```

## 配置步骤

### 1. 创建 Redis 客户端

```go
import (
    "github.com/dobyte/due/v2/cache/redis"
)

redisClient := redis.NewClient(&redis.Options{
    Addr:     "127.0.0.1:6379",
    Password: "",
    DB:       0,
    PoolSize: 10,
})
```

### 2. 创建 EventBus

```go
import (
    "github.com/dobyte/due/eventbus/redis/v2"
)

bus := redis.NewEventBus(
    redis.WithClient(redisClient),
    redis.WithChannel("game.events"), // 事件频道
)
```

### 3. 发布事件

```go
eventData := []byte(`{"uid": 12345, "username": "player1"}`)
err := bus.Publish(ctx, "user.login", eventData)
```

### 4. 订阅事件

```go
bus.Subscribe(ctx, "user.login", func(data []byte) {
    // 处理事件
})
```

## 架构图

```
┌─────────────┐                    ┌─────────────┐
│   Gate 1    │                    │   Gate 2    │
│  (Publisher)│                    │  (Publisher)│
└──────┬──────┘                    └──────┬──────┘
       │                                   │
       └──────────────┬────────────────────┘
                      │
                      ▼
            ┌─────────────────┐
            │  Redis Pub/Sub  │
            │  game.events    │
            └────────┬────────┘
                     │
       ┌─────────────┼─────────────┐
       ▼             ▼             ▼
┌──────────┐  ┌──────────┐  ┌──────────┐
│  Node    │  │   Mesh   │  │ Analytics│
│(Subscriber)│  │(Subscriber)│  │(Subscriber)│
└──────────┘  └──────────┘  └──────────┘
```

## 最佳实践

1. **事件命名**: 使用 `domain.event` 格式，如 `user.login`
2. **事件序列化**: 使用 JSON 格式序列化事件数据
3. **错误处理**: 始终检查 Publish 和 Subscribe 的错误
4. **事件版本**: 对于重要事件，考虑添加版本号
