package transport

import (
	"bytes"
	"context"
	"github.com/google/uuid"
	"github.com/hongjun500/chat-go/internal/protocol"
	"github.com/hongjun500/chat-go/pkg/logger"
	"io"
	"net"
	"sync"
	"time"
)

// tcpSession TCP 会话实现
type tcpSession struct {
	*BaseSession
	conn      net.Conn
	codec     *FrameCodec
	protocol  *protocol.Protocol
	writeMu   sync.Mutex
	closeChan chan struct{}
}

// NewTCPSession 创建 TCP 会话
func NewTCPSession(id string, conn net.Conn, codecType int) *tcpSession {
	return &tcpSession{
		BaseSession: NewBaseSession(id, conn.RemoteAddr().String()),
		conn:        conn,
		codec:       NewFrameCodec(),
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
	
	return s.codec.WriteFrame(s.conn, buff.Bytes())
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
		gateway.OnSessionClose(s, nil)
		_ = s.Close()
	}()
	
	for {
		// 设置读取超时
		if opt.ReadTimeout > 0 {
			_ = s.conn.SetReadDeadline(time.Now().Add(time.Duration(opt.ReadTimeout) * time.Second))
		}
		
		// 读取帧数据
		frameData, err := s.codec.ReadFrame(s.conn)
		if err != nil {
			if err != io.EOF && s.State() == SessionStateActive {
				logger.L().Sugar().Warnw("tcp_read_error", "session", s.ID(), "err", err)
			}
			return
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
	sessionManager *SessionManager
}

// NewTCPServer 创建 TCP 服务器
func NewTCPServer() *TCPServer {
	return &TCPServer{
		sessionManager: NewSessionManager(),
	}
}

// Name 获取传输类型名称
func (s *TCPServer) Name() string {
	return Tcp
}

// Start 启动 TCP 服务器
func (s *TCPServer) Start(ctx context.Context, addr string, gateway Gateway, opt Options) error {
	if opt.MaxFrameSize <= 0 {
		opt.MaxFrameSize = 1 << 20 // 默认 1MB
	}
	
	ln, err := net.Listen(Tcp, addr)
	if err != nil {
		return err
	}
	
	logger.L().Sugar().Infow("tcp_listen", "addr", addr)
	
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
	session := NewTCPSession(id, conn, opt.TCPCodec)
	
	// 添加到会话管理器
	s.sessionManager.AddSession(session)
	
	// 清理处理
	session.AddCloseHandler(func() {
		s.sessionManager.RemoveSession(id)
	})
	
	// 通知网关会话开启
	gateway.OnSessionOpen(session)
	
	// 启动读取循环
	session.readLoop(gateway, opt)
}
