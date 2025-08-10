package transport

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/command"
	"github.com/hongjun500/chat-go/pkg/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// StartWSWithRegistry 使用默认选项启动 WebSocket 服务
func StartWSWithRegistry(addr string, hub *chat.Hub, reg *command.Registry) error {
	return StartWSWithOptions(addr, hub, reg, Options{OutBuffer: 256})
}

// StartWSWithOptions 启动 WebSocket 服务，支持缓冲区配置
func StartWSWithOptions(addr string, hub *chat.Hub, reg *command.Registry, opt Options) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, hub, reg, opt)
	})
	logger.L().Sugar().Infow("ws_listen", "addr", addr, "path", "/ws")
	return http.ListenAndServe(addr, mux)
}

func wsHandler(w http.ResponseWriter, r *http.Request, hub *chat.Hub, reg *command.Registry, opt Options) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := uuid.New().String()
	client := chat.NewClientWithBuffer(id, nil, opt.OutBuffer)

	// writer: flush outgoing queue to websocket
	go func() {
		for msg := range client.Outgoing() {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				logger.L().Sugar().Warnw("ws_write_error", "client", client.ID, "err", err)
				hub.UnregisterClient(client)
				return
			}
		}
		// ensure ws closed when channel drained
		_ = conn.Close()
	}()

	// initial greeting
	_ = conn.WriteMessage(websocket.TextMessage, []byte("请输入昵称并回车："))

	// 心跳设置：读超时 + pong handler
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})
	// 定期发送 ping
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			_ = conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second))
		}
	}()

	nameSet := false
	var cancelDirect func()
	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			hub.UnregisterClient(client)
			if cancelDirect != nil {
				cancelDirect()
			}
			return
		}
		if mt != websocket.TextMessage {
			continue
		}
		line := strings.TrimSpace(string(data))
		if line == "" {
			continue
		}
		if !nameSet {
			// 封禁校验
			if hub.IsBanned(line) {
				_ = conn.WriteMessage(websocket.TextMessage, []byte("该用户已被封禁"))
				hub.UnregisterClient(client)
				return
			}
			client.Name = line
			nameSet = true
			hub.RegisterClient(client)
			client.Send("昵称设置成功：" + line)
			// 订阅点对点消息：仅当 To == 自己昵称时推送
			cancelDirect = hub.SubscribeCancelable(chat.EventMessageDirect, func(e chat.Event) {
				de := e.(*chat.DirectMessageEvent)
				if de.To == client.Name {
					client.Send("[私信] " + de.From + ": " + de.Content)
				}
			})
			continue
		}
		if strings.HasPrefix(line, "/") {
			handled, err := reg.Execute(line, &command.Context{Hub: hub, Client: client, Raw: line})
			if handled {
				if err != nil {
					client.Send("命令错误: " + err.Error())
				}
				continue
			}
		}
		hub.BroadcastLocal(client.Name, line)
		time.Sleep(1 * time.Millisecond)
	}
}
