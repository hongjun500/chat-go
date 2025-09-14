package main

import (
	"context"
	"github.com/hongjun500/chat-go/internal/protocol"
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

	// 创建协议管理器替代直接的编解码器
	tcpProtocolManager := protocol.NewProtocolManager(cfg.TCPCodec)
	wsProtocolManager := protocol.NewProtocolManager(cfg.WSCodec)

	logger.L().Sugar().Infow("protocol_configuration",
		"tcp_codec", cfg.TCPCodec,
		"ws_codec", cfg.WSCodec,
		"tcp_codec_name", tcpProtocolManager.GetCodec().Name(),
		"ws_codec_name", wsProtocolManager.GetCodec().Name())

	// 并发启动 TCP/WS/HTTP（静态页 ws.html 用于 WebSocket 测试）
	// 新抽象：使用协议无关的 Gateway + 统一的Transport接口
	go func() {
		tcpSrv := transport.NewTCPServer(cfg.TCPAddr)
		gw := transport.NewSimpleGateway(tcpProtocolManager)
		setupBasicMessageHandlers(gw, hub, cmdReg)
		logger.L().Sugar().Infow("starting_tcp_server", "addr", cfg.TCPAddr, "codec", cfg.TCPCodec)
		_ = tcpSrv.Start(context.Background(), cfg.TCPAddr, gw, transport.Options{
			OutBuffer:            cfg.OutBuffer,
			ReadTimeout:          time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout:         time.Duration(cfg.WriteTimeout) * time.Second,
			MaxFrameSize:         cfg.MaxFrameSize,
			TCPProtocolManager:   tcpProtocolManager,
		})
	}()
	go func() {
		wsSrv := transport.NewWebSocketServer("/ws")
		gw := transport.NewSimpleGateway(wsProtocolManager)
		setupBasicMessageHandlers(gw, hub, cmdReg)
		logger.L().Sugar().Infow("starting_ws_server", "addr", cfg.WSAddr, "codec", cfg.WSCodec)
		_ = wsSrv.Start(context.Background(), cfg.WSAddr, gw, transport.Options{
			OutBuffer:           cfg.OutBuffer,
			ReadTimeout:         time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout:        time.Duration(cfg.WriteTimeout) * time.Second,
			WSProtocolManager:   wsProtocolManager,
		})
	}()
	go func() {
		logger.L().Sugar().Infow("starting_http_server", "addr", cfg.HTTPAddr)
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

// setupBasicMessageHandlers 设置基本的消息处理器
func setupBasicMessageHandlers(gw *transport.SimpleGateway, hub *chat.Hub, cmdReg *command.Registry) {
	protocolManager := gw.GetProtocolManager()
	
	// 注册文本消息处理器
	protocolManager.RegisterMessageHandler(protocol.MsgText, func(env *protocol.Envelope) error {
		logger.L().Sugar().Infow("received_text_message", "from", env.From, "data", string(env.Data))
		// 这里可以集成到 hub 进行消息广播
		return nil
	})
	
	// 注册命令消息处理器
	protocolManager.RegisterMessageHandler(protocol.MsgCommand, func(env *protocol.Envelope) error {
		logger.L().Sugar().Infow("received_command", "from", env.From, "data", string(env.Data))
		// 这里可以集成到命令注册表
		return nil
	})
	
	// 注册确认消息处理器
	protocolManager.RegisterMessageHandler(protocol.MsgAck, func(env *protocol.Envelope) error {
		logger.L().Sugar().Infow("received_ack", "status", string(env.Data), "correlation", env.Correlation)
		return nil
	})
}
