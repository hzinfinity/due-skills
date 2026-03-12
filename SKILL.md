---
name: due-skills
description: |
  Comprehensive knowledge base for due game server framework (v2.5.2).

  **Use this skill when:**
  - Working with due framework (any version, especially v2.5.2)
  - Building distributed game servers or real-time applications
  - Implementing Gate services (TCP/KCP/WebSocket) for client connections
  - Creating Node services with Actor model for stateful game logic
  - Setting up Mesh microservices for stateless business logic
  - Configuring service discovery (Consul/Etcd/Nacos)
  - Implementing event buses (Redis/NATS/Kafka/RabbitMQ)
  - Adding caching layers (Redis/Memcache)
  - Handling message routing, serialization, or RPC communication
  - Using Container pattern to manage component lifecycle
  - Implementing player session management and binding

  **Always consult this skill for due-related tasks** - it contains v2.5.2 specific API changes, correct module paths (github.com/dobyte/due/v2), and production-ready patterns that prevent common mistakes like using wrong package imports or outdated APIs.

  **Features:**
  - Complete architecture guides with Gate → Node → Mesh patterns
  - Actor model implementation examples with Router handlers
  - Multi-protocol support (TCP/KCP/WebSocket)
  - Production best practices for game servers
  - Common pitfall solutions (wrong imports, deprecated APIs)
  - Updated for due v2.5.2 API changes (Container pattern, Router handlers)
license: Apache-2.0
allowed-tools:
  - Read
  - Grep
  - Glob
---

# due Skills for AI Agents

This skill provides comprehensive due game server framework knowledge (v2.5.2), optimized for AI agents helping developers build production-ready distributed game servers. due is a lightweight, high-performance distributed game server framework (Apache 2.0 license), featuring standardized development patterns and proven deployment in enterprise game projects.

**v2.5.2 Key Changes:**
- Adjusted HTTP component route registration method
- Fixed Node component Push method ordering bug
- Fixed Node component WaitGroup count exception
- Refactored cluster internal RPC communication transport module
- Improved cluster debug mode causing RPC communication interruption
- Minor performance improvements

## 🎯 When to Use This Skill

Invoke this skill when working with due:
- **Creating game servers**: Gate services, Node services, or Mesh microservices
- **Protocol implementation**: TCP, KCP, or WebSocket client connections
- **Actor model**: Implementing stateful game logic with due Actor system
- **Service discovery**: Consul, Etcd, or Nacos integration
- **Event-driven architecture**: Redis, NATS, Kafka, or RabbitMQ event buses
- **Caching strategies**: Redis or Memcache integration
- **Message routing**: Custom route handling and serialization

## 📚 Knowledge Structure

This skill organizes due knowledge into focused modules. **Load specific guides as needed** rather than reading everything at once:

### Quick Start Guide
**File**: [getting-started/README.md](getting-started/README.md)
**When to load**: Starting a new due project, environment setup
**Contains**: Installation, first server, docker-compose setup, basic concepts

### Pattern Guides (Detailed Reference)

#### 1. Architecture Patterns
**File**: [references/architecture-patterns.md](references/architecture-patterns.md)
**When to load**: Understanding due architecture, designing server topology
**Contains**:
- Gate → Node → Mesh three-tier architecture
- Stateful vs Stateless service design
- Service discovery and registration
- Inter-service communication (gRPC/RPCX)

#### 2. Gate Development Patterns
**File**: [references/gate-patterns.md](references/gate-patterns.md)
**When to load**: Building gateway services for client connections
**Contains**:
- TCP protocol implementation
- KCP protocol implementation
- WebSocket protocol implementation
- Connection management and session handling
- Message packet format: `size + header + route + seq + message`
- Heartbeat mechanism

#### 3. Node Development Patterns
**File**: [references/node-patterns.md](references/node-patterns.md)
**When to load**: Implementing core game logic with Actor model
**Contains**:
- Actor model fundamentals in due
- Creating and managing Actors
- Message passing between Actors
- Stateful game logic implementation
- Actor lifecycle management

#### 4. Mesh Development Patterns
**File**: [references/mesh-patterns.md](references/mesh-patterns.md)
**When to load**: Building stateless microservices
**Contains**:
- Stateless service design
- Microservice communication patterns
- Load balancing strategies
- Service mesh integration

#### 5. Component Usage Patterns
**File**: [references/component-patterns.md](references/component-patterns.md)
**When to load**: Using due components (logging, caching, eventbus, etc.)
**Contains**:
- **Logging**: Console, File, Aliyun, Tencent log drivers
- **Cache**: Redis and Memcache integration
- **EventBus**: Redis, NATS, Kafka, RabbitMQ backends
- **Registry**: Consul, Etcd, Nacos service registration
- **Config**: Consul, Etcd, Nacos config centers with JSON/YAML/TOML/XML
- **Lock**: Distributed locks with Redis/Memcache
- **Crypto**: RSA and ECC encryption
- **Transport**: gRPC and RPCX communication

#### 6. Message Protocol
**File**: [references/protocol-patterns.md](references/protocol-patterns.md)
**When to load**: Defining custom message formats, serialization
**Contains**:
- Default packet format: `size + header + route + seq + message`
- Heartbeat packet: `size + header + extcode + heartbeat_time`
- Custom route and sequence number configuration
- Message serialization patterns
- Request/Response correlation with sequence numbers

### Supporting Resources

#### Best Practices
**File**: [best-practices/overview.md](best-practices/overview.md)
**When to load**: Production deployment, performance optimization
**Contains**: Configuration management, logging strategies, monitoring, scaling

