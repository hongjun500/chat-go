package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type JSONCodec struct{}

func (JSONCodec) Encode(w io.Writer, m *Envelope) error {
	enc := json.NewEncoder(w)
	return enc.Encode(m)
}

func (JSONCodec) Decode(r io.Reader, m *Envelope, maxSize int) error {
	// If r is already limited by framed transport we rely on that. Otherwise, guard with io.LimitReader
	rr := r
	if maxSize > 0 {
		rr = io.LimitReader(r, int64(maxSize))
	}
	dec := json.NewDecoder(rr)
	if err := dec.Decode(m); err != nil {
		return err
	}
	// Basic sanity check to avoid nested arrays etc.
	// json.Decoder with Encode adds a trailing newline; tolerate it.
	// Optionally validate that the raw starts with '{'
	if m.Type == "" {
		// Not strictly required, but helps catch malformed inputs
		return fmt.Errorf("missing field: type")
	}
	// Verify it's likely a JSON object. This is a soft check when rr is bytes.Reader etc.
	if lb, ok := rr.(*bytes.Reader); ok {
		b, _ := lb.ReadByte()
		if b != '{' {
			return fmt.Errorf("payload not object")
		}
	}
	return nil
}

func (JSONCodec) ContentType() string { return "application/json" }
