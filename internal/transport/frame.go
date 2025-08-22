package transport

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// FrameCodec encodes/decodes length-prefixed frames: [len uint32 BE][payload bytes]
// This codec is transport-agnostic and doesn't assume any specific message format.
type FrameCodec struct {
	r *bufio.Reader
	w *bufio.Writer
}

// NewFrameCodec creates a new frame codec for the given connection
func NewFrameCodec(conn net.Conn) *FrameCodec {
	return &FrameCodec{r: bufio.NewReader(conn), w: bufio.NewWriter(conn)}
}

// WriteFrame writes a length-prefixed payload
func (c *FrameCodec) WriteFrame(payload []byte) error {
	if c == nil || c.w == nil {
		return fmt.Errorf("codec or writer is nil")
	}
	if len(payload) > 16*1024*1024 { // 16MB hard limit
		return fmt.Errorf("frame too large: %d", len(payload))
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	if _, err := c.w.Write(header[:]); err != nil {
		return err
	}
	if _, err := c.w.Write(payload); err != nil {
		return err
	}
	return c.w.Flush()
}

// ReadFrame reads a single length-prefixed payload
func (c *FrameCodec) ReadFrame(maxSize int) ([]byte, error) {
	if c == nil || c.r == nil {
		return nil, fmt.Errorf("codec or reader is nil")
	}
	var header [4]byte
	if _, err := io.ReadFull(c.r, header[:]); err != nil {
		return nil, err
	}
	n := int(binary.BigEndian.Uint32(header[:]))
	if n <= 0 || (maxSize > 0 && n > maxSize) {
		return nil, fmt.Errorf("invalid frame size: %d", n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(c.r, buf); err != nil {
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