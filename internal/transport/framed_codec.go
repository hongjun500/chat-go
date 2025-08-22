package transport

import (
	"bytes"
)

// FramedMessageCodec combines frame processing with message encoding/decoding
// It provides a higher-level interface that handles both framing and message serialization
type FramedMessageCodec struct {
	frameCodec   *FrameCodec
	messageCodec MessageCodec
}

// NewFramedMessageCodec creates a new framed message codec
func NewFramedMessageCodec(frameCodec *FrameCodec, messageCodec MessageCodec) *FramedMessageCodec {
	return &FramedMessageCodec{
		frameCodec:   frameCodec,
		messageCodec: messageCodec,
	}
}

// Encode encodes a message using the message codec and writes it as a frame
func (fmc *FramedMessageCodec) Encode(msg *Envelope) error {
	var buf bytes.Buffer
	if err := fmc.messageCodec.Encode(&buf, msg); err != nil {
		return err
	}
	return fmc.frameCodec.WriteFrame(buf.Bytes())
}

// Decode reads a frame and decodes it using the message codec
func (fmc *FramedMessageCodec) Decode(msg *Envelope, maxSize int) error {
	frameData, err := fmc.frameCodec.ReadFrame(maxSize)
	if err != nil {
		return err
	}
	return fmc.messageCodec.Decode(bytes.NewReader(frameData), msg, maxSize)
}

// ContentType returns the content type of the underlying message codec
func (fmc *FramedMessageCodec) ContentType() string {
	return fmc.messageCodec.ContentType()
}