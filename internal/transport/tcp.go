package transport

import (
	"bytes"
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/hongjun500/chat-go/internal/protocol"
	"github.com/hongjun500/chat-go/pkg/logger"
)

// tcpSession TCP 会话实现
type tcpSession struct {
	*Base
	conn            net.Conn
	frameCodec      *FrameCodec
	protocolManager *protocol.Manager
	writeMu         sync.Mutex
	closeChan       chan struct{}
	lastActive      atomic.Value // 上次活动时间
}

// newTcpSession 创建 TCP 会话
func newTcpSession(id string, conn net.Conn, protocolManager *protocol.Manager) *tcpSession {
	se := &tcpSession{
		Base:            NewBase(id, conn.RemoteAddr().String()),
		conn:            conn,
		frameCodec:      NewFrameCodec(),
		protocolManager: protocolManager,
		closeChan:       make(chan struct{}),
	}
	se.lastActive.Store(time.Now())
	return se
}

// SendEnvelope 发送信封消息
func (s *tcpSession) SendEnvelope(e *protocol.Envelope) error {
	if s.State() == SessionStateClosed {
		return ErrSessionClosed
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	var buff bytes.Buffer
	if err := s.protocolManager.EncodeMessage(&buff, e); err != nil {
		return err
	}

	return s.frameCodec.WriteFrame(s.conn, buff.Bytes())
}

// Close 关闭会话
func (s *tcpSession) Close() error {
	var err error
	s.closeOnce.Do(func() {
		err = s.conn.Close()
		close(s.closeChan)
	})
	return err
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
func (s *TCPServer) Start(ctx context.Context, addr string, gateway Gateway, opt Options) error {
	// 如果提供了addr参数，使用它覆盖构造时的地址
	listenAddr := s.addr
	if addr != "" {
		listenAddr = addr
	}

	if opt.MaxFrameSize <= 0 {
		opt.MaxFrameSize = 1 << 20 // 默认 1MB
	}

	ln, err := net.Listen(Tcp, listenAddr)
	if err != nil {
		return err
	}

	logger.L().Sugar().Infow("tcp_listen", "addr", listenAddr)

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
	// 创建会话（使用协议管理器）
	session := newTcpSession(id, conn, opt.GetTCPProtocolManager())
	// 创建会话上下文
	sc := NewSessionContext(session)
	// 通知网关会话开启
	gateway.OnSessionOpen(sc)
	// 会话生命周期监控
	go session.lifecycleWatcher(ctx, gateway, sc)
	// 心跳监控
	go session.heartbeatWatcher(opt)
	// 启动读取循环
	session.readLoop(gateway, sc, opt)
}

// lifecycleWatcher 会话生命周期监控
func (s *tcpSession) lifecycleWatcher(ctx context.Context, gateway Gateway, sc *SessionContext) {
	select {
	case <-ctx.Done():
		_ = s.Close()
	case <-s.closeChan:
		// 会话主动关闭
		gateway.OnSessionClose(sc)
	}
}

// heartbeatWatcher 心跳监控
func (s *tcpSession) heartbeatWatcher(opt Options) {
	if opt.HeartbeatInterval <= 0 || opt.HeartbeatTimeout <= 0 {
		return
	}
	ticker := time.NewTicker(opt.HeartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		last := s.lastActive.Load().(time.Time)
		// 超过心跳超时，关闭会话
		if time.Since(last) > opt.HeartbeatTimeout {
			logger.L().Sugar().Infow("tcp_heartbeat_timeout", "session", s.ID())
			_ = s.Close()
			return
		}
	}
}

// readLoop 读取循环（内部方法）
func (s *tcpSession) readLoop(gateway Gateway, sessionContext *SessionContext, opt Options) {

	defer func() {
		// 通知网关会话关闭
		gateway.OnSessionClose(sessionContext)
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
			if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			if s.State() == SessionStateActive {
				logger.L().Sugar().Warnw("tcp_read_error", "session", s.ID(), "err", err)
			}
			return // 读取错误，退出循环
		}
		// 解码消息
		var envelope protocol.Envelope
		if err := s.protocolManager.DecodeMessage(bytes.NewReader(frameData), &envelope, opt.MaxFrameSize); err != nil {
			logger.L().Sugar().Warnw("tcp_decode_error", "session", s.ID(), "err", err)
			continue
		}
		// 更新最后活动时间
		s.lastActive.Store(time.Now())
		if envelope.Type == protocol.MsgHeartbeat {
			continue // 忽略心跳消息不传递给网关
		}
		// 传递给网关处理
		gateway.OnEnvelope(sessionContext, &envelope)
	}
}
