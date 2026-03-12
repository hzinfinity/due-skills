package main

import (
	"github.com/duetalk/due/v2.5.2"
	"github.com/duetalk/due/v2.5.2/cluster"
	"github.com/duetalk/due/v2.5.2/cluster/consul"
	"github.com/duetalk/due/v2.5.2/entity"
	"github.com/duetalk/due/v2.5.2/gate"
	"github.com/duetalk/due/v2.5.2/gate/handler"
	"github.com/duetalk/due/v2.5.2/log"
	"github.com/duetalk/due/v2.5.2/node"
	"github.com/duetalk/due/v2.5.2/registry/redis"
	"github.com/duetalk/due/v2.5.2/transport"
	"github.com/duetalk/due/v2.5.2/transport/ws"
)

// 玩家实体
type Player struct {
	entity.Entity
	Name string `json:"name"`
}

// 登录请求
type LoginRequest struct {
	Name string `json:"name"`
}

// 登录响应
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// 聊天消息
type ChatMessage struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

// 玩家登录事件
type PlayerLoginEvent struct {
	PlayerID uint64 `json:"player_id"`
	Name     string `json:"name"`
}

// 玩家登出事件
type PlayerLogoutEvent struct {
	PlayerID uint64 `json:"player_id"`
	Name     string `json:"name"`
}

func main() {
	// 创建 due 应用
	app := due.New()

	// 配置 Consul 服务发现
	consulConfig := consul.Config{
		Address: "127.0.0.1:8500",
	}

	// 配置 Redis 定位
	redisConfig := redis.Config{
		Address: "127.0.0.1:6379",
	}

	// 创建 Gate - WebSocket 端口 8800
	g := app.NewGate(
		gate.WithID("gate"),
		gate.WithServer(&ws.Server{
			Addr: ":8800",
		}),
		gate.WithConsul(&consulConfig),
		gate.WithRegistry(&redisConfig),
	)

	// 创建 Node 处理消息
	n := app.NewNode(
		node.WithID("node"),
		node.WithConsul(&consulConfig),
		node.WithRegistry(&redisConfig),
	)

	// 定义路由
	defineRoutes(g, n)

	// 启动应用
	app.Run()
}

func defineRoutes(g gate.Gate, n node.Node) {
	// 玩家实体路由
	player := g.NewEntity[Player]()

	// 登录处理 - Gate 层
	player.Bind(handler.Message[LoginRequest](func(ctx handler.Context[Player], req *LoginRequest) {
		p := ctx.Entity()
		p.Name = req.Name

		log.Info("玩家登录", "name", p.Name, "id", p.ID())

		// 发送登录响应给客户端
		ctx.Response(&LoginResponse{
			Success: true,
			Message: "登录成功",
		})

		// 通知 Node 广播玩家登录事件
		n.Multicall().PlayerLogin(&PlayerLoginEvent{
			PlayerID: p.ID(),
			Name:     p.Name,
		})
	}))

	// 聊天消息处理 - Gate 层接收，Node 层广播
	player.Bind(handler.Message[ChatMessage](func(ctx handler.Context[Player], req *ChatMessage) {
		p := ctx.Entity()

		log.Info("收到聊天消息", "sender", p.Name, "content", req.Content)

		// 转发给 Node 进行广播处理
		n.Multicall().ChatMessage(req)
	}))

	// Node 上的广播处理
	n.Bind(func(ctx node.Context, event *PlayerLoginEvent) {
		log.Info("广播玩家登录", "name", event.Name, "playerID", event.PlayerID)

		// 广播给所有在线玩家
		ctx.Multicall(entity.All()).PlayerNotify(&ChatMessage{
			Sender:  "系统",
			Content: event.Name + " 已加入聊天室",
		})
	})

	n.Bind(func(ctx node.Context, msg *ChatMessage) {
		p := ctx.Entity[Player]()
		if p == nil {
			return
		}

		log.Info("广播聊天消息", "sender", p.Name, "content", msg.Content)

		// 广播给所有在线玩家
		ctx.Multicall(entity.All()).PlayerNotify(&ChatMessage{
			Sender:  p.Name,
			Content: msg.Content,
		})
	})

	// 玩家登出通知处理
	n.Bind(func(ctx node.Context, event *PlayerLogoutEvent) {
		log.Info("广播玩家登出", "name", event.Name, "playerID", event.PlayerID)

		// 广播给所有在线玩家
		ctx.Multicall(entity.All()).PlayerNotify(&ChatMessage{
			Sender:  "系统",
			Content: event.Name + " 已离开聊天室",
		})
	})

	// 客户端通知接收
	player.Bind(handler.Notify[ChatMessage](func(ctx handler.Context[Player], msg *ChatMessage) {
		p := ctx.Entity()
		log.Debug("收到通知消息", "player", p.Name, "sender", msg.Sender, "content", msg.Content)
		// 消息已通过 WebSocket 发送给客户端
	}))

	// Gate 关闭时处理玩家登出
	g.OnClose(func(ctx gate.Context) {
		p := ctx.Entity[Player]()
		if p != nil && p.Name != "" {
			log.Info("玩家登出", "name", p.Name, "id", p.ID())

			// 通知 Node 广播登出事件
			n.Multicall().PlayerLogout(&PlayerLogoutEvent{
				PlayerID: p.ID(),
				Name:     p.Name,
			})
		}
	})
}
