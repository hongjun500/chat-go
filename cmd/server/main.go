package main

import (
	"context"
	"time"

	"github.com/hongjun500/chat-go/internal/bus/redisstream"
	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/command"
	"github.com/hongjun500/chat-go/internal/config"
	"github.com/hongjun500/chat-go/internal/observe"
	"github.com/hongjun500/chat-go/internal/subscriber"
	"github.com/hongjun500/chat-go/internal/transport"
	"github.com/hongjun500/chat-go/pkg/logger"
)

func main() {
	cfg := config.Load()
	logger.SetLevel(cfg.LogLevel)
	hub := chat.NewHub()
	// 初始化命令注册表（解环：在 main 中创建并传递）
	cmdReg := command.NewRegistry()
	if err := command.RegisterBuiltins(cmdReg); err != nil {
		panic(err)
	}
	// 注册标准订阅者合集
	subscriber.RegisterAll(hub)

	// 并发启动 TCP/WS/HTTP（静态页 ws.html 用于 WebSocket 测试）
	go func() {
		_ = transport.StartTcpWithOptions(cfg.TCPAddr, hub, cmdReg, transport.Options{OutBuffer: cfg.OutBuffer})
	}()
	go func() {
		_ = transport.StartWSWithOptions(cfg.WSAddr, hub, cmdReg, transport.Options{OutBuffer: cfg.OutBuffer})
	}()
	go func() {
		_ = observe.StartHTTP(cfg.HTTPAddr)
	}()

	// 可选：Redis Stream 分布式同步
	if cfg.RedisEnable && cfg.RedisAddr != "" {
		go func() {
			bus := redisstream.New(cfg.RedisAddr, cfg.RedisDB, cfg.RedisStream, cfg.RedisGroup)
			_ = bus.EnsureGroup(context.Background())

			// 发布本地事件：仅针对 chat 消息类（可扩展系统通知/文件等）
			hub.Subscribe(chat.EventMessageLocal, func(e chat.Event) {
				me := e.(*chat.MessageEvent)
				_ = bus.Publish(context.Background(), &redisstream.Message{Type: "message", When: me.When, From: me.From, Text: me.Content})
			})
			hub.Subscribe(chat.EventMessageDirect, func(e chat.Event) {
				de := e.(*chat.DirectMessageEvent)
				_ = bus.Publish(context.Background(), &redisstream.Message{Type: "direct", When: de.When, From: de.From, To: de.To, Text: de.Content})
			})

			// 消费远端事件 -> 转为本地 Remote 事件
			_ = bus.Consume(context.Background(), "consumer-"+time.Now().Format("150405"), func(ctx context.Context, m *redisstream.Message) error {
				switch m.Type {
				case "message":
					hub.BroadcastRemote(m.From, m.Text, m.When)
				case "direct":
					hub.Emit(&chat.DirectMessageEvent{When: m.When, From: m.From, To: m.To, Content: m.Text})
				}
				return nil
			})
		}()
	}
	select {}
}
