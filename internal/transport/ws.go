package transport

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/hongjun500/chat-go/internal/protocol"
	"github.com/hongjun500/chat-go/pkg/logger"
)

// wsSession WebSocket 会话实现
type wsSession struct {
	*BaseSession
	conn            *websocket.Conn
	protocolManager *protocol.ProtocolManager
	writeMu         sync.Mutex
	closeChan       chan struct{}
}

// newWsSession 创建 WebSocket 会话
func newWsSession(id string, conn *websocket.Conn, protocolManager *protocol.ProtocolManager) *wsSession {
	return &wsSession{
		BaseSession:     NewBaseSession(id, conn.RemoteAddr().String()),
		conn:            conn,
		protocolManager: protocolManager,
		closeChan:       make(chan struct{}),
	}
}

// SendEnvelope 发送信封消息
func (s *wsSession) SendEnvelope(envelope *protocol.Envelope) error {
	if s.State() == SessionStateClosed {
		return ErrSessionClosed
	}

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	var buffer bytes.Buffer
	if err := s.protocolManager.EncodeMessage(&buffer, envelope); err != nil {
		return err
	}

	return s.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
}

// Close 关闭会话
func (s *wsSession) Close() error {
	var err error
	s.closeOnce.Do(func() {
		s.markClosed()
		err = s.conn.Close()
		close(s.closeChan)
	})
	return err
}

// WebSocketServer WebSocket 服务器实现
type WebSocketServer struct {
	Path string // WebSocket endpoint path, defaults to "/ws"
}

// NewWebSocketServer 创建 WebSocket 服务器
func NewWebSocketServer(path string) *WebSocketServer {
	if path == "" {
		path = "/ws"
	}
	return &WebSocketServer{
		Path: path,
	}
}

// Name 获取传输类型名称
func (ws *WebSocketServer) Name() string {
	return WebSocket
}

// Start 启动 WebSocket 服务器
func (ws *WebSocketServer) Start(ctx context.Context, addr string, gateway Gateway, opt Options) error {
	mux := http.NewServeMux()
	mux.HandleFunc(ws.Path, func(w http.ResponseWriter, r *http.Request) {
		ws.handleConnection(w, r, gateway, opt)
	})

	logger.L().Sugar().Infow("websocket_listen", "addr", addr, "path", ws.Path)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// 优雅关闭
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	return server.ListenAndServe()
}

// handleConnection 处理新的 WebSocket 连接
func (ws *WebSocketServer) handleConnection(w http.ResponseWriter, r *http.Request, gateway Gateway, opt Options) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.L().Sugar().Warnw("websocket_upgrade_error", "err", err)
		return
	}

	id := uuid.New().String()
	session := newWsSession(id, conn, opt.GetWSProtocolManager())

	// 通知网关会话开启
	gateway.OnSessionOpen(session)

	// 设置心跳
	ws.setupHeartbeat(session, opt)

	// 启动读取循环
	go ws.readLoop(session, gateway, opt)
}

// setupHeartbeat 设置心跳机制
func (ws *WebSocketServer) setupHeartbeat(session *wsSession, opt Options) {
	// 设置读取超时
	timeout := 60 * time.Second
	if opt.ReadTimeout > 0 {
		timeout = opt.ReadTimeout
	}

	_ = session.conn.SetReadDeadline(time.Now().Add(timeout))

	// 设置pong处理器
	session.conn.SetPongHandler(func(string) error {
		return session.conn.SetReadDeadline(time.Now().Add(timeout))
	})

	// 定期发送ping
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = session.conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second))
			case <-session.closeChan:
				return
			}
		}
	}()
}

// readLoop 读取循环
func (ws *WebSocketServer) readLoop(session *wsSession, gateway Gateway, opt Options) {
	defer func() {
		gateway.OnSessionClose(session)
		_ = session.Close()
	}()

	for {
		mt, data, err := session.conn.ReadMessage()
		if err != nil {
			if session.State() == SessionStateActive {
				logger.L().Sugar().Warnw("websocket_read_error", "session", session.ID(), "err", err)
			}
			return
		}

		if mt != websocket.TextMessage {
			continue
		}

		// 尝试解析为 Envelope
		var envelope protocol.Envelope
		if err := session.protocolManager.DecodeMessage(bytes.NewBuffer(data), &envelope, opt.MaxFrameSize); err == nil && envelope.Type != "" {
			gateway.OnEnvelope(session, &envelope)
		} else {
			// 回退处理纯文本消息（向后兼容）
			ws.handleLegacyTextMessage(session, string(data), gateway)
		}
	}
}

// handleLegacyTextMessage 处理纯文本消息（向后兼容）
func (ws *WebSocketServer) handleLegacyTextMessage(session *wsSession, text string, gateway Gateway) {
	if text == "" {
		return
	}

	factory := session.protocolManager.GetMessageFactory()
	
	// 简单的文本消息处理
	if len(text) > 0 && text[0] == '/' {
		// 命令消息
		envelope := factory.CreateCommandMessage(text)
		gateway.OnEnvelope(session, envelope)
	} else {
		// 普通文本消息
		envelope := factory.CreateTextMessage(text)
		gateway.OnEnvelope(session, envelope)
	}
}
