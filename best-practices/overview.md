# due 最佳实践

本文档介绍 due 框架在生产环境中的最佳实践。

## 配置管理

### 环境分离

```yaml
# config/dev.yaml
development:
  gate:
    port: 8800
    log_level: debug
  database:
    host: localhost
    port: 3306

# config/prod.yaml
production:
  gate:
    port: 8800
    log_level: info
  database:
    host: db.cluster.internal
    port: 3306
```

### 配置加载

```go
func LoadConfig(env string) (*Config, error) {
    source := consul.NewSource(
        consul.WithAddr(getConsulAddr()),
        consul.WithPath(fmt.Sprintf("config/game/%s", env)),
    )

    cfg := config.NewConfig(config.WithSource(source))
    if err := cfg.Load(); err != nil {
        return nil, err
    }

    var c Config
    if err := cfg.Scan(&c); err != nil {
        return nil, err
    }

    return &c, nil
}
```

## 日志规范

### 日志级别使用

```go
// Debug: 调试信息，仅开发环境输出
log.Debug("进入函数", "func", "handleLogin")

// Info: 关键业务流程
log.Info("用户登录成功", "uid", userID, "ip", clientIP)

// Warn: 可恢复的异常
log.Warn("数据库连接超时，使用备用连接", "retry", retryCount)

// Error: 不可恢复的错误
log.Error("数据库连接失败", "error", err, "host", dbHost)
```

### 日志字段规范

```go
// 统一使用结构化日志
log.WithFields(log.Fields{
    "uid":      userID,
    "username": username,
    "trace_id": traceID,
}).Info("用户操作")
```

### 日志轮转

```go
driver := file.NewDriver(
    file.WithDir("./logs"),
    file.WithFilename("game.log"),
    file.WithMaxSize(100 * 1024 * 1024),   // 100MB 轮转
    file.WithMaxBackups(7),                 // 保留 7 天
    file.WithMaxAge(30 * 24 * time.Hour),   // 最多保留 30 天
    file.WithCompress(true),                // 压缩旧日志
)
```

## 性能优化

### 连接池配置

```go
// Redis 连接池
redisDriver := redis.NewDriver(
    redis.WithAddr("127.0.0.1:6379"),
    redis.WithPoolSize(50),           // 连接池大小
    redis.WithMinIdleConns(10),       // 最小空闲连接
    redis.WithPoolTimeout(time.Second * 30),
    redis.WithIdleTimeout(time.Minute * 5),
)

// 数据库连接池
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(20)
db.SetConnMaxLifetime(time.Hour)
```

### Actor 优化

```go
// 合理设置 Worker 数量
node := NewNode(
    node.WithWorkerSize(runtime.NumCPU() * 2),  // CPU 核心数的 2 倍
    node.WithQueueSize(1024),                    // 消息队列大小
)

// 避免阻塞 Actor 消息处理
func (a *PlayerActor) OnMessage(ctx context.Context, message *actor.Message) {
    // ❌ 错误：阻塞操作
    // time.Sleep(time.Second * 5)
    // heavyComputation()

    // ✅ 正确：异步处理
    go a.heavyComputation()
}
```

### 消息批处理

```go
// 批量插入
func (a *PlayerActor) batchInsert(data []*PlayerData) {
    // 积累一定数量后批量处理
    if len(a.pending) >= 100 {
        a.flush()
    }
    a.pending = append(a.pending, data...)
}
```

## 错误处理

### 统一错误包装

```go
import "github.com/pkg/errors"

// ❌ 错误：直接返回错误
return err

// ✅ 正确：包装错误上下文
return errors.Wrapf(err, "处理登录失败，uid=%d", userID)
```

### 错误码规范

```go
const (
    // 通用错误 1000-1999
    CodeOK             = 0
    CodeInternalError  = 1001
    CodeInvalidParam   = 1002
    CodeTimeout        = 1003

    // 用户错误 2000-2999
    CodeUserNotFound   = 2001
    CodeUserBanned     = 2002
    CodeSessionExpired = 2003

    // 游戏错误 3000-3999
    CodeItemNotFound   = 3001
    CodeNotEnoughGold  = 3002
)

type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}
```

