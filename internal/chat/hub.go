package chat

import (
	"sync"
	"time"

	"github.com/hongjun500/chat-go/internal/observe"
)

type EventHandler func(Event)

type handlerEntry struct {
	id uint64
	fn EventHandler
}

type Hub struct {
	clients sync.Map // key: client.ID -> *Client

	// 按 EventType 注册的处理器
	handlersMu sync.RWMutex
	handlers   map[EventType][]handlerEntry
	nextHID    uint64

	// 封禁名单：用户名 -> 过期时间（零值表示永久）
	banMu  sync.RWMutex
	banned map[string]time.Time
}

func NewHub() *Hub {
	return &Hub{
		handlers: make(map[EventType][]handlerEntry),
		banned:   make(map[string]time.Time),
	}
}

// Subscribe 注册事件处理器
func (h *Hub) Subscribe(t EventType, fn EventHandler) { _ = h.SubscribeCancelable(t, fn) }

// SubscribeCancelable 注册并返回一个取消函数，用于移除该处理器
func (h *Hub) SubscribeCancelable(t EventType, fn EventHandler) (cancel func()) {
	h.handlersMu.Lock()
	h.nextHID++
	id := h.nextHID
	h.handlers[t] = append(h.handlers[t], handlerEntry{id: id, fn: fn})
	h.handlersMu.Unlock()

	return func() {
		h.handlersMu.Lock()
		entries := h.handlers[t]
		if len(entries) > 0 {
			filtered := entries[:0]
			for _, e := range entries {
				if e.id != id {
					filtered = append(filtered, e)
				}
			}
			if len(filtered) == 0 {
				delete(h.handlers, t)
			} else {
				h.handlers[t] = append([]handlerEntry(nil), filtered...)
			}
		}
		h.handlersMu.Unlock()
	}
}

// Emit 异步分发事件给所有 handler,非阻塞返回
func (h *Hub) Emit(e Event) {
	h.handlersMu.RLock()
	entries, ok := h.handlers[e.Type()]
	// 拷贝切片以避免并发修改影响
	var copied []handlerEntry
	if ok && len(entries) > 0 {
		copied = append(copied, entries...)
	}
	h.handlersMu.RUnlock()
	if len(copied) == 0 {
		return
	}
	for _, entry := range copied {
		go func(f EventHandler) {
			defer func() { _ = recover() }()
			f(e)
		}(entry.fn)
	}
}

// RegisterClient 注册客户端并发出 UserJoined 事件
func (h *Hub) RegisterClient(c *Client) {
	h.clients.Store(c.ID, c)
	h.Emit(&UserEvent{When: time.Now(), User: c, Desc: "joined"})
	observe.AddOnline(1)
}

// UnregisterClient 注销客户端并发出 UserLeave 事件
func (h *Hub) UnregisterClient(c *Client) {
	if _, loaded := h.clients.LoadAndDelete(c.ID); loaded {
		c.Close()
		h.Emit(&UserEvent{When: time.Now(), User: c, Desc: "leave"})
		observe.AddOnline(-1)
		return
	}
	// 即便未加载成功，也确保连接被关闭
	c.Close()
}

// KickByName 注销所有昵称匹配的客户端，返回是否找到
func (h *Hub) KickByName(name string) bool {
	found := false
	h.clients.Range(func(_, v any) bool {
		if c, ok := v.(*Client); ok && c.Name == name {
			h.UnregisterClient(c)
			found = true
		}
		return true
	})
	return found
}

// BanFor 将用户名封禁指定时长；d<=0 表示永久
func (h *Hub) BanFor(name string, d time.Duration) {
	h.banMu.Lock()
	if d <= 0 {
		h.banned[name] = time.Time{}
	} else {
		h.banned[name] = time.Now().Add(d)
	}
	h.banMu.Unlock()
}

// IsBanned 判断用户名是否在封禁名单（过期会自动清理）
func (h *Hub) IsBanned(name string) bool {
	h.banMu.RLock()
	until, ok := h.banned[name]
	h.banMu.RUnlock()
	if !ok {
		return false
	}
	if until.IsZero() {
		return true
	}
	if time.Now().Before(until) {
		return true
	}
	// 过期清理
	h.banMu.Lock()
	delete(h.banned, name)
	h.banMu.Unlock()
	return false
}

// BroadcastLocal 触发本地消息事件
func (h *Hub) BroadcastLocal(from, content string) {
	h.Emit(&MessageEvent{When: time.Now(), From: from, Content: content, Local: true})
}

// BroadcastRemote 触发远端同步消息事件（来自其它节点）
func (h *Hub) BroadcastRemote(from, content string, t time.Time) {
	h.Emit(&MessageEvent{When: t, From: from, Content: content, Local: false})
}

// ListNames 返回在线用户名（简单实现）
func (h *Hub) ListNames() []string {
	var out []string
	h.clients.Range(func(k, v any) bool {
		if c, ok := v.(*Client); ok {
			out = append(out, c.Name)
		}
		return true
	})
	return out
}

// SendToAll 用于本地广播（handler 可调用），直接将 msg 发到每个 client.Send()
func (h *Hub) SendToAll(msg string) {
	h.clients.Range(func(_, v any) bool {
		if c, ok := v.(*Client); ok {
			c.Send(msg)
		}
		return true
	})
}

// SendToUser 按用户名点对点发送，返回是否找到目标
func (h *Hub) SendToUser(userName string, msg string) bool {
	found := false
	h.clients.Range(func(_, v any) bool {
		c, ok := v.(*Client)
		if !ok {
			return true
		}
		if c.Name == userName {
			c.Send(msg)
			found = true
			// 不 break，避免同名并发连接时发送给多个；如需只发一个可返回 false 以中止
		}
		return true
	})
	return found
}
