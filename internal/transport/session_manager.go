package transport

import (
	"sync"
	"sync/atomic"
)

// SessionManager 会话管理器
type SessionManager struct {
	sync.Map // key: id string, value: *SessionContext
	count    int64
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

// Add Session & return SessionContext
func (sm *SessionManager) Add(session Session) *SessionContext {
	sc := NewSessionContext(session)
	sm.Store(session.ID(), sc)
	atomic.AddInt64(&sm.count, 1)
	return sc
}

// Remove 移除会话
func (sm *SessionManager) Remove(id string) {
	sm.Delete(id)
	atomic.AddInt64(&sm.count, -1)
}

// Count 获取当前会话数量
func (sm *SessionManager) Count() int64 {
	return atomic.LoadInt64(&sm.count)
}

// Get 获取会话
func (sm *SessionManager) Get(id string) (*SessionContext, bool) {
	session, exists := sm.Load(id)
	if !exists {
		return nil, false
	}
	if sc, ok := session.(*SessionContext); ok {
		return sc, true
	}
	return nil, false
}

// GetAll 获取所有会话
func (sm *SessionManager) GetAll() []*SessionContext {
	scs := make([]*SessionContext, sm.Count())
	sm.Range(func(key, value any) bool {
		scs = append(scs, value.(*SessionContext))
		return true
	})
	return scs
}
