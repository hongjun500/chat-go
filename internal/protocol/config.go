package protocol

import "fmt"

// CodecConfig 编码器配置接口
type CodecConfig interface {
	// GetDefaultCodec 获取默认编码器类型
	GetDefaultCodec() int
	// SetDefaultCodec 设置默认编码器类型
	SetDefaultCodec(codecType int) error
	// GetCodecByName 根据名称获取编码器类型
	GetCodecByName(name string) (int, error)
	// IsCodecSupported 检查编码器是否支持
	IsCodecSupported(codecType int) bool
}

// DefaultCodecConfig 默认编码器配置实现
type DefaultCodecConfig struct {
	defaultCodec int
}

// NewDefaultCodecConfig 创建默认编码器配置
func NewDefaultCodecConfig(defaultCodec int) *DefaultCodecConfig {
	return &DefaultCodecConfig{
		defaultCodec: defaultCodec,
	}
}

// GetDefaultCodec 获取默认编码器类型
func (c *DefaultCodecConfig) GetDefaultCodec() int {
	return c.defaultCodec
}

// SetDefaultCodec 设置默认编码器类型
func (c *DefaultCodecConfig) SetDefaultCodec(codecType int) error {
	if !c.IsCodecSupported(codecType) {
		return fmt.Errorf("unsupported codec type: %d", codecType)
	}
	c.defaultCodec = codecType
	return nil
}

// GetCodecByName 根据名称获取编码器类型
func (c *DefaultCodecConfig) GetCodecByName(name string) (int, error) {
	switch name {
	case Json:
		return CodecJson, nil
	case Protobuf:
		return CodecProtobuf, nil
	case Msgpack:
		return CodecMsgpack, nil
	default:
		return -1, fmt.Errorf("unsupported codec name: %s", name)
	}
}

// IsCodecSupported 检查编码器是否支持
func (c *DefaultCodecConfig) IsCodecSupported(codecType int) bool {
	_, exists := CodecFactories[codecType]
	return exists
}

// CodecNameMapping 编码器名称到类型的映射
var CodecNameMapping = map[string]int{
	Json:     CodecJson,
	Protobuf: CodecProtobuf,
	Msgpack:  CodecMsgpack,
}

// CodecTypeMapping 编码器类型到名称的映射
var CodecTypeMapping = map[int]string{
	CodecJson:     Json,
	CodecProtobuf: Protobuf,
	CodecMsgpack:  Msgpack,
}