## 安全实践

### 输入验证

```go
type LoginRequest struct {
    Username string `json:"username" validate:"required,min=3,max=20"`
    Password string `json:"password" validate:"required,min=6,max=32"`
}

func (a *PlayerActor) handleLogin(ctx context.Context, message *actor.Message) {
    var req LoginRequest
    if err := json.Unmarshal(message.Data, &req); err != nil {
        log.Error("参数解析失败", "error", err)
        return
    }

    if err := validate.Struct(&req); err != nil {
        log.Error("参数验证失败", "error", err)
        return
    }

    // 处理登录
}
```

### 敏感数据加密

```go
// 密码加密存储
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 敏感数据传输加密
ciphertext, err := rsa.Encrypt(publicKey, sensitiveData)
```

### 限流防护

```go
import "golang.org/x/time/rate"

// 创建限流器
limiter := rate.NewLimiter(rate.Every(time.Second), 100)  // 每秒 100 个请求

func (a *PlayerActor) handleRequest(ctx context.Context, message *actor.Message) {
    if !limiter.Allow() {
        log.Warn("请求限流", "uid", a.uid)
        return
    }

    // 处理请求
}
```

## 监控告警

### 指标收集

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    loginCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "game_login_total",
            Help: "Total login count",
        },
        []string{"status"},
    )

    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "game_request_duration_seconds",
            Help:    "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"route"},
    )
)

// 注册指标
prometheus.MustRegister(loginCounter, requestDuration)
```

### 健康检查

```go
type HealthStatus struct {
    Status   string            `json:"status"`
    Checks   map[string]string `json:"checks"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    status := HealthStatus{
        Status: "healthy",
        Checks: make(map[string]string),
    }

    // 检查数据库
    if err := db.Ping(); err != nil {
        status.Checks["database"] = "unhealthy"
        status.Status = "unhealthy"
    } else {
        status.Checks["database"] = "healthy"
    }

    // 检查 Redis
    if _, err := redis.Ping(); err != nil {
        status.Checks["redis"] = "unhealthy"
        status.Status = "unhealthy"
    } else {
        status.Checks["redis"] = "healthy"
    }

    json.NewEncoder(w).Encode(status)
}
```

## 部署实践

### Docker 部署

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gate ./cmd/gate

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/gate .
COPY config/ ./config/

EXPOSE 8800
CMD ["./gate"]
```

### Docker Compose

```yaml
version: '3'
services:
  gate:
    build: .
    ports:
      - "8800:8800"
    environment:
      - ENV=production
      - CONSUL_ADDR=consul:8500
    depends_on:
      - consul
      - redis

  node:
    build: .
    environment:
      - ENV=production
      - CONSUL_ADDR=consul:8500
    depends_on:
      - consul
      - redis
      - mysql

  consul:
    image: consul:latest
    ports:
      - "8500:8500"

  redis:
    image: redis:latest
    ports:
      - "6379:6379"

  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=root
    ports:
      - "3306:3306"
```

### 优雅关闭

```go
func main() {
    gate := ws.NewGate(...)

    // 监听信号
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        sig := <-make(chan os.Signal, 1)
        log.Info("收到关闭信号", "signal", sig)
        cancel()
    }()

    // 启动服务
    go gate.Serve()

    // 等待关闭
    <-ctx.Done()

    // 优雅关闭
    gate.Shutdown(context.Background())
}
```

## 最佳实践检查清单

### 开发阶段

- [ ] 使用配置中心管理配置
- [ ] 实现结构化日志
- [ ] 定义统一错误码
- [ ] 实现输入验证
- [ ] 编写单元测试

### 部署阶段

- [ ] 使用 Docker 容器化
- [ ] 配置健康检查
- [ ] 设置日志轮转
- [ ] 配置监控指标
- [ ] 实现优雅关闭

### 运维阶段

- [ ] 配置告警规则
- [ ] 定期检查日志
- [ ] 监控系统资源
- [ ] 备份关键数据
- [ ] 定期安全审计
