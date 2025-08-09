package chat

import (
	"net"
	"sync"
)

type Client struct {
	ID   string
	Name string
	Conn net.Conn

	out    chan string
	mu     sync.Mutex
	closed bool
}

func NewClient(id string, conn net.Conn) *Client {
	return &Client{
		ID:   id,
		Conn: conn,
		out:  make(chan string, 16),
	}
}

func (c *Client) Send(message string) {
	select {
	case c.out <- message:
	default:
		// 避免阻塞，如果缓冲满了就丢弃（也可以记录做其它处理）
	}
}

func (c *Client) Outgoing() <-chan string {
	return c.out
}

func (c *Client) closeSend() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	close(c.out)
	c.closed = true
}
