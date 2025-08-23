package protocol

import (
	"encoding/json"
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

	// Security
	Signature string `json:"signature,omitempty"` // message signature for integrity verification
	Encrypted bool   `json:"encrypted,omitempty"` // indicates if the payload is encrypted

	// Priority
	Priority int `json:"priority,omitempty"` // message priority level

	// Sharding
	ChunkIndex  int `json:"chunk_index,omitempty"`  // index of the current chunk
	TotalChunks int `json:"total_chunks,omitempty"` // total number of chunks

	// Localization
	Language string `json:"language,omitempty"` // message language

	// Status
	Status string `json:"status,omitempty"` // message status (e.g., sent, received, read)

	// Payloads
	Payload    json.RawMessage   `json:"payload,omitempty"`    // structured payload (JSON)
	Data       []byte            `json:"data,omitempty"`       // large/binary payload; JSON base64-encoded
	Attributes map[string]string `json:"attributes,omitempty"` // custom attributes for extensibility
}

// ---- Typed payloads and helpers ----

type TextPayload struct {
	Text string `json:"text"`
}
type SetNamePayload struct {
	Name string `json:"name"`
}
type ChatPayload struct {
	Content string `json:"content"`
}
type DirectPayload struct {
	Content string   `json:"content"`
	To      []string `json:"to"`
}
type CommandPayload struct {
	Raw string `json:"raw"`
}
type AckPayload struct {
	Status string `json:"status"`
}
type PingPayload struct {
	Seq int64 `json:"seq"`
}
type PongPayload struct {
	Seq int64 `json:"seq"`
}
