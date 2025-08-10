package transport

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

// TCPMode indicates the payload format for TCP transport
type TCPMode string

const (
	// TCPModeLegacy uses line-based plain text (backward compatible)
	TCPModeLegacy TCPMode = "legacy"
	// TCPModeJSON uses length-prefixed JSON frames
	TCPModeJSON TCPMode = "json"
)

// Envelope is a unified message envelope transmitted over transports
// For simplicity we keep a superset of possible fields; clients/server fill what they need
type Envelope struct {
	Type    string            `json:"type"`              // text|set_name|chat|direct|command|ping|pong|ack|notice|file_meta
	Text    string            `json:"text,omitempty"`    // for Type=text/notice and server prompts
	Name    string            `json:"name,omitempty"`    // for Type=set_name
	Content string            `json:"content,omitempty"` // for Type=chat/direct
	From    string            `json:"from,omitempty"`
	To      string            `json:"to,omitempty"`
	Raw     string            `json:"raw,omitempty"`    // for Type=command
	Mid     string            `json:"mid,omitempty"`    // message id used with ack
	Status  string            `json:"status,omitempty"` // for Type=ack
	Seq     int64             `json:"seq,omitempty"`    // for Type=ping/pong
	Meta    map[string]string `json:"meta,omitempty"`
	Ts      int64             `json:"ts,omitempty"` // unix ms
}

// FrameCodec encodes/decodes length-prefixed JSON frames: [len uint32 BE][payload bytes]
type FrameCodec struct {
	r *bufio.Reader
	w *bufio.Writer
}

func NewFrameCodec(conn net.Conn) *FrameCodec {
	return &FrameCodec{r: bufio.NewReader(conn), w: bufio.NewWriter(conn)}
}

func (c *FrameCodec) Encode(msg *Envelope) error {
	// Backward-compatible helper: JSON marshal then write frame
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.WriteFrame(payload)
}

func (c *FrameCodec) Decode(msg *Envelope, maxSize int) error {
	// Backward-compatible helper: read frame then JSON unmarshal
	buf, err := c.ReadFrame(maxSize)
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(bytes.TrimLeft(buf, " \t\r\n"), []byte("{")) {
		return fmt.Errorf("frame is not JSON object")
	}
	return json.Unmarshal(buf, msg)
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
