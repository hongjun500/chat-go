package transport

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

// FrameCodec 数据包的编解码器，使用长度前缀帧格式
type FrameCodec struct {
	readMu  sync.Mutex // 读锁
	writeMu sync.Mutex // 写锁
	bufPool *sync.Pool // 用于复用缓冲区
}

func NewFrameCodec() *FrameCodec {
	return &FrameCodec{
		bufPool: &sync.Pool{
			New: func() any {
				// 使用 64KB 缓冲区，适合大多数场景
				return make([]byte, 64*1024)
			},
		},
	}
}

// WriteFrame 写入一个帧
func (c *FrameCodec) WriteFrame(conn net.Conn, payload []byte) error {
	if c == nil {
		return fmt.Errorf("framecodec is nil")
	}
	if len(payload) > 16*1024*1024 { // 16MB hard limit
		return fmt.Errorf("frame too large: %d", len(payload))
	}
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(payload)))
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	// 先发长度，再发内容
	if _, err := conn.Write(header); err != nil {
		return err
	}
	if _, err := conn.Write(payload); err != nil {
		return err
	}
	return nil
}

// ReadFrame 读取一个帧
func (c *FrameCodec) ReadFrame(conn net.Conn /*, maxSize int*/) ([]byte, error) {
	if c == nil || conn == nil {
		return nil, fmt.Errorf("framecodec or reader is nil")
	}
	c.readMu.Lock()
	defer c.readMu.Unlock()
	header := make([]byte, 4)
	// 使用 io.ReadFull 确保读取完整的 4 字节长度
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}
	// 解析帧长度
	length := int(binary.BigEndian.Uint32(header))

	// 使用 bufPool 获取一个缓冲区，避免频繁分配
	buf := c.bufPool.Get().([]byte)
	if cap(buf) < length {
		// 容量不足，创建新缓冲区（旧缓冲区丢弃，由GC处理）
		buf = make([]byte, length)
	} else {
		// 复用缓冲区，调整长度
		buf = buf[:length]
	}
	if _, err := io.ReadFull(conn, buf); err != nil {
		c.bufPool.Put(buf) // 读取失败，放回缓冲池
		return nil, err
	}
	// 创建数据的拷贝以确保安全（调用者可以持有）
	data := make([]byte, length)
	copy(data, buf)

	// 放回缓冲区（重置为最大容量）
	c.bufPool.Put(buf[:cap(buf)])
	return data, nil
}
