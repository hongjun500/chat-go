package transport

import "io"

// MessageCodec abstracts payload encoding for framed transports (e.g., JSON, Protobuf)
type MessageCodec interface {
	Encode(w io.Writer, m *Envelope) error
	Decode(r io.Reader, m *Envelope, maxSize int) error
	ContentType() string
}
