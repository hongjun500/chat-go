package codec

import (
	"io"

	"github.com/hongjun500/chat-go/internal/protocol"
	"github.com/hongjun500/chat-go/internal/protocol/pb"
	"google.golang.org/protobuf/proto"
)

// ProtobufCodec 将 Envelope 编码为 Protocol Buffers 格式
type ProtobufCodec struct{}

func (p *ProtobufCodec) ContentType() string {
	return ApplicationProtobuf
}

func (p *ProtobufCodec) Encode(w io.Writer, e *protocol.Envelope) error {
	protoMessage := &pb.Envelope{
		Version:       e.Version,
		Type:          toPBMsgType(e.Type),
		Encoding:      toPBEncoding(e.Encoding),
		MessageId:     e.MessageID,
		CorrelationId: e.Correlation,
		From:          e.From,
		To:            e.To,
		Timestamp:     e.Timestamp,
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
func (p *ProtobufCodec) Decode(r io.Reader, e *protocol.Envelope, maxSize int) error {
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
	result := protocol.Envelope{
		Version:     protoMessage.GetVersion(),
		Type:        fromPBMsgType(protoMessage.GetType()),
		Encoding:    fromPBEncoding(protoMessage.GetEncoding()),
		MessageID:   protoMessage.GetMessageId(),
		Correlation: protoMessage.GetCorrelationId(),
		From:        protoMessage.GetFrom(),
		To:          protoMessage.GetTo(),
		Timestamp:   protoMessage.GetTimestamp(),
		Data:        protoMessage.GetData(),
	}
	*e = result
	return nil
}

// --- 辅助方法 ---

func toPBEncoding(enc protocol.Encoding) pb.Encoding {
	switch enc {
	case protocol.EncodingJSON:
		return pb.Encoding_ENCODING_JSON
	case protocol.EncodingProtobuf:
		return pb.Encoding_ENCODING_PROTOBUF
	case protocol.EncodingBinary:
		return pb.Encoding_ENCODING_BINARY
	default:
		return pb.Encoding_ENCODING_UNSPECIFIED
	}
}

func fromPBEncoding(enc pb.Encoding) protocol.Encoding {
	switch enc {
	case pb.Encoding_ENCODING_JSON:
		return protocol.EncodingJSON
	case pb.Encoding_ENCODING_PROTOBUF:
		return protocol.EncodingProtobuf
	case pb.Encoding_ENCODING_BINARY:
		return protocol.EncodingBinary
	default:
		return ""
	}
}

func toPBMsgType(t protocol.MessageType) pb.MessageType {
	switch t {
	case protocol.MsgText:
		return pb.MessageType_MSG_TYPE_TEXT
	case protocol.MsgCommand:
		return pb.MessageType_MSG_TYPE_COMMAND
	case protocol.MsgFileMeta:
		return pb.MessageType_MSG_TYPE_FILE_META
	case protocol.MsgFileChunk:
		return pb.MessageType_MSG_TYPE_FILE_CHUNK
	case protocol.MsgAck:
		return pb.MessageType_MSG_TYPE_ACK
	case protocol.MsgPing:
		return pb.MessageType_MSG_TYPE_PING
	case protocol.MsgPong:
		return pb.MessageType_MSG_TYPE_PONG
	default:
		return pb.MessageType_MSG_TYPE_UNSPECIFIED
	}
}

func fromPBMsgType(t pb.MessageType) protocol.MessageType {
	switch t {
	case pb.MessageType_MSG_TYPE_TEXT:
		return protocol.MsgText
	case pb.MessageType_MSG_TYPE_COMMAND:
		return protocol.MsgCommand
	case pb.MessageType_MSG_TYPE_FILE_META:
		return protocol.MsgFileMeta
	case pb.MessageType_MSG_TYPE_FILE_CHUNK:
		return protocol.MsgFileChunk
	case pb.MessageType_MSG_TYPE_ACK:
		return protocol.MsgAck
	case pb.MessageType_MSG_TYPE_PING:
		return protocol.MsgPing
	case pb.MessageType_MSG_TYPE_PONG:
		return protocol.MsgPong
	default:
		return ""
	}
}
