package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// LegacyFrameCodec provides backward compatibility for existing code that expects
// FrameCodec to have Encode/Decode methods with hardcoded JSON.
// This is deprecated and should be replaced with FramedMessageCodec.
type LegacyFrameCodec struct {
	*FrameCodec
}

// NewLegacyFrameCodec creates a legacy frame codec for backward compatibility
func NewLegacyFrameCodec(frameCodec *FrameCodec) *LegacyFrameCodec {
	return &LegacyFrameCodec{FrameCodec: frameCodec}
}

// Encode encodes an Envelope as JSON and writes it as a frame
// Deprecated: Use FramedMessageCodec with JSONCodec instead
func (c *LegacyFrameCodec) Encode(msg *Envelope) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.WriteFrame(payload)
}

// Decode reads a frame and decodes it as JSON into an Envelope
// Deprecated: Use FramedMessageCodec with JSONCodec instead
func (c *LegacyFrameCodec) Decode(msg *Envelope, maxSize int) error {
	buf, err := c.ReadFrame(maxSize)
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(bytes.TrimLeft(buf, " \t\r\n"), []byte("{")) {
		return fmt.Errorf("frame is not JSON object")
	}
	return json.Unmarshal(buf, msg)
}
