package transport

import (
	"sync"
	"sync/atomic"

	"github.com/hongjun500/chat-go/internal/protocol"
)

// SessionState 会话状态
type SessionState int

const (
	SessionStateActive SessionState = iota
	SessionStateClosed
)

const (
	SessionContextUnClosed = iota
	SessionContextClosed
)

// Session 传输层统一的会话管理接口
// 负责底层连接的生命周期管理和数据传输
type Session interface {
	ID() string
	RemoteAddr() string
	SendEnvelope(*protocol.Envelope) error // 发送消息到客户端
	Close() error
}

// Base 基础会话实现，提供通用功能
type Base struct {
	id         string
	remoteAddr string
	state      SessionState
	stateMu    sync.RWMutex
	closeOnce  sync.Once
}

type SessionContext struct {
	Id         string
	RemoteAddr string
	sess       Session

	closed    int32
	closeOnce sync.Once
}

// NewBase 提供会话的基础信息
func NewBase(id, remoteAddr string) *Base {
	return &Base{
		id:         id,
		remoteAddr: remoteAddr,
		state:      SessionStateActive,
	}
}

// ID 获取会话ID
func (s *Base) ID() string {
	return s.id
}

// RemoteAddr 获取远程地址
func (s *Base) RemoteAddr() string {
	return s.remoteAddr
}

// State 获取会话状态
func (s *Base) State() SessionState {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return s.state
}

func NewSessionContext(s Session) *SessionContext {
	return &SessionContext{Id: s.ID(), RemoteAddr: s.RemoteAddr(), sess: s}
}

func (sc *SessionContext) Send(e *protocol.Envelope) error {
	if atomic.LoadInt32(&sc.closed) == SessionContextClosed {
		return ErrSessionContextClosed
	}
	return sc.sess.SendEnvelope(e)
}

func (sc *SessionContext) Close() error {
	var err error
	sc.closeOnce.Do(func() {
		atomic.StoreInt32(&sc.closed, SessionContextClosed)
		err = sc.sess.Close()
	})
	return err
}
