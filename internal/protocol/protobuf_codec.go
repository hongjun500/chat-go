package protocol

import (
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
	protoMessage := &pb.Envelope{
		Version:       e.Version,
		Type:          toPBMsgType(e.Type),
		Encoding:      toPBEncoding(e.Encoding),
		MessageId:     e.Mid,
		CorrelationId: e.Correlation,
		Timestamp:     e.Ts,
		Data:          e.Data,
	}

	data, err := proto.Marshal(protoMessage)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// Decode ProtobufCodec 实现 Codec 接口
func (p *ProtobufCodec) Decode(r io.Reader, e *Envelope, maxSize int) error {
	reader := r
	if maxSize > 0 {
		reader = io.LimitReader(r, int64(maxSize))
	}

	buf := make([]byte, maxSize)
	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}
	data := buf[:n]

	protoMessage := &pb.Envelope{}
	if err := proto.Unmarshal(data, protoMessage); err != nil {
		return err
	}

	// 将 pb.Envelope 转换为 protocol.Envelope
	result := Envelope{
		Version:     protoMessage.GetVersion(),
		Type:        fromPBMsgType(protoMessage.GetType()),
		Encoding:    fromPBEncoding(protoMessage.GetEncoding()),
		Mid:         protoMessage.GetMessageId(),
		Correlation: protoMessage.GetCorrelationId(),
		Ts:          protoMessage.GetTimestamp(),
		Data:        protoMessage.GetData(),
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
