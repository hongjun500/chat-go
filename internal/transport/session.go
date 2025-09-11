package transport

import (
	"github.com/hongjun500/chat-go/internal/protocol"
	"sync"
)

// SessionState 会话状态
type SessionState int

const (
	SessionStateActive SessionState = iota
	SessionStateClosed
)

// BaseSession 基础会话实现，提供通用功能
type BaseSession struct {
	id           string
	remoteAddr   string
	state        SessionState
	stateMu      sync.RWMutex
	metadata     map[string]string
	metaMu       sync.RWMutex
	closeHandlers []func()
	closeOnce    sync.Once
}

// NewBaseSession 创建基础会话
func NewBaseSession(id, remoteAddr string) *BaseSession {
	return &BaseSession{
		id:         id,
		remoteAddr: remoteAddr,
		state:      SessionStateActive,
		metadata:   make(map[string]string),
	}
}

// ID 获取会话ID
func (s *BaseSession) ID() string {
	return s.id
}

// RemoteAddr 获取远程地址
func (s *BaseSession) RemoteAddr() string {
	return s.remoteAddr
}

// State 获取会话状态
func (s *BaseSession) State() SessionState {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return s.state
}

// SetMetadata 设置元数据
func (s *BaseSession) SetMetadata(key, value string) {
	s.metaMu.Lock()
	defer s.metaMu.Unlock()
	s.metadata[key] = value
}

// GetMetadata 获取元数据
func (s *BaseSession) GetMetadata(key string) (string, bool) {
	s.metaMu.RLock()
	defer s.metaMu.RUnlock()
	value, exists := s.metadata[key]
	return value, exists
}

// AddCloseHandler 添加关闭处理器
func (s *BaseSession) AddCloseHandler(handler func()) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.state == SessionStateClosed {
		// 如果已关闭，立即执行
		go handler()
		return
	}
	s.closeHandlers = append(s.closeHandlers, handler)
}

// markClosed 标记会话为已关闭状态（内部方法）
func (s *BaseSession) markClosed() {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	
	if s.state == SessionStateClosed {
		return
	}
	
	s.state = SessionStateClosed
	
	// 执行关闭处理器
	for _, handler := range s.closeHandlers {
		go handler()
	}
	s.closeHandlers = nil
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]Session
	mu       sync.RWMutex
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]Session),
	}
}

// AddSession 添加会话
func (sm *SessionManager) AddSession(session Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[session.ID()] = session
}

// RemoveSession 移除会话
func (sm *SessionManager) RemoveSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) (Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, exists := sm.sessions[sessionID]
	return session, exists
}

// GetAllSessions 获取所有会话
func (sm *SessionManager) GetAllSessions() []Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	sessions := make([]Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// BroadcastToAll 向所有会话广播消息
func (sm *SessionManager) BroadcastToAll(envelope *protocol.Envelope) {
	sessions := sm.GetAllSessions()
	for _, session := range sessions {
		// 非阻塞发送，避免单个会话阻塞影响其他会话
		go func(s Session) {
			_ = s.SendEnvelope(envelope)
		}(session)
	}
}