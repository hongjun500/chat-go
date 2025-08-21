package chat

import (
	"sync"

	"github.com/hongjun500/chat-go/internal/observe"
)

// Client 客户端连接实例
// 职责：维护用户状态与待发送消息缓冲；不直接操作底层连接。
type Client struct {
	ID        string
	Name      string
	Meta      map[string]string // 扩展元数据
	out       chan string
	closeOnce sync.Once
	closed    chan struct{}
}

// NewClientWithBuffer 允许指定发送缓冲区大小
func NewClientWithBuffer(id string, bufferSize int) *Client {
	if bufferSize <= 0 {
		bufferSize = 256
	}
	return &Client{
		ID:     id,
		out:    make(chan string, bufferSize),
		closed: make(chan struct{}),
		Meta:   make(map[string]string),
	}
}

// Send 非阻塞写入到 client 输出缓冲，缓冲溢出策略：暂时直接丢弃
func (c *Client) Send(message string) {
	select {
	case c.out <- message:
	default:
		// 缓冲已满：策略可以改为断开、覆盖或统计丢弃。
		// 这里选择简单丢弃以保证系统健康。
		observe.IncDropped()
	}
}

// Outgoing 返回只读输出通道，transport 读取并写到网络
func (c *Client) Outgoing() <-chan string {
	return c.out
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)
		close(c.out)
	})
}

// IsClosed 非阻塞判断是否已关闭
func (c *Client) IsClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}
