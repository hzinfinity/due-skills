# due 常见问题与解决方案

本文档收录 due 框架开发中的常见问题和解决方案。

## 安装问题

### 问题：go get 下载失败

**现象**：
```
go get github.com/dobyte/due: dial tcp: lookup github.com: no such host
```

**解决方案**：
```bash
# 配置 GOPROXY
go env -w GOPROXY=https://goproxy.cn,direct

# 清除缓存后重试
go clean -modcache
go get github.com/dobyte/due
```

### 问题：依赖冲突

**现象**：
```
go: found github.com/dobyte/due in github.com/dobyte/due v1.0.0
but required version is v1.2.0
```

**解决方案**：
```bash
# 清理模块缓存
go clean -modcache

# 更新依赖
go get -u github.com/dobyte/due@latest

# 整理依赖
go mod tidy
```

## 连接问题

### 问题：WebSocket 连接失败

**现象**：
```
WebSocket connection failed: HTTP status code 404
```

**解决方案**：
```go
// 检查 WebSocket 路径配置
gate := ws.NewGate(
    ws.WithPath("/ws"),  // 确保路径匹配
    // ...
)

// 检查防火墙/代理设置
// 确保端口开放
```

### 问题：连接频繁断开

**现象**：客户端连接后不久就断开

**解决方案**：
```go
// 1. 检查心跳配置
gate := ws.NewGate(
    ws.WithHeartbeatInterval(30),  // 30 秒心跳
)

// 2. 客户端实现心跳
setInterval(() => {
    ws.send(heartbeatPacket);
}, 25000);  // 25 秒发送一次心跳

// 3. 检查超时设置
gate := ws.NewGate(
    ws.WithReadTimeout(time.Minute * 2),
    ws.WithWriteTimeout(time.Minute * 2),
)
```

### 问题：KCP 连接不稳定

**现象**：KCP 连接质量差，延迟高

**解决方案**：
```go
// 调整 KCP 模式
gate := kcp.NewGate(
    kcp.WithMode(0),  // 0: 快速模式，1: 正常模式，2: 流畅模式
    kcp.WithFlushInterval(10),  // 刷新间隔（毫秒）
    kcp.WithNoDelay(1, 10, 2, 1),  // 无延迟配置
)

// 根据网络环境选择模式：
// - 优质网络：模式 0
// - 一般网络：模式 1
// - 弱网络：模式 2
```

## Actor 问题

### 问题：Actor 消息处理阻塞

**现象**：消息处理延迟高，响应慢

**解决方案**：
```go
// ❌ 错误：阻塞操作
func (a *PlayerActor) OnMessage(ctx context.Context, message *actor.Message) {
    heavyComputation()  // 耗时操作阻塞消息循环
}

// ✅ 正确：异步处理
func (a *PlayerActor) OnMessage(ctx context.Context, message *actor.Message) {
    switch message.Route {
    case HeavyRoute:
        go a.heavyComputation()  // 异步处理
    default:
        a.handleMessage(message)
    }
}
```

### 问题：Actor 状态丢失

**现象**：Actor 重启后状态数据丢失

**解决方案**：
```go
type PlayerActor struct {
    *actor.Base
    data *PlayerData
}

// 初始化时加载状态
func (a *PlayerActor) OnInit() {
    a.data = loadFromDB(a.ID())
}

// 定期保存状态
func (a *PlayerActor) OnSave() {
    saveToDB(a.data)
}

// 销毁前保存状态
func (a *PlayerActor) OnDestroy() {
    saveToDB(a.data)
}
```

### 问题：Actor 泄漏（未释放）

**现象**：内存持续增长，Actor 数量不断增加

**解决方案**：
```go
// 1. 设置闲置超时
actor.SetIdleTimeout(uid, time.Minute*30)

// 2. 实现 OnDestroy 清理资源
func (a *PlayerActor) OnDestroy() {
    a.cleanup()
}

// 3. 监控 Actor 数量
log.Info("Actor 数量", "count", actor.Count())

// 4. 处理 OnDisconnect 事件
func (a *PlayerActor) OnDisconnect() {
    // 玩家断开连接，可以触发销毁
    actor.Destroy(a.ID())
}
```

## 服务发现问题

### 问题：服务注册失败

**现象**：服务启动后未在注册中心出现

**解决方案**：
```go
// 1. 检查注册中心地址
reg := consul.NewRegistry(
    consul.WithAddr("127.0.0.1:8500"),  // 确保地址正确
)

// 2. 检查服务 ID 和名称
reg := consul.NewRegistry(
    consul.WithID("gate-001"),
    consul.WithName("gate"),
)

// 3. 检查健康检查配置
reg := consul.NewRegistry(
    consul.WithHealthCheck(true),
    consul.WithCheckInterval(time.Second * 10),
)

// 4. 查看注册中心日志
// consul logs | grep gate
```

### 问题：服务发现超时

**现象**：调用其他服务时超时

**解决方案**：
```go
// 1. 配置传输超时
trans := grpc.NewTransport(
    grpc.WithDialTimeout(time.Second),
    grpc.WithTimeout(time.Second * 3),
)

// 2. 检查服务健康状态
services, _ := reg.Discover("node")
for _, s := range services {
    if !s.Healthy {
        log.Warn("服务不健康", "id", s.ID)
    }
}

// 3. 实现重试机制
func callWithRetry(uid int64, msg *actor.Message, maxRetry int) error {
    for i := 0; i < maxRetry; i++ {
        if err := actor.Send(uid, msg); err == nil {
            return nil
        }
        time.Sleep(time.Millisecond * 100 * (1 << i))
    }
    return errors.New("max retry exceeded")
}
```