#### Troubleshooting
**File**: [troubleshooting/common-issues.md](troubleshooting/common-issues.md)
**When to load**: Debugging errors, connection issues, runtime problems
**Contains**: Common error messages, solutions, configuration pitfalls

#### Project Structure
**File**: [references/project-structure.md](references/project-structure.md)
**When to load**: Understanding due directory layout, creating new projects
**Contains**:
```
due/
├── .docker/          # Docker configurations
├── cluster/          # Cluster management
├── component/        # Reusable components
├── config/           # Configuration management
├── core/             # Core framework logic
├── encoding/         # Message encoding/decoding
├── eventbus/         # Event bus implementations
├── log/              # Logging components
├── network/          # Network protocols
├── registry/         # Service registration
├── session/          # Session management
└── transport/        # RPC transport
```

## 🚀 Common Workflows

These workflows guide you through typical due development tasks:

### Creating a New Gate Service

**Steps:**
1. Create project directory structure
2. Initialize go.mod with due dependency
3. Create main.go with Gate configuration
4. Implement protocol handler (TCP/KCP/WS)
5. Define message routes and handlers
6. Start server with docker-compose

**Detailed guide**: [references/gate-patterns.md](references/gate-patterns.md)

### Implementing Actor-Based Game Logic

**Steps:**
1. Define Actor type and state structure
2. Implement Actor initialization
3. Register message handlers
4. Handle incoming messages in Actor loop
5. Send messages between Actors
6. Manage Actor lifecycle

**Detailed guide**: [references/node-patterns.md](references/node-patterns.md#actor-implementation)

### Setting Up Event-Driven Communication

**Steps:**
1. Choose EventBus backend (Redis/NATS/Kafka/RabbitMQ)
2. Configure EventBus in service config
3. Publish events from producers
4. Subscribe to events in consumers
5. Handle event serialization

**Detailed guide**: [references/component-patterns.md](references/component-patterns.md#eventbus)

### Configuring Service Discovery

**Steps:**
1. Choose registry backend (Consul/Etcd/Nacos)
2. Configure registry in service config
3. Register Gate and Node services
4. Enable service discovery for inter-service calls
5. Handle service health checks

**Detailed guide**: [references/architecture-patterns.md](references/architecture-patterns.md#service-discovery)

## ⚡ Key Principles

When generating or reviewing due code, always apply these principles:

### ✅ Always Follow

- **Three-tier separation**: Keep Gate (connection) → Node (logic) → Mesh (stateless) distinct
- **Actor model**: Use Actors for stateful game logic, not global state
- **Message routing**: Define clear route patterns for all messages
- **Configuration**: Externalize config with config center support
- **Service registration**: Always register services with discovery system
- **Error handling**: Use proper error wrapping and logging
- **Session management**: Store client state in Session, not global variables

### ❌ Never Do

- Put game logic directly in Gate layer (violates separation)
- Use global mutable state for game data (use Actors instead)
- Hard-code service addresses (use service discovery)
- Skip message validation in handlers
- Block Actor message loops with long operations
- Forget to handle connection close events

## 📖 Progressive Learning Path

Follow this path based on your needs:

### 🟢 New to due?

1. **Start here**: [getting-started/README.md](getting-started/README.md)
   Install due, understand architecture, run first example

2. **Try examples**: Clone due repository and run docker-compose examples
   See working Gate + Node setup with WebSocket

### 🟡 Building game servers?

1. **Review architecture**: [references/architecture-patterns.md](references/architecture-patterns.md)
   Understand Gate/Node/Mesh roles and communication

2. **Implement Actors**: [references/node-patterns.md](references/node-patterns.md)
   Learn Actor model for stateful game logic

3. **Add protocols**: [references/gate-patterns.md](references/gate-patterns.md)
   Choose TCP/KCP/WebSocket based on game type

### 🔵 Production deployment?

1. **Review best practices**: [best-practices/overview.md](best-practices/overview.md)
   Configuration, logging, monitoring, scaling

2. **Check common issues**: [troubleshooting/common-issues.md](troubleshooting/common-issues.md)
   Avoid typical mistakes and debugging tips

## 🔗 Related Resources

- **Official docs**: [https://github.com/dobyte/due](https://github.com/dobyte/due) - Source code and examples
- **Go documentation**: Standard library references for concurrent programming
- **Actor model**: Understanding Actor-based concurrency patterns

## 📝 Version Compatibility

- **Target version**: due v2.5.2 (Apache 2.0 licensed)
- **Go version**: Go 1.18 or later recommended
- **Module path**: github.com/dobyte/due/v2
- **Dependencies**: grpc, rpcx, redis, nats, kafka, rabbitmq drivers as needed

## 🚀 Quick Start (v2.5.2)

```bash
# Get due v2.5.2
go get -u github.com/dobyte/due/v2@latest

# Get required components
go get -u github.com/dobyte/due/locate/redis/v2@latest
go get -u github.com/dobyte/due/network/ws/v2@latest
go get -u github.com/dobyte/due/registry/consul/v2@latest
go get -u github.com/dobyte/due/transport/rpcx/v2@latest
```

**Gate Server Example (v2.5.2):**
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
   server := ws.NewServer()
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

**Node Server Example (v2.5.2):**
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

func initListen(proxy *node.Proxy) {
   proxy.Router().AddRouteHandler(routeID, isSync, handlerFunc)
}
```

---

**Quick invocation**: Use `/due-skills` or ask "How do I [task] with due?"
**Need help?** Reference the specific pattern guide for detailed examples.
