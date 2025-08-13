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

// Envelope defines a thin, evolvable message header plus a polymorphic payload.
//
// Design goals:
// - Keep transport-agnostic concerns (routing, reliability, observability) in the header
// - Carry business-specific data in Payload (JSON) or Data (binary)
// - Avoid a "fat" struct with many optional fields; add new features by adding new payload types
// - One frame carries exactly one Envelope
//
// Typical types: text|set_name|chat|direct|command|ping|pong|ack|file_meta|file_chunk
type Envelope struct {
	// Protocol evolution
	Version     string `json:"version,omitempty"`     // logical protocol version
	Type        string `json:"type"`                  // discriminator for payload
	Schema      string `json:"schema,omitempty"`      // optional payload schema identifier
	Datacontent string `json:"datacontent,omitempty"` // e.g. application/json, application/octet-stream

	// Reliability & tracing
	Mid         string `json:"mid,omitempty"` // message id for idempotency
	Correlation string `json:"correlation_id,omitempty"`
	Causation   string `json:"causation_id,omitempty"`
	TraceID     string `json:"trace_id,omitempty"`

	// Routing & tenancy (optional)
	Tenant       string   `json:"tenant,omitempty"`
	Conversation string   `json:"conversation_id,omitempty"`
	From         string   `json:"from,omitempty"`
	To           []string `json:"to,omitempty"`
	PartitionKey string   `json:"partition_key,omitempty"`

	// Time & flow control
	Ts        int64             `json:"ts,omitempty"` // unix ms
	TTLms     int64             `json:"ttl_ms,omitempty"`
	ExpiresAt int64             `json:"expires_at,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`

	// Payloads
	Payload json.RawMessage `json:"payload,omitempty"` // structured payload (JSON)
	Data    []byte          `json:"data,omitempty"`    // large/binary payload; JSON base64-encoded
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
