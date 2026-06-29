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

// AddContext 将已有的 SessionContext 注册到管理器
func (sm *SessionManager) AddContext(sc *SessionContext) *SessionContext {
	if sc == nil {
		return nil
	}
	sm.Store(sc.Id, sc)
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
	// 修复：避免使用 make with count 然后 append 导致长度问题，改用动态构建切片
	scs := make([]*SessionContext, 0)
	sm.Range(func(key, value any) bool {
		if v, ok := value.(*SessionContext); ok {
			scs = append(scs, v)
		}
		return true
	})
	return scs
}
