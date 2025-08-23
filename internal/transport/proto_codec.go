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

// 定义统一的 Codec 接口
// ProtobufCodec 实现 Codec 接口
func (p *ProtobufCodec) Decode(r io.Reader, m *Envelope, maxSize int) error {
	// Apply size limit if specified
	reader := r
	if maxSize > 0 {
		reader = io.LimitReader(r, int64(maxSize))
	}
	
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	protoMessage := &EnvelopeProto{}
	if err := proto.Unmarshal(data, protoMessage); err != nil {
		return err
	}
	// 将 EnvelopeProto 转换为 Envelope
	*m = Envelope{
		Version:      protoMessage.Version,
		Type:         protoMessage.Type,
		Schema:       protoMessage.Schema,
		Datacontent:  protoMessage.Datacontent,
		Mid:          protoMessage.Mid,
		Correlation:  protoMessage.Correlation,
		Causation:    protoMessage.Causation,
		TraceID:      protoMessage.TraceId,
		Tenant:       protoMessage.Tenant,
		Conversation: protoMessage.Conversation,
		From:         protoMessage.From,
		To:           protoMessage.To,
		PartitionKey: protoMessage.PartitionKey,
		Ts:           protoMessage.Ts,
		TTLms:        protoMessage.TtlMs,
		ExpiresAt:    protoMessage.ExpiresAt,
		Meta:         protoMessage.Meta,
		Signature:    protoMessage.Signature,
		Encrypted:    protoMessage.Encrypted,
		Priority:     int(protoMessage.Priority),
		ChunkIndex:   int(protoMessage.ChunkIndex),
		TotalChunks:  int(protoMessage.TotalChunks),
		Language:     protoMessage.Language,
		Status:       protoMessage.Status,
		Payload:      protoMessage.Payload,
		Data:         protoMessage.Data,
		Attributes:   protoMessage.Attributes,
	}
	return nil
}
