package protocol

import (
	"fmt"
	"io"

	"github.com/hongjun500/chat-go/internal/protocol/pb"
	"google.golang.org/protobuf/proto"
)

// ProtobufCodec 将 Envelope 编码为 Protocol Buffers 格式
type ProtobufCodec struct{}

func (p *ProtobufCodec) Name() string {
	return Protobuf
}

func (p *ProtobufCodec) Encode(w io.Writer, e *Envelope) error {
	if e == nil {
		return fmt.Errorf("ProtobufCodec.Encode: envelope is nil")
	}
	if w == nil {
		return fmt.Errorf("ProtobufCodec.Encode: writer is nil")
	}
	
	protoMessage := &pb.Envelope{
		Version:       e.Version,
		Type:          toPBMsgType(e.Type),
		Encoding:      toPBEncoding(e.Encoding),
		MessageId:     e.Mid,
		CorrelationId: e.Correlation,
		From:          e.From,
		To:            e.To,
		Timestamp:     e.Ts,
		Data:          e.Data,
	}

	data, err := proto.Marshal(protoMessage)
	if err != nil {
		return fmt.Errorf("ProtobufCodec.Encode: failed to marshal protobuf message (Type=%s, Mid=%s): %w", e.Type, e.Mid, err)
	}
	
	if _, err = w.Write(data); err != nil {
		return fmt.Errorf("ProtobufCodec.Encode: failed to write marshaled data: %w", err)
	}
	return nil
}

// Decode ProtobufCodec 实现 Codec 接口
func (p *ProtobufCodec) Decode(r io.Reader, e *Envelope, maxSize int) error {
	if r == nil {
		return fmt.Errorf("ProtobufCodec.Decode: reader is nil")
	}
	if e == nil {
		return fmt.Errorf("ProtobufCodec.Decode: envelope is nil")
	}
	
	reader := r
	if maxSize > 0 {
		reader = io.LimitReader(r, int64(maxSize))
	}

	buf := make([]byte, maxSize)
	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("ProtobufCodec.Decode: failed to read data: %w", err)
	}
	
	if n == 0 {
		return fmt.Errorf("ProtobufCodec.Decode: no data to decode")
	}
	
	data := buf[:n]

	protoMessage := &pb.Envelope{}
	if err := proto.Unmarshal(data, protoMessage); err != nil {
		return fmt.Errorf("ProtobufCodec.Decode: failed to unmarshal protobuf data (size=%d bytes): %w", len(data), err)
	}

	// 将 pb.Envelope 转换为 protocol.Envelope
	result := Envelope{
		Version:     protoMessage.GetVersion(),
		Type:        fromPBMsgType(protoMessage.GetType()),
		Encoding:    fromPBEncoding(protoMessage.GetEncoding()),
		Mid:         protoMessage.GetMessageId(),
		Correlation: protoMessage.GetCorrelationId(),
		From:        protoMessage.GetFrom(),
		To:          protoMessage.GetTo(),
		Ts:          protoMessage.GetTimestamp(),
		Data:        protoMessage.GetData(),
	}
	
	if result.Type == "" {
		return fmt.Errorf("ProtobufCodec.Decode: decoded envelope missing required field 'type'")
	}
	
	*e = result
	return nil
}

// --- 辅助方法 ---

func toPBEncoding(enc Encoding) pb.Encoding {
	switch enc {
	case EncodingJSON:
		return pb.Encoding_ENCODING_JSON
	case EncodingProtobuf:
		return pb.Encoding_ENCODING_PROTOBUF
	default:
		return pb.Encoding_ENCODING_UNSPECIFIED
	}
}

func fromPBEncoding(enc pb.Encoding) Encoding {
	switch enc {
	case pb.Encoding_ENCODING_JSON:
		return EncodingJSON
	case pb.Encoding_ENCODING_PROTOBUF:
		return EncodingProtobuf
	default:
		return ""
	}
}

func toPBMsgType(t MessageType) pb.MessageType {
	switch t {
	case MsgText:
		return pb.MessageType_MSG_TYPE_TEXT
	case MsgCommand:
		return pb.MessageType_MSG_TYPE_COMMAND
	case MsgFileMeta:
		return pb.MessageType_MSG_TYPE_FILE_META
	case MsgFileChunk:
		return pb.MessageType_MSG_TYPE_FILE_CHUNK
	case MsgAck:
		return pb.MessageType_MSG_TYPE_ACK
	case MsgPing:
		return pb.MessageType_MSG_TYPE_PING
	case MsgPong:
		return pb.MessageType_MSG_TYPE_PONG
	default:
		return pb.MessageType_MSG_TYPE_UNSPECIFIED
	}
}

func fromPBMsgType(t pb.MessageType) MessageType {
	switch t {
	case pb.MessageType_MSG_TYPE_TEXT:
		return MsgText
	case pb.MessageType_MSG_TYPE_COMMAND:
		return MsgCommand
	case pb.MessageType_MSG_TYPE_FILE_META:
		return MsgFileMeta
	case pb.MessageType_MSG_TYPE_FILE_CHUNK:
		return MsgFileChunk
	case pb.MessageType_MSG_TYPE_ACK:
		return MsgAck
	case pb.MessageType_MSG_TYPE_PING:
		return MsgPing
	case pb.MessageType_MSG_TYPE_PONG:
		return MsgPong
	default:
		return ""
	}
}
