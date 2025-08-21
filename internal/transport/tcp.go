package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/pkg/logger"
)

// tcpConn implements Session and holds a *chat.Client for Hub integration
type tcpConn struct {
	id        string
	conn      net.Conn
	codec     *FrameCodec
	client    *chat.Client
	closeOnce sync.Once
	closeChan chan struct{}
}

// TCPServer implements Transport using length-prefixed frames and MessageCodec on top
type TCPServer struct {
	Codec MessageCodec
}

func (t *tcpConn) ID() string {
	return t.id
}
func (t *tcpConn) RemoteAddr() string {
	if t.conn != nil {
		return t.conn.RemoteAddr().String()
	}
	return ""
}
func (t *tcpConn) SendEnvelope(m *Envelope) error {
	return t.codec.Encode(m)
}
func (t *tcpConn) Close() error {
	var err error
	t.closeOnce.Do(func() {
		err = t.conn.Close()
		close(t.closeChan)
	})
	return err
}

// getClient helper expects Session to be tcpConn or similar wrapper
func getClient(sess Session) *chat.Client {
	if ts, ok := sess.(*tcpConn); ok {
		return ts.client
	}
	return nil
}

func (s *TCPServer) Start(ctx context.Context, addr string, gateway Gateway, opt Options) error {
	if opt.MaxFrameSize <= 0 {
		opt.MaxFrameSize = 1 << 20
	}
	ln, err := net.Listen(Tcp, addr)
	if err != nil {
		return err
	}
	logger.L().Sugar().Infow("tcp_listen", "addr", addr)
	go func() { <-ctx.Done(); _ = ln.Close() }()
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			logger.L().Sugar().Warnw("tcp_accept_error", "err", err)
			continue
		}
		go s.serveConn(ctx, conn, gateway, opt)
	}
}

func (s *TCPServer) Name() string {
	return Tcp
}

func (s *TCPServer) serveConn(ctx context.Context, conn net.Conn, gateway Gateway, opt Options) {
	id := uuid.New().String()
	// framed codec for network IO
	framed := NewFrameCodec(conn)
	// chat client for Hub
	c := chat.NewClientWithBuffer(id, opt.OutBuffer)
	c.Meta = map[string]string{"level": "0"}
	sess := &tcpConn{id: id, conn: conn, codec: framed, client: c}
	gateway.OnSessionOpen(sess)

	// writer: drain client outgoing to session (wrap plain text into Envelope with typed payload)
	go func() {
		for msg := range c.Outgoing() {
			if opt.WriteTimeout > 0 {
				_ = conn.SetWriteDeadline(time.Now().Add(opt.WriteTimeout))
			}
			// encode payload using codec then write frame
			var buf bytes.Buffer
			payload, _ := json.Marshal(TextPayload{Text: msg})
			if err := s.Codec.Encode(&buf, &Envelope{Type: "text", Payload: payload, Ts: time.Now().UnixMilli()}); err != nil {
				logger.L().Sugar().Warnw("tcp_write_error", "client", c.ID, "err", err)
				_ = conn.Close()
				return
			}
			if err := framed.WriteFrame(buf.Bytes()); err != nil {
				logger.L().Sugar().Warnw("tcp_write_error", "client", c.ID, "err", err)
				_ = conn.Close()
				return
			}
		}
		_ = conn.Close()
	}()

	// reader loop
	for {
		if opt.ReadTimeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(opt.ReadTimeout))
		}
		// Read frame bytes then decode with configured codec
		raw, err := framed.ReadFrame(opt.MaxFrameSize)
		if err != nil {
			if err != io.EOF {
				logger.L().Sugar().Warnw("tcp_decode_error", "client", id, "err", err)
			}
			gateway.OnSessionClose(sess, err)
			return
		}
		var env Envelope
		if err := s.Codec.Decode(bytes.NewReader(raw), &env, opt.MaxFrameSize); err != nil {
			logger.L().Sugar().Warnw("tcp_codec_error", "client", id, "err", err)
			continue
		}
		gateway.OnEnvelope(sess, &env)
	}
}
