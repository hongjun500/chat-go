package protocol

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// PayloadCodec provides centralized encoding/decoding for envelope payloads
// It maps message types to their corresponding payload structures
type PayloadCodec struct {
	typeRegistry map[MessageType]reflect.Type
}

// NewPayloadCodec creates a new payload codec with built-in type mappings
func NewPayloadCodec() *PayloadCodec {
	pc := &PayloadCodec{
		typeRegistry: make(map[MessageType]reflect.Type),
	}
	
	// Register built-in payload types
	pc.RegisterPayloadType(MsgText, reflect.TypeOf(TextPayload{}))
	pc.RegisterPayloadType("set_name", reflect.TypeOf(SetNamePayload{}))
	pc.RegisterPayloadType("chat", reflect.TypeOf(ChatPayload{}))
	pc.RegisterPayloadType("direct", reflect.TypeOf(DirectPayload{}))
	pc.RegisterPayloadType(MsgCommand, reflect.TypeOf(CommandPayload{}))
	pc.RegisterPayloadType(MsgAck, reflect.TypeOf(AckPayload{}))
	pc.RegisterPayloadType(MsgPing, reflect.TypeOf(PingPayload{}))
	pc.RegisterPayloadType(MsgPong, reflect.TypeOf(PongPayload{}))
	pc.RegisterPayloadType(MsgFileMeta, reflect.TypeOf(FileMetaPayload{}))
	pc.RegisterPayloadType(MsgFileChunk, reflect.TypeOf(FileChunkPayload{}))
	
	return pc
}

// RegisterPayloadType registers a new message type with its corresponding payload structure
func (pc *PayloadCodec) RegisterPayloadType(msgType MessageType, payloadType reflect.Type) {
	pc.typeRegistry[msgType] = payloadType
}

// DecodePayload decodes the JSON payload from an envelope into the appropriate struct
func (pc *PayloadCodec) DecodePayload(envelope *Envelope, target interface{}) error {
	if envelope.Payload == nil || len(envelope.Payload) == 0 {
		return fmt.Errorf("empty payload for message type: %s", envelope.Type)
	}
	
	return json.Unmarshal(envelope.Payload, target)
}

// DecodePayloadByType decodes the JSON payload and returns a typed payload struct
func (pc *PayloadCodec) DecodePayloadByType(envelope *Envelope) (interface{}, error) {
	payloadType, exists := pc.typeRegistry[envelope.Type]
	if !exists {
		return nil, fmt.Errorf("unknown message type: %s", envelope.Type)
	}
	
	// Create a new instance of the payload type
	payloadPtr := reflect.New(payloadType).Interface()
	
	if err := pc.DecodePayload(envelope, payloadPtr); err != nil {
		return nil, fmt.Errorf("failed to decode payload for type %s: %w", envelope.Type, err)
	}
	
	// Return the value (not pointer) of the payload
	return reflect.ValueOf(payloadPtr).Elem().Interface(), nil
}

// EncodePayload encodes a payload struct into JSON for use in an envelope
func (pc *PayloadCodec) EncodePayload(payload interface{}) (json.RawMessage, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode payload: %w", err)
	}
	return json.RawMessage(data), nil
}

// CreateEnvelope creates a new envelope with the specified type and payload
func (pc *PayloadCodec) CreateEnvelope(msgType MessageType, payload interface{}) (*Envelope, error) {
	encodedPayload, err := pc.EncodePayload(payload)
	if err != nil {
		return nil, err
	}
	
	return &Envelope{
		Type:    msgType,
		Payload: encodedPayload,
	}, nil
}

// MustEncodePayload is like EncodePayload but panics on error (for backward compatibility)
func (pc *PayloadCodec) MustEncodePayload(payload interface{}) json.RawMessage {
	data, err := pc.EncodePayload(payload)
	if err != nil {
		panic(fmt.Sprintf("failed to encode payload: %v", err))
	}
	return data
}

// SafeDecodePayload attempts to decode a payload, returning false if it fails
func (pc *PayloadCodec) SafeDecodePayload(envelope *Envelope, target interface{}) bool {
	return pc.DecodePayload(envelope, target) == nil
}

