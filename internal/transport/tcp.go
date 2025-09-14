package transport

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hongjun500/chat-go/internal/protocol"
	"github.com/hongjun500/chat-go/pkg/logger"
)

// tcpSession TCP 会话实现
type tcpSession struct {
	*BaseSession
	conn       net.Conn
	frameCodec *FrameCodec
	protocol   *protocol.Protocol
	writeMu    sync.Mutex
	closeChan  chan struct{}
}

// newTcpSession 创建 TCP 会话
func newTcpSession(id string, conn net.Conn, codecType int) *tcpSession {
	return &tcpSession{
		BaseSession: NewBaseSession(id, conn.RemoteAddr().String()),
		conn:        conn,
		frameCodec:  NewFrameCodec(),
		protocol:    protocol.NewProtocol(codecType),
		closeChan:   make(chan struct{}),
	}
}

// SendEnvelope 发送信封消息
func (s *tcpSession) SendEnvelope(e *protocol.Envelope) error {
	if s.State() == SessionStateClosed {
		return ErrSessionClosed
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	var buff bytes.Buffer
	if err := s.protocol.GetCodec().Encode(&buff, e); err != nil {
		return err
	}

	return s.frameCodec.WriteFrame(s.conn, buff.Bytes())
}

// Close 关闭会话
func (s *tcpSession) Close() error {
	var err error
	s.closeOnce.Do(func() {
		s.markClosed()
		err = s.conn.Close()
		close(s.closeChan)
	})
	return err
}

// readLoop 读取循环（内部方法）
func (s *tcpSession) readLoop(gateway Gateway, opt Options) {
	defer func() {
		// 通知网关会话关闭
		gateway.OnSessionClose(s)
		_ = s.Close()
	}()

	for {
		// 设置读取超时
		if opt.ReadTimeout > 0 {
			_ = s.conn.SetReadDeadline(time.Now().Add(opt.ReadTimeout * time.Second))
		}
		// 读取帧数据
		frameData, err := s.frameCodec.ReadFrame(s.conn)
		if err != nil {
			if err != io.EOF && s.State() == SessionStateActive {
				logger.L().Sugar().Warnw("tcp_read_error", "session", s.ID(), "err", err)
			}
			continue
		}
		// 解码消息
		var envelope protocol.Envelope
		if err := s.protocol.GetCodec().Decode(bytes.NewReader(frameData), &envelope, opt.MaxFrameSize); err != nil {
			logger.L().Sugar().Warnw("tcp_decode_error", "session", s.ID(), "err", err)
			continue
		}

		// 传递给网关处理
		gateway.OnEnvelope(s, &envelope)
	}
}

// TCPServer TCP 服务器实现
type TCPServer struct {
	addr string
}

// NewTCPServer 创建 TCP 服务器
func NewTCPServer(addr string) *TCPServer {
	return &TCPServer{
		addr: addr,
	}
}

// Name 获取传输类型名称
func (s *TCPServer) Name() string {
	return Tcp
}

// Start 启动 TCP 服务器
func (s *TCPServer) Start(ctx context.Context, gateway Gateway, opt Options) error {
	if opt.MaxFrameSize <= 0 {
		opt.MaxFrameSize = 1 << 20 // 默认 1MB
	}

	ln, err := net.Listen(Tcp, s.addr)
	if err != nil {
		return err
	}

	logger.L().Sugar().Infow("tcp_listen", "addr", s.addr)

	// 优雅关闭
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			logger.L().Sugar().Warnw("tcp_accept_error", "err", err)
			continue
		}
		go s.handleConnection(ctx, conn, gateway, opt)
	}
}

// handleConnection 处理新连接
func (s *TCPServer) handleConnection(ctx context.Context, conn net.Conn, gateway Gateway, opt Options) {
	id := uuid.New().String()
	// 创建会话
	session := newTcpSession(id, conn, opt.TCPCodec)

	// 通知网关会话开启
	gateway.OnSessionOpen(session)
	// 使用 ctx 控制会话生命周期
	go func() {
		select {
		case <-ctx.Done():
			_ = session.Close() // 上下文取消时关闭会话
		case <-session.closeChan:
			// 会话主动关闭
			gateway.OnSessionClose(session)
		}
	}()
	// 启动读取循环
	session.readLoop(gateway, opt)
}
