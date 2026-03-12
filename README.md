# due-skills (v2.5.2)

due 游戏服务器框架的 AI 助手技能知识库。

## 版本说明

**目标版本**: due v2.5.2

**v2.5.2 更新内容**:
- 调整 http 组件路由注册方法
- 修复 node 组件 Push 方法无序的 BUG
- 修复 node 组件 WaitGroup 计数异常 BUG
- 重构集群内部 RPC 通信传输模块
- 改进集群调试模式引发的 RPC 通信中断问题
- 小幅提升性能

## 安装

将此目录克隆到你的 skills 目录：

```bash
# 项目级别（推荐）
git clone https://github.com/your-org/due-skills.git .claude/skills/due-skills

# 个人级别（所有项目可用）
git clone https://github.com/your-org/due-skills.git ~/.claude/skills/due-skills
```

## 使用

在 Claude Code 中：
- **自动加载**：处理 due 框架相关文件时自动加载
- **手动调用**：输入 `/due-skills` 直接调用
- **带参数调用**：`/due-skills 创建 WebSocket 网关` 用于特定任务

## 目录结构

```
due-skills/
├── SKILL.md                    # 主要入口和元数据
├── README.md                   # 本文件
├── getting-started/            # 快速开始指南 (v2.5.2)
│   └── README.md
├── references/                 # 详细模式文档
│   ├── architecture-patterns.md    # 架构设计
│   ├── gate-patterns.md            # 网关开发 (v2.5.2)
│   ├── node-patterns.md            # Node 开发 (v2.5.2)
│   ├── mesh-patterns.md            # Mesh 开发 (v2.5.2)
│   ├── protocol-patterns.md        # 通信协议
│   ├── component-patterns.md       # 组件使用 (v2.5.2)
│   └── project-structure.md        # 项目结构
├── best-practices/             # 最佳实践
│   └── overview.md
├── troubleshooting/            # 常见问题
│   └── common-issues.md
└── examples/                   # 示例代码 (v2.5.2)
    └── README.md
```

## 内容概览

### 快速开始 (v2.5.2)
- 安装 due v2.5.2 框架
- 使用 Container 管理组件
- 创建 Gate 和 Node 服务

### 架构设计
- Gate → Node → Mesh 三层架构
- 服务发现（Consul/Etcd/Nacos）
- Actor 模型（使用路由处理器）

### 开发模式
- WebSocket/TCP/KCP 网关开发（使用 `gate.NewGate()`）
- 路由处理器模式（`proxy.Router().AddRouteHandler()`）
- Mesh 微服务开发（使用 `mesh.NewMesh()`）
- 消息协议设计

### 组件使用 (v2.5.2)
- **日志**（Console/File/Aliyun/Tencent）
- **配置**（Consul/Etcd/Nacos）
- **缓存**（Redis/Memcache）
- **事件总线**（Redis/NATS/Kafka/RabbitMQ）
- **服务注册**（Consul/Etcd/Nacos）
- **加密**（RSA/ECC）

### 最佳实践
- 配置管理
- 日志规范
- 性能优化
- 错误处理
- 安全实践

### 常见问题
- 安装问题
- 连接问题
- Actor 问题
- 服务发现问题
- 性能问题

## v2.5.2 关键 API 变化

### Gate 组件

```go
// v2.5.2 - 使用 Container 和 gate.NewGate()
container := due.NewContainer()
component := gate.NewGate(
    gate.WithServer(server),
    gate.WithLocator(locator),
    gate.WithRegistry(registry),
)
container.Add(component)
container.Serve()
```

### Node 组件

```go
// v2.5.2 - 使用路由处理器
container := due.NewContainer()
component := node.NewNode(
    node.WithLocator(locator),
    node.WithRegistry(registry),
)
component.Proxy().Router().AddRouteHandler(routeID, isSync, handlerFunc)
container.Add(component)
container.Serve()
```

### Mesh 组件

```go
// v2.5.2 - 使用 mesh.NewMesh()
container := due.NewContainer()
component := mesh.NewMesh(
    mesh.WithLocator(locator),
    mesh.WithRegistry(registry),
    mesh.WithTransporter(transporter),
)
component.AddServiceProvider("service.name", "Service", provider)
container.Add(component)
container.Serve()
```

## 模块路径

due v2.5.2 使用以下模块路径：

```
github.com/dobyte/due/v2              # 主框架
github.com/dobyte/due/locate/redis/v2 # 定位器
github.com/dobyte/due/network/ws/v2   # WebSocket
github.com/dobyte/due/network/tcp/v2  # TCP
github.com/dobyte/due/registry/consul/v2 # 注册中心
github.com/dobyte/due/transport/rpcx/v2 # 传输器
```

## 许可证

Apache-2.0 License（与 due v2.5.2 保持一致）
