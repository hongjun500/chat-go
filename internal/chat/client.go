package chat

import (
	"net"
	"sync"

	"github.com/hongjun500/chat-go/internal/observe"
)

type ConnWriter interface {
	Write([]byte) error
	Close() error
}

// Client 客户端连接实例
type Client struct {
	ID        string
	Name      string
	Conn      net.Conn
	Meta      map[string]string // 扩展元数据
	out       chan string
	closeOnce sync.Once
	closed    chan struct{}

	mu sync.Mutex
}

// NewClientWithBuffer 允许指定发送缓冲区大小
func NewClientWithBuffer(id string, conn net.Conn, bufferSize int) *Client {
	if bufferSize <= 0 {
		bufferSize = 256
	}
	return &Client{
		ID:     id,
		Conn:   conn,
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
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
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

// delete
func (c *Client) closeSend() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if true {
		return
	}
	close(c.out)
	c.closed = make(chan struct{})
}
