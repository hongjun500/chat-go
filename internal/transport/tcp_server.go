package transport

import (
	"bytes"
	"context"
	"io"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/pkg/logger"
)

// tcpSession implements Session and holds a *chat.Client for Hub integration
type tcpSession struct {
	id     string
	conn   net.Conn
	codec  *FrameCodec
	client *chat.Client
}

func (s *tcpSession) ID() string { return s.id }
func (s *tcpSession) RemoteAddr() string {
	if s.conn != nil {
		return s.conn.RemoteAddr().String()
	}
	return ""
}
func (s *tcpSession) SendEnvelope(m *Envelope) error { return s.codec.Encode(m) }
func (s *tcpSession) Close() error                   { return s.conn.Close() }

// getClient helper expects Session to be tcpSession or similar wrapper
func getClient(sess Session) *chat.Client {
	if ts, ok := sess.(*tcpSession); ok {
		return ts.client
	}
	return nil
}

// TCPServer implements Transport using length-prefixed frames and MessageCodec on top
type TCPServer struct{ Codec MessageCodec }

func (s *TCPServer) Name() string { return "tcp" }

func (s *TCPServer) Start(ctx context.Context, addr string, gateway Gateway, opt Options) error {
	if opt.MaxFrameSize <= 0 {
		opt.MaxFrameSize = 1 << 20
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	logger.L().Sugar().Infow("tcp_listen_v2", "addr", addr)
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

func (s *TCPServer) serveConn(ctx context.Context, conn net.Conn, gateway Gateway, opt Options) {
	id := uuid.New().String()
	// framed codec for network IO
	framed := NewFrameCodec(conn)
	// chat client for Hub
	c := chat.NewClientWithBuffer(id, conn, opt.OutBuffer)
	c.Meta = map[string]string{"level": "0"}
	sess := &tcpSession{id: id, conn: conn, codec: framed, client: c}
	gateway.OnSessionOpen(sess)

	// writer: drain client outgoing to session (wrap plain text into Wire and write framed+encoded)
	go func() {
		for msg := range c.Outgoing() {
			if opt.WriteTimeout > 0 {
				_ = conn.SetWriteDeadline(time.Now().Add(opt.WriteTimeout))
			}
			// encode payload using codec then write frame
			var buf bytes.Buffer
			if err := s.Codec.Encode(&buf, &Envelope{Type: "text", Text: msg, Ts: time.Now().UnixMilli()}); err != nil {
				logger.L().Sugar().Warnw("tcp_v2_write_error", "client", c.ID, "err", err)
				_ = conn.Close()
				return
			}
			if err := framed.WriteFrame(buf.Bytes()); err != nil {
				logger.L().Sugar().Warnw("tcp_v2_write_error", "client", c.ID, "err", err)
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
				logger.L().Sugar().Warnw("tcp_v2_decode_error", "client", id, "err", err)
			}
			gateway.OnSessionClose(sess, err)
			return
		}
		var env Envelope
		if err := s.Codec.Decode(bytes.NewReader(raw), &env, opt.MaxFrameSize); err != nil {
			logger.L().Sugar().Warnw("tcp_v2_codec_error", "client", id, "err", err)
			continue
		}
		gateway.OnEnvelope(sess, &env)
	}
}
