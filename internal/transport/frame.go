package transport

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// FrameCodec 数据包的编解码器，使用长度前缀帧格式
type FrameCodec struct {
	conn    net.Conn
	readMu  sync.Mutex // 读锁
	writeMu sync.Mutex // 写锁
	bufPool *sync.Pool // 用于复用缓冲区
}

func NewFrameCodec(conn net.Conn) *FrameCodec {
	return &FrameCodec{
		conn: conn,
		bufPool: &sync.Pool{
			New: func() any {
				// 使用 64KB 缓冲区，适合大多数场景
				return make([]byte, 64*1024)
			},
		},
	}
}

// WriteFrame 写入一个帧
func (c *FrameCodec) WriteFrame(payload []byte) error {
	if c == nil || c.conn == nil {
		return fmt.Errorf("framecodec or writer is nil")
	}
	if len(payload) > 16*1024*1024 { // 16MB hard limit
		return fmt.Errorf("frame too large: %d", len(payload))
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	// 零拷贝写，避免 bufio 额外拷贝
	buffers := net.Buffers{header[:], payload}
	_, err := buffers.WriteTo(c.conn)
	return err
}

// ReadFrame 读取一个帧
func (c *FrameCodec) ReadFrame(maxSize int) ([]byte, error) {
	if c == nil || c.conn == nil {
		return nil, fmt.Errorf("framecodec or reader is nil")
	}
	var header [4]byte
	c.readMu.Lock()
	defer c.readMu.Unlock()
	// 使用 io.ReadFull 确保读取完整的 4 字节长度
	if _, err := io.ReadFull(c.conn, header[:]); err != nil {
		return nil, err
	}
	// 解析帧长度
	n := int(binary.BigEndian.Uint32(header[:]))
	if n <= 0 || (maxSize > 0 && n > maxSize) {
		return nil, fmt.Errorf("invalid frame size: %d", n)
	}
	// 使用 bufPool 获取一个缓冲区，避免频繁分配
	buf := c.bufPool.Get().([]byte)
	if cap(buf) < n {
		buf = make([]byte, n) // 如果缓冲区不够大，重新分配
	} else {
		buf = buf[:n] // 如果缓冲区足够大，重置长度
	}

	defer func() {
		if c == nil || buf == nil {
			return
		}
		c.bufPool.Put(buf) // 将缓冲区放回池中以供复用
	}()

	if _, err := io.ReadFull(c.conn, buf); err != nil {
		c.bufPool.Put(buf) // 读取失败，放回缓冲池
		return nil, err
	}
	return buf, nil
}

// SafeDeadline applies deadline if d>0
func SafeDeadline(conn net.Conn, d time.Duration) {
	if conn == nil || d <= 0 {
		return
	}
	_ = conn.SetDeadline(time.Now().Add(d))
}
