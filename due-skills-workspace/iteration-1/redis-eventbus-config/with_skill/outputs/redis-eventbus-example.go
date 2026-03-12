// due v2.5.2 Redis EventBus 配置示例
// 用于解耦服务间通信，发布/订阅事件模式

package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/dobyte/due/eventbus/redis/v2"
	"github.com/dobyte/due/v2/cache/redis"
	"github.com/redis/go-redis/v9"
)

// UserLoginEvent 用户登录事件结构
type UserLoginEvent struct {
	UID      int64     `json:"uid"`
	Username string    `json:"username"`
	IP       string    `json:"ip"`
	LoginAt  time.Time `json:"login_at"`
}

// UserLogoutEvent 用户登出事件结构
type UserLogoutEvent struct {
	UID       int64     `json:"uid"`
	LogoutAt  time.Time `json:"logout_at"`
	Reason    string    `json:"reason"`
}

func main() {
	ctx := context.Background()

	// ==========================================
	// 步骤 1: 创建 Redis 客户端
	// ==========================================
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // 如果有密码请设置
		DB:       0,
		PoolSize: 10,
	})

	// 测试 Redis 连接
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis 连接失败：%v", err)
	}
	log.Println("Redis 连接成功")

	// ==========================================
	// 步骤 2: 创建 Redis EventBus
	// ==========================================
	// 注意：due v2.5.2 的 Redis EventBus 使用 channel 模式
	bus := redis.NewEventBus(
		redis.WithClient(redisClient),
		redis.WithChannel("game.events"), // 事件频道名称
	)

	// ==========================================
	// 步骤 3: 订阅事件 (在消费者服务中)
	// ==========================================

	// 订阅 user.login 事件
	// 可以在多个不同的服务中订阅相同的事件，实现事件广播
	bus.Subscribe(ctx, "user.login", func(data []byte) {
		var event UserLoginEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("解析 user.login 事件失败：%v", err)
			return
		}
		// 处理用户登录事件
		log.Printf("[事件处理器] 用户登录 - UID: %d, 用户名：%s, IP: %s",
			event.UID, event.Username, event.IP)

		// 在这里可以执行：
		// - 发送登录通知
		// - 更新在线用户统计
		// - 记录登录日志
		// - 触发其他业务流程
	})

	// 订阅 user.logout 事件
	bus.Subscribe(ctx, "user.logout", func(data []byte) {
		var event UserLogoutEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("解析 user.logout 事件失败：%v", err)
			return
		}
		log.Printf("[事件处理器] 用户登出 - UID: %d, 原因：%s",
			event.UID, event.Reason)
	})

	// ==========================================
	// 步骤 4: 发布事件 (在生产者服务中)
	// ==========================================

	// 发布 user.login 事件
	loginEvent := UserLoginEvent{
		UID:      12345,
		Username: "player1",
		IP:       "192.168.1.100",
		LoginAt:  time.Now(),
	}

	loginData, _ := json.Marshal(loginEvent)
	if err := bus.Publish(ctx, "user.login", loginData); err != nil {
		log.Printf("发布 user.login 事件失败：%v", err)
	} else {
		log.Println("user.login 事件发布成功")
	}

	// 发布 user.logout 事件
	logoutEvent := UserLogoutEvent{
		UID:      12345,
		LogoutAt: time.Now(),
		Reason:   "user_active_logout",
	}

	logoutData, _ := json.Marshal(logoutEvent)
	if err := bus.Publish(ctx, "user.logout", logoutData); err != nil {
		log.Printf("发布 user.logout 事件失败：%v", err)
	} else {
		log.Println("user.logout 事件发布成功")
	}

	// 保持程序运行以接收事件
	time.Sleep(time.Second * 2)
	log.Println("示例完成")
}

// ==========================================
// 实际项目中的集成方式
// ==========================================

// Gate 服务中发布事件示例
/*
package gate

import (
	"context"
	"encoding/json"
	"github.com/dobyte/due/eventbus/redis/v2"
	"github.com/dobyte/due/v2/cluster/gate"
)

// 初始化 EventBus
var eventBus *redis.EventBus

func InitEventBus(redisClient *redis.Client) {
	eventBus = redis.NewEventBus(
		redis.WithClient(redisClient),
		redis.WithChannel("game.events"),
	)
}

// 在连接处理中发布登录事件
func onPlayerLogin(ctx gate.Context, uid int64, username string) {
	event := UserLoginEvent{
		UID:      uid,
		Username: username,
		IP:       ctx.Session().RemoteAddr(),
		LoginAt:  time.Now(),
	}
	data, _ := json.Marshal(event)
	eventBus.Publish(context.Background(), "user.login", data)
}
*/

// Node 服务中订阅事件示例
/*
package node

import (
	"context"
	"encoding/json"
	"github.com/dobyte/due/eventbus/redis/v2"
	"github.com/dobyte/due/v2/log"
)

// 初始化 EventBus 并订阅事件
func SubscribeEvents(bus *redis.EventBus) {
	// 订阅用户登录事件，用于同步在线状态
	bus.Subscribe(context.Background(), "user.login", func(data []byte) {
		var event UserLoginEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Errorf("解析事件失败：%v", err)
			return
		}
		// 更新本地缓存的在线状态
		updateOnlineStatus(event.UID, true)
	})

	// 订阅用户登出事件
	bus.Subscribe(context.Background(), "user.logout", func(data []byte) {
		var event UserLogoutEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Errorf("解析事件失败：%v", err)
			return
		}
		// 更新本地缓存的在线状态
		updateOnlineStatus(event.UID, false)
	})
}
*/

// Mesh 微服务中订阅事件示例
/*
package mesh

import (
	"context"
	"encoding/json"
	"github.com/dobyte/due/eventbus/redis/v2"
)

// 数据分析服务订阅登录事件
func StartAnalyticsService(bus *redis.EventBus) {
	bus.Subscribe(context.Background(), "user.login", func(data []byte) {
		var event UserLoginEvent
		json.Unmarshal(data, &event)
		// 记录到数据分析系统
		recordLoginAnalytics(event)
	})
}
*/
