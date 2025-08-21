package transport

import (
	"io"

	"google.golang.org/protobuf/proto"
)

// ProtobufCodec 将 Envelope 编码为 Protocol Buffers 格式
// 使用 m.proto 中定义的 EnvelopeProto 消息
type ProtobufCodec struct{}

func (p *ProtobufCodec) ContentType() string {
	return ApplicationProtobuf
}

func (p *ProtobufCodec) Encode(w io.Writer, m *Envelope) error {
	protoMessage := &EnvelopeProto{
		Version:      m.Version,
		Type:         m.Type,
		Schema:       m.Schema,
		Datacontent:  m.Datacontent,
		Mid:          m.Mid,
		Correlation:  m.Correlation,
		Causation:    m.Causation,
		TraceId:      m.TraceID,
		Tenant:       m.Tenant,
		Conversation: m.Conversation,
		From:         m.From,
		To:           m.To,
		PartitionKey: m.PartitionKey,
		Ts:           m.Ts,
		TtlMs:        m.TTLms,
		ExpiresAt:    m.ExpiresAt,
		Meta:         m.Meta,
		Signature:    m.Signature,
		Encrypted:    m.Encrypted,
		Priority:     int32(m.Priority),
		ChunkIndex:   int32(m.ChunkIndex),
		TotalChunks:  int32(m.TotalChunks),
		Language:     m.Language,
		Status:       m.Status,
		Payload:      m.Payload,
		Data:         m.Data,
		Attributes:   m.Attributes,
	}
	data, err := proto.Marshal(protoMessage)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (p *ProtobufCodec) Decode(r io.Reader, m *Envelope, maxSize int) error {
	data := make([]byte, maxSize)
	n, err := r.Read(data)
	if err != nil {
		return err
	}
	protoMessage := &EnvelopeProto{}
	if err := proto.Unmarshal(data[:n], protoMessage); err != nil {
		return err
	}
	m.Version = protoMessage.Version
	m.Type = protoMessage.Type
	m.Schema = protoMessage.Schema
	m.Datacontent = protoMessage.Datacontent
	m.Mid = protoMessage.Mid
	m.Correlation = protoMessage.Correlation
	m.Causation = protoMessage.Causation
	m.TraceID = protoMessage.TraceId
	m.Tenant = protoMessage.Tenant
	m.Conversation = protoMessage.Conversation
	m.From = protoMessage.From
	m.To = protoMessage.To
	m.PartitionKey = protoMessage.PartitionKey
	m.Ts = protoMessage.Ts
	m.TTLms = protoMessage.TtlMs
	m.ExpiresAt = protoMessage.ExpiresAt
	m.Meta = protoMessage.Meta
	m.Signature = protoMessage.Signature
	m.Encrypted = protoMessage.Encrypted
	m.Priority = int(protoMessage.Priority)
	m.ChunkIndex = int(protoMessage.ChunkIndex)
	m.TotalChunks = int(protoMessage.TotalChunks)
	m.Language = protoMessage.Language
	m.Status = protoMessage.Status
	m.Payload = protoMessage.Payload
	m.Data = protoMessage.Data
	m.Attributes = protoMessage.Attributes
	return nil
}
