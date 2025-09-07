package transport

import (
	"bytes"
	"context"
	"github.com/hongjun500/chat-go/internal/protocol"
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
	id             string
	conn           net.Conn
	codec          protocol.MessageCodec
	frameCodec     *FrameCodec // new structured approach
	client         *chat.Client
	closeOnce      sync.Once
	closeChan      chan struct{}
	payloadEncoder *protocol.PayloadEncoder
}

// TCPServer implements Transport using length-prefixed frames and MessageCodec on top
type TCPServer struct {
	Codec          protocol.MessageCodec
	PayloadEncoder *protocol.PayloadEncoder
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

/*
	func (fmc *FramedMessageCodec) Encode(msg *protocol.Envelope) error {
		var buf bytes.Buffer
		if err := fmc.messageCodec.Encode(&buf, msg); err != nil {
			return err
		}
		return fmc.frameCodec.WriteFrame(buf.Bytes())
	}

// Decode reads a frame and decodes it using the message codec

	func (fmc *FramedMessageCodec) Decode(msg *protocol.Envelope, maxSize int) error {
		frameData, err := fmc.frameCodec.ReadFrame(maxSize)
		if err != nil {
			return err
		}
		return fmc.messageCodec.Decode(bytes.NewReader(frameData), msg, maxSize)
	}
*/
func (t *tcpConn) SendEnvelope(m *protocol.Envelope) error {
	var buf bytes.Buffer
	return t.codec.Encode(&buf, m)
}
func (t *tcpConn) Close() error {
	var err error
	t.closeOnce.Do(func() {
		err = t.conn.Close()
		close(t.closeChan)
	})
	return err
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
	// Create frame codec for low-level framing
	framed := NewFrameCodec(conn)

	// Initialize payload encoder if not set
	payloadEncoder := s.PayloadEncoder
	if payloadEncoder == nil {
		payloadEncoder = protocol.DefaultPayloadEncoder
	}

	// chat client for Hub
	c := chat.NewClientWithBuffer(id, opt.OutBuffer)
	c.Meta = map[string]string{"level": "0"}
	sess := &tcpConn{
		id:             id,
		conn:           conn,
		client:         c,
		frameCodec:     framed,
		payloadEncoder: payloadEncoder,
	}
	gateway.OnSessionOpen(sess)

	// writer: drain client outgoing to session (wrap plain text into Envelope with typed payload)
	go func() {
		for msg := range c.Outgoing() {
			if opt.WriteTimeout > 0 {
				_ = conn.SetWriteDeadline(time.Now().Add(opt.WriteTimeout))
			}
			// Use the payload encoder to create structured envelope
			envelope, _ := sess.payloadEncoder.EncodeText(msg)
			envelope.Ts = time.Now().UnixMilli()
			// just to ensure it's valid
			if err := s.Codec.Encode(&bytes.Buffer{}, envelope); err != nil {
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
		// Use the new structured approach - decode frame and message in one step
		var env protocol.Envelope
		if err := s.Codec.Decode(&bytes.Buffer{}, &env, opt.MaxFrameSize); err != nil {
			if err != io.EOF {
				logger.L().Sugar().Warnw("tcp_decode_error", "client", id, "err", err)
			}
			gateway.OnSessionClose(sess, err)
			return
		}
		gateway.OnEnvelope(sess, &env)
	}
}
