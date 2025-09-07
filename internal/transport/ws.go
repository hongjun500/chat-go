package transport

import (
	"bytes"
	"context"
	"github.com/hongjun500/chat-go/internal/protocol"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/pkg/logger"
)

// wsConn implements Session for WebSocket connections
type wsConn struct {
	id             string
	conn           *websocket.Conn
	codec          protocol.MessageCodec
	client         *chat.Client
	closeOnce      sync.Once
	closeChan      chan struct{}
	payloadEncoder *protocol.PayloadEncoder
}

// WebSocketServer implements Transport using WebSocket connections
type WebSocketServer struct {
	Codec protocol.MessageCodec
	Path  string // WebSocket endpoint path, defaults to "/ws"
}

func (w *wsConn) ID() string {
	return w.id
}

func (w *wsConn) RemoteAddr() string {
	return w.conn.RemoteAddr().String()
}

func (w *wsConn) SendEnvelope(m *protocol.Envelope) error {
	var buffer bytes.Buffer
	if err := w.codec.Encode(&buffer, m); err != nil {
		return err
	}
	return w.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
}

func (w *wsConn) Close() error {
	var err error
	w.closeOnce.Do(func() {
		err = w.conn.Close()
		close(w.closeChan)
	})
	return err
}

func (ws *WebSocketServer) Name() string {
	return WebSocket
}

func (ws *WebSocketServer) Start(ctx context.Context, addr string, gateway Gateway, opt Options) error {
	if ws.Codec == nil {
		ws.Codec = &protocol.JSONCodec{} // default to JSON
	}
	if ws.Path == "" {
		ws.Path = "/ws"
	}
	mux := http.NewServeMux()
	mux.HandleFunc(ws.Path, func(w http.ResponseWriter, r *http.Request) {
		ws.handleConnection(w, r, gateway, opt)
	})

	logger.L().Sugar().Infow("websocket_listen", "addr", addr, "path", ws.Path)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	return server.ListenAndServe()
}

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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	client := chat.NewClientWithBuffer(id, opt.OutBuffer)
	client.Meta = map[string]string{"level": "0"}

	sess := &wsConn{
		id:        id,
		conn:      conn,
		codec:     ws.Codec,
		client:    client,
		closeChan: make(chan struct{}),
	}

	gateway.OnSessionOpen(sess)

	// Writer goroutine: send outgoing messages from client as Envelope
	go func() {
		defer func() {
			gateway.OnSessionClose(sess, nil)
			_ = sess.Close()
		}()

		for msg := range client.Outgoing() {
			if opt.WriteTimeout > 0 {
				_ = conn.SetWriteDeadline(time.Now().Add(opt.WriteTimeout))
			}

			// Convert plain text message to Envelope using payload encoder
			envelope, _ := sess.payloadEncoder.EncodeText(msg)
			envelope.Ts = time.Now().UnixMilli()

			if err := sess.SendEnvelope(envelope); err != nil {
				logger.L().Sugar().Warnw("ws_write_error", "client", client.ID, "err", err)
				return
			}
		}
	}()

	// Setup heartbeat
	if opt.ReadTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(opt.ReadTimeout))
	} else {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}

	conn.SetPongHandler(func(string) error {
		deadline := time.Now().Add(60 * time.Second)
		if opt.ReadTimeout > 0 {
			deadline = time.Now().Add(opt.ReadTimeout)
		}
		return conn.SetReadDeadline(deadline)
	})

	// Periodic ping
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second))
			case <-sess.closeChan:
				return
			}
		}
	}()

	// Reader loop: read messages and convert to Envelope
	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			gateway.OnSessionClose(sess, err)
			return
		}

		if mt != websocket.TextMessage {
			continue
		}

		// Try to parse as Envelope first, fallback to legacy text message
		var envelope protocol.Envelope
		if err := ws.Codec.Decode(bytes.NewBuffer(data), &envelope, opt.MaxFrameSize); err == nil && envelope.Type != "" {
			gateway.OnEnvelope(sess, &envelope)
		} else {
			text := string(data)
			if text == "" {
				continue
			}
			ws.handleLegacyMessage(sess, text, gateway)
		}
	}
}

// handleLegacyMessage processes plain text WebSocket messages for backward compatibility
func (ws *WebSocketServer) handleLegacyMessage(sess *wsConn, text string, gateway Gateway) {
	client := sess.client

	// If no name set, treat as set_name
	if client.Name == "" {
		envelope, _ := sess.payloadEncoder.EncodeSetName(text)
		envelope.Ts = time.Now().UnixMilli()
		gateway.OnEnvelope(sess, envelope)
		return
	}

	// If starts with /, treat as command
	if len(text) > 0 && text[0] == '/' {
		envelope, _ := sess.payloadEncoder.EncodeCommand(text)
		envelope.Ts = time.Now().UnixMilli()
		gateway.OnEnvelope(sess, envelope)
		return
	}

	// Otherwise treat as chat message
	envelope, _ := sess.payloadEncoder.EncodeChat(text)
	envelope.From = client.Name
	envelope.Ts = time.Now().UnixMilli()
	gateway.OnEnvelope(sess, envelope)
}
