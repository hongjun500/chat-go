package transport

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

const (
	MaxFrameSize = 16 * 1024 * 1024 // 安全上限
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
func (c *FrameCodec) ReadFrame(conn net.Conn) ([]byte, error) {
	buf, n, err := c.readFrameInternal(conn)
	if err != nil {
		return nil, err
	}
	// 安全拷贝
	data := make([]byte, n)
	copy(data, (*buf)[:n])

	// 放回缓冲池（重置为 0 长度）
	*buf = (*buf)[:0]
	c.bufPool.Put(buf)
	return data, nil
}

// 内部通用读取逻辑，返回 *buf 和有效长度
func (c *FrameCodec) readFrameInternal(conn net.Conn) (*[]byte, int, error) {
	if c == nil || conn == nil {
		return nil, 0, fmt.Errorf("framecodec or conn is nil")
	}
	c.readMu.Lock()
	defer c.readMu.Unlock()

	// 读取长度头
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, 0, err
	}
	length := int(binary.BigEndian.Uint32(header))

	// 安全检查
	if length <= 0 || length > MaxFrameSize {
		return nil, 0, fmt.Errorf("invalid frame size: %d", length)
	}

	// 从池里拿 buffer
	bufPtr := c.bufPool.Get().(*[]byte)
	buf := *bufPtr
	if cap(buf) < length {
		buf = make([]byte, length)
	}
	buf = buf[:length]

	// 填充数据
	if _, err := io.ReadFull(conn, buf); err != nil {
		*bufPtr = (*bufPtr)[:0] // 归还前 reset
		c.bufPool.Put(bufPtr)
		return nil, 0, err
	}

	*bufPtr = buf
	return bufPtr, length, nil
}