## 数据库问题

### 问题：数据库连接池耗尽

**现象**：`too many connections` 错误

**解决方案**：
```go
// 1. 配置连接池
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(20)
db.SetConnMaxLifetime(time.Hour)

// 2. 使用连接池监控
stats := db.Stats()
log.Info("连接池状态",
    "max_open", stats.MaxOpenConnections,
    "open", stats.OpenConnections,
    "in_use", stats.InUse,
    "idle", stats.Idle,
)

// 3. 确保释放连接
rows, _ := db.Query("SELECT ...")
defer rows.Close()
```

### 问题：Redis 连接超时

**现象**：`i/o timeout` 错误

**解决方案**：
```go
// 1. 配置连接池和超时
driver := redis.NewDriver(
    redis.WithAddr("127.0.0.1:6379"),
    redis.WithPoolSize(50),
    redis.WithDialTimeout(time.Second * 5),
    redis.WithReadTimeout(time.Second * 3),
    redis.WithWriteTimeout(time.Second * 3),
)

// 2. 实现重试机制
func redisGetWithRetry(key string, maxRetry int) (string, error) {
    var lastErr error
    for i := 0; i < maxRetry; i++ {
        val, err := cache.Get(key)
        if err == nil {
            return val, nil
        }
        lastErr = err
        time.Sleep(time.Millisecond * 100 * (1 << i))
    }
    return "", lastErr
}
```

## 消息问题

### 问题：消息丢失

**现象**：发送的消息未到达接收方

**解决方案**：
```go
// 1. 使用可靠消息队列
bus := kafka.NewEventBus(
    kafka.WithBrokers([]string{"127.0.0.1:9092"}),
    kafka.WithTopic("game-events"),
    kafka.WithConsumerGroup("game-consumer"),
)

// 2. 实现消息确认机制
bus.Subscribe("user.login", func(event *Event) {
    // 处理事件
    bus.Ack(event.ID)  // 确认消息
})

// 3. 实现消息持久化
func (a *PlayerActor) handleMessage(msg *actor.Message) {
    // 先持久化消息
    saveMessage(msg)
    // 再处理
    process(msg)
    // 标记为已处理
    markProcessed(msg.ID)
}
```

### 问题：消息重复

**现象**：同一条消息被处理多次

**解决方案**：
```go
// 实现幂等性处理
type PlayerActor struct {
    *actor.Base
    processed map[int64]bool  // 已处理的消息 ID
}

func (a *PlayerActor) OnMessage(ctx context.Context, message *actor.Message) {
    // 检查是否已处理
    if a.processed[message.Seq] {
        log.Info("重复消息，跳过", "seq", message.Seq)
        return
    }

    // 标记为已处理
    a.processed[message.Seq] = true

    // 处理消息
    a.process(message)
}
```

## 性能问题

### 问题：高并发下性能下降

**现象**：并发用户增多后响应变慢

**解决方案**：
```go
// 1. 增加 Worker 数量
node := NewNode(
    node.WithWorkerSize(runtime.NumCPU() * 4),
)

// 2. 优化消息队列
node := NewNode(
    node.WithQueueSize(2048),  // 增加队列大小
)

// 3. 使用缓存
func getPlayerData(uid int64) (*PlayerData, error) {
    // 先查缓存
    if data, err := cache.Get(fmt.Sprintf("player:%d", uid)); err == nil {
        return data, nil
    }

    // 缓存未命中，查数据库
    data := loadFromDB(uid)

    // 写入缓存
    cache.Set(fmt.Sprintf("player:%d", uid), data, time.Hour)

    return data, nil
}
```

### 问题：内存泄漏

**现象**：服务运行一段时间后内存持续增长

**解决方案**：
```go
// 1. 使用 pprof 分析
import _ "net/http/pprof"

go func() {
    http.ListenAndServe("localhost:6060", nil)
}()

// 访问 http://localhost:6060/debug/pprof/

// 2. 检查 goroutine 泄漏
runtime.NumGoroutine()

// 3. 确保资源释放
func processData(data []byte) {
    defer func() {
        // 清理资源
    }()
    // 处理逻辑
}
```

## 调试技巧

### 启用调试日志

```go
log.SetLevel(log.DebugLevel)

// 添加请求追踪
log.WithFields(log.Fields{
    "trace_id": generateTraceID(),
    "uid":      userID,
    "route":    route,
}).Debug("请求详情")
```

### 使用 pprof 分析

```go
import (
    _ "net/http/pprof"
    "net/http"
)

func main() {
    // 启动 pprof
    go http.ListenAndServe("localhost:6060", nil)

    // 启动服务
    gate.Serve()
}

// 使用方式：
// go tool pprof http://localhost:6060/debug/pprof/heap
// go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### 性能分析

```bash
# CPU 分析
go tool pprof cpu.prof

# 内存分析
go tool pprof heap.prof

# 阻塞分析
go tool pprof block.prof
```

## 快速诊断清单

遇到问题时，按以下顺序检查：

- [ ] 检查日志输出（错误信息、堆栈）
- [ ] 检查配置（端口、地址、超时）
- [ ] 检查服务状态（注册中心、健康检查）
- [ ] 检查网络连接（防火墙、端口开放）
- [ ] 检查资源使用（CPU、内存、连接数）
- [ ] 检查依赖版本（go mod tidy）
- [ ] 使用 pprof 分析性能问题