// PayloadDecoder provides a convenient interface for decoding specific payload types
type PayloadDecoder struct {
	codec *PayloadCodec
}

// NewPayloadDecoder creates a new payload decoder using the provided codec
func NewPayloadDecoder(codec *PayloadCodec) *PayloadDecoder {
	return &PayloadDecoder{codec: codec}
}

// DecodeText decodes a text payload from an envelope
func (pd *PayloadDecoder) DecodeText(envelope *Envelope) (*TextPayload, error) {
	var payload TextPayload
	if err := pd.codec.DecodePayload(envelope, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// DecodeSetName decodes a set_name payload from an envelope
func (pd *PayloadDecoder) DecodeSetName(envelope *Envelope) (*SetNamePayload, error) {
	var payload SetNamePayload
	if err := pd.codec.DecodePayload(envelope, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// DecodeChat decodes a chat payload from an envelope
func (pd *PayloadDecoder) DecodeChat(envelope *Envelope) (*ChatPayload, error) {
	var payload ChatPayload
	if err := pd.codec.DecodePayload(envelope, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// DecodeDirect decodes a direct message payload from an envelope
func (pd *PayloadDecoder) DecodeDirect(envelope *Envelope) (*DirectPayload, error) {
	var payload DirectPayload
	if err := pd.codec.DecodePayload(envelope, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// DecodeCommand decodes a command payload from an envelope
func (pd *PayloadDecoder) DecodeCommand(envelope *Envelope) (*CommandPayload, error) {
	var payload CommandPayload
	if err := pd.codec.DecodePayload(envelope, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// DecodePing decodes a ping payload from an envelope
func (pd *PayloadDecoder) DecodePing(envelope *Envelope) (*PingPayload, error) {
	var payload PingPayload
	if err := pd.codec.DecodePayload(envelope, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// PayloadEncoder provides a convenient interface for encoding payloads
type PayloadEncoder struct {
	codec *PayloadCodec
}

// NewPayloadEncoder creates a new payload encoder using the provided codec
func NewPayloadEncoder(codec *PayloadCodec) *PayloadEncoder {
	return &PayloadEncoder{codec: codec}
}

// EncodeText creates an envelope with a text payload
func (pe *PayloadEncoder) EncodeText(text string) (*Envelope, error) {
	return pe.codec.CreateEnvelope(MsgText, TextPayload{Text: text})
}

// EncodeChat creates an envelope with a chat payload
func (pe *PayloadEncoder) EncodeChat(content string) (*Envelope, error) {
	return pe.codec.CreateEnvelope("chat", ChatPayload{Content: content})
}

// EncodeSetName creates an envelope with a set_name payload
func (pe *PayloadEncoder) EncodeSetName(name string) (*Envelope, error) {
	return pe.codec.CreateEnvelope("set_name", SetNamePayload{Name: name})
}

// EncodeCommand creates an envelope with a command payload
func (pe *PayloadEncoder) EncodeCommand(raw string) (*Envelope, error) {
	return pe.codec.CreateEnvelope(MsgCommand, CommandPayload{Raw: raw})
}

// EncodeDirect creates an envelope with a direct message payload
func (pe *PayloadEncoder) EncodeDirect(to []string, content string) (*Envelope, error) {
	return pe.codec.CreateEnvelope("direct", DirectPayload{To: to, Content: content})
}

// EncodeAck creates an envelope with an ack payload
func (pe *PayloadEncoder) EncodeAck(status string) (*Envelope, error) {
	return pe.codec.CreateEnvelope(MsgAck, AckPayload{Status: status})
}

// EncodePong creates an envelope with a pong payload
func (pe *PayloadEncoder) EncodePong(seq int64) (*Envelope, error) {
	return pe.codec.CreateEnvelope(MsgPong, PongPayload{Seq: seq})
}

// Global instance for convenience (backward compatibility)
var (
	DefaultPayloadCodec   = NewPayloadCodec()
	DefaultPayloadDecoder = NewPayloadDecoder(DefaultPayloadCodec)
	DefaultPayloadEncoder = NewPayloadEncoder(DefaultPayloadCodec)
)

// Convenience functions for backward compatibility
func MustJSON(v interface{}) json.RawMessage {
	return DefaultPayloadCodec.MustEncodePayload(v)
}