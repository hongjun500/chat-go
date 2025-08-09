package chat

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type EventHandler func(Event)

type Hub struct {
	clients sync.Map // key: client.ID -> *Client

	// 按 EventType 注册的处理器
	handlersMu sync.RWMutex
	handlers   map[EventType][]EventHandler
}

func NewHub() *Hub {
	return &Hub{
		handlers: make(map[EventType][]EventHandler),
	}
}

// Subscribe 注册事件处理器
func (h *Hub) Subscribe(t EventType, fn EventHandler) {
	h.handlersMu.Lock()
	defer h.handlersMu.Unlock()
	h.handlers[t] = append(h.handlers[t], fn)
}

// Emit 异步分发事件给所有 handler,非阻塞返回
func (h *Hub) Emit(e Event) {
	h.handlersMu.RLock()
	handlers, ok := h.handlers[e.Type()]
	// 拷贝切片以避免并发修改影响
	var copied []EventHandler
	if ok && len(handlers) > 0 {
		copied = append(copied, handlers...)
	}
	h.handlersMu.RUnlock()
	if len(copied) == 0 {
		return
	}
	for _, fn := range copied {
		// 异步调用，保护 handler 崩溃不影响其他
		go func(f EventHandler) {
			defer func() { _ = recover() }()
			f(e)
		}(fn)
	}
}

// RegisterClient 注册客户端并发出 UserJoined 事件
func (h *Hub) RegisterClient(c *Client) {
	h.clients.Store(c.ID, c)
	h.Emit(&UserEvent{When: time.Now(), User: c, Desc: "joined"})
}

// UnregisterClient 注销客户端并发出 UserLeave 事件
func (h *Hub) UnregisterClient(c *Client) {
	if _, loaded := h.clients.LoadAndDelete(c.ID); loaded {
		c.Close()
		h.Emit(&UserEvent{When: time.Now(), User: c, Desc: "leave"})
		return
	}
	// 即便未加载成功，也确保连接被关闭
	c.Close()
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

func formatMsg(m *Message) string {
	t := m.Timestamp.Format("2006-01-02 15:04:05")
	if strings.EqualFold(m.From, "系统") {
		return fmt.Sprintf("[%s] %s: %s", t, m.From, m.Content)
	}
	return fmt.Sprintf("[%s] %s: %s", t, m.From, m.Content)
}
