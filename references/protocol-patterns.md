# due 通信协议模式

本文档详细介绍 due 框架中的通信协议格式和序列化模式。

## 数据包格式

### 默认格式

due 默认使用以下数据包格式：

```
┌──────────┬───────────┬──────────┬───────────┬───────────┐
│   size   │  header   │   route  │    seq    │  message  │
│ 2 bytes  │ 1 byte    │ 2 bytes  │ 2 bytes   │ N bytes   │
└──────────┴───────────┴──────────┴───────────┴───────────┘
```

#### 字段说明

| 字段 | 大小 | 说明 |
|------|------|------|
| size | 2 bytes | 数据包总长度（不包含 size 本身） |
| header | 1 byte | 消息头标识（用于区分消息类型） |
| route | 2 bytes | 消息路由/类型 ID |
| seq | 2 bytes | 序列号，用于请求 - 响应匹配 |
| message | N bytes | 消息体（序列化后的数据） |

### 心跳包格式

```
┌──────────┬───────────┬───────────┬────────────────┐
│   size   │  header   │  extcode  │  heartbeat_time│
│ 2 bytes  │ 1 byte    │ 1 byte    │ 4 bytes       │
└──────────┴───────────┴───────────┴────────────────┘
```

#### 字段说明

| 字段 | 大小 | 说明 |
|------|------|------|
| size | 2 bytes | 心跳包总长度 |
| header | 1 byte | 心跳标识（固定值） |
| extcode | 1 byte | 扩展码 |
| heartbeat_time | 4 bytes | 心跳时间戳 |

## 序列化器

### JSON 序列化器

```go
import "github.com/dobyte/due/core/serializer/json"

serializer := json.NewSerializer()

// 序列化
data, err := serializer.Marshal(message)

// 反序列化
err := serializer.Unmarshal(data, &message)
```

### Protobuf 序列化器

```go
import "github.com/dobyte/due/core/serializer/protobuf"

serializer := protobuf.NewSerializer()

// 序列化
data, err := serializer.Marshal(protoMessage)

// 反序列化
err := serializer.Unmarshal(data, &protoMessage)
```

### 自定义序列化器

实现 Serializer 接口：

```go
type Serializer interface {
    Marshal(v interface{}) ([]byte, error)
    Unmarshal(data []byte, v interface{}) error
}

// 自定义实现
type CustomSerializer struct{}

func (s *CustomSerializer) Marshal(v interface{}) ([]byte, error) {
    // 自定义序列化逻辑
    return data, nil
}

func (s *CustomSerializer) Unmarshal(data []byte, v interface{}) error {
    // 自定义反序列化逻辑
    return nil
}
```

## 协议配置

### Gate 协议配置

```go
gate := ws.NewGate(
    ws.WithSerializer(json.NewSerializer()),
    // 其他配置...
)
```

### Node 协议配置

```go
node := NewNode(
    node.WithSerializer(protobuf.NewSerializer()),
    // 其他配置...
)
```

## 请求 - 响应模式

### 客户端请求

```go
// 发送请求（带 seq）
seq := generateSeq()
conn.Send(&Message{
    Route: 1,
    Seq:   seq,
    Data:  requestData,
})

// 等待响应
response := waitForResponse(seq, timeout)
```

### 服务端响应

```go
func (a *PlayerActor) handleRequest(ctx context.Context, message *actor.Message) {
    // 处理请求
    result := processData(message.Data)

    // 发送响应（使用相同的 seq）
    a.session.Send(&actor.Message{
        Route: message.Route,
        Seq:   message.Seq,
        Data:  result,
    })
}
```

## 消息路由

### 路由定义

```go
// 定义消息路由常量
const (
    RouteLogin   = 1
    RouteMove    = 2
    RouteChat    = 3
    RouteLogout  = 4
)
```

### 路由匹配

```go
// Gate 层路由匹配
gate.Match(RouteLogin, func(ctx context.Context, conn *ws.Conn, message []byte) {
    // 处理登录
})

gate.Match(RouteMove, func(ctx context.Context, conn *ws.Conn, message []byte) {
    // 处理移动
})

// Actor 层路由匹配
func (a *PlayerActor) OnMessage(ctx context.Context, message *actor.Message) {
    switch message.Route {
    case RouteLogin:
        a.handleLogin(ctx, message)
    case RouteMove:
        a.handleMove(ctx, message)
    }
}
```

## 自定义协议

### 自定义 Encoder/Decoder

```go
type Encoder interface {
    Encode(message []byte) ([]byte, error)
}

type Decoder interface {
    Decode(conn net.Conn) ([]byte, error)
}

// 自定义实现
type CustomEncoder struct{}

func (e *CustomEncoder) Encode(message []byte) ([]byte, error) {
    // 添加自定义头部
    header := []byte{0x01, 0x02}
    return append(header, message...), nil
}

type CustomDecoder struct{}

func (d *CustomDecoder) Decode(conn net.Conn) ([]byte, error) {
    // 读取并解析自定义头部
    header := make([]byte, 2)
    _, err := io.ReadFull(conn, header)
    if err != nil {
        return nil, err
    }

    // 读取消息体
    // ...
    return message, nil
}
```

### 使用自定义协议

```go
gate := ws.NewGate(
    ws.WithEncoder(&CustomEncoder{}),
    ws.WithDecoder(&CustomDecoder{}),
)
```

## 消息体设计

### 基础消息结构

```go
type Message struct {
    Route int64       `json:"route"`  // 路由
    Seq   int64       `json:"seq"`    // 序列号
    Data  interface{} `json:"data"`   // 数据体
}
```

### 请求消息

```go
type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Token    string `json:"token,omitempty"`
}

type MoveRequest struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
    Z float64 `json:"z"`
}

type ChatRequest struct {
    Channel   int    `json:"channel"`
    Content   string `json:"content"`
    TargetID  int64  `json:"target_id,omitempty"`
}
```

### 响应消息

```go
type LoginResponse struct {
    Code     int    `json:"code"`
    Message  string `json:"message"`
    UID      int64  `json:"uid"`
    Session  string `json:"session"`
}

type MoveResponse struct {
    Code    int     `json:"code"`
    UID     int64   `json:"uid"`
    X       float64 `json:"x"`
    Y       float64 `json:"y"`
    Z       float64 `json:"z"`
}

type ChatResponse struct {
    Code    int    `json:"code"`
    UID     int64  `json:"uid"`
    Name    string `json:"name"`
    Content string `json:"content"`
    Time    int64  `json:"time"`
}
```

## 错误处理

### 错误码定义

```go
const (
    CodeOK         = 0
    CodeError      = 1
    CodeAuthFailed = 1001
    CodeNotFound   = 1002
    CodeTimeout    = 1003
)
```

### 错误响应

```go
type ErrorResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Route   int64  `json:"route"`
    Seq     int64  `json:"seq"`
}

func sendError(conn *ws.Conn, route int64, seq int64, code int, msg string) {
    conn.Send(&ErrorResponse{
        Code:    code,
        Message: msg,
        Route:   route,
        Seq:     seq,
    })
}
```

## 最佳实践

### ✅ 推荐做法

- 使用 Protobuf 进行高效序列化
- 为所有消息定义明确的路由 ID
- 使用 seq 进行请求 - 响应匹配
- 实现心跳机制检测连接状态
- 为错误码建立统一规范

### ❌ 避免做法

- 在消息体中传输敏感数据（使用加密）
- 忽略 seq 导致响应无法匹配
- 消息体过大（考虑分片传输）
- 不验证消息格式
