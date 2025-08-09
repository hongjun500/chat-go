package chat

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Hub struct {
	// 注册 / 注销通道由 transport 层使用
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	save       chan *Message

	clients map[string]*Client
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 128),
		save:       make(chan *Message, 256),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c.ID] = c
			h.mu.Unlock()
			c.Send("欢迎！请输入你的昵称：")
		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c.ID]; ok {
				delete(h.clients, c.ID)
				c.closeSend()
				// 广播离开的消息
				if c.Name != "" || strings.TrimSpace(c.Name) != "" {
					h.broadcast <- &Message{
						From:      "系统",
						Content:   fmt.Sprintf("%s 已离开", c.Name),
						Timestamp: time.Now(),
					}
				}
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.mu.Lock()
			for _, cl := range h.clients {
				cl.Send(formatMsg(msg))
			}
			h.mu.Unlock()
		}

	}
}

func formatMsg(m *Message) string {
	t := m.Timestamp.Format("2006-01-02 15:04:05")
	if strings.EqualFold(m.From, "系统") {
		return fmt.Sprintf("[%s] %s: %s", t, m.From, m.Content)
	}
	return fmt.Sprintf("[%s] %s: %s", t, m.From, m.Content)
}

// ListNames 返回所有在线用户昵称，如果未设置昵称用 ID 的短形式
func (h *Hub) ListNames() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]string, 0, len(h.clients))
	for _, c := range h.clients {
		if c.Name != "" {
			out = append(out, c.Name)
		} else {
			out = append(out, c.ID)
		}
	}
	return out
}

func (h *Hub) Register(c *Client) {
	h.register <- c
}

func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

func (h *Hub) Broadcast(msg *Message) {
	select {
	case h.broadcast <- msg:
	default:

	}
	select {
	case h.save <- msg:
	default:

	}
}

func (h *Hub) SaveChannel() chan *Message {
	return h.save
}
