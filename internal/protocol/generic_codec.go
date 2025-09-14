package protocol

import (
	"fmt"
	"io"
)

// GenericCodec 通用编码器实现，减少重复代码
type GenericCodec struct {
	name       string
	encoder    func(w io.Writer, e *Envelope) error
	decoder    func(r io.Reader, e *Envelope, maxSize int) error
	validator  func(e *Envelope) error
}

// NewGenericCodec 创建通用编码器
func NewGenericCodec(name string, encoder func(w io.Writer, e *Envelope) error, decoder func(r io.Reader, e *Envelope, maxSize int) error) *GenericCodec {
	return &GenericCodec{
		name:    name,
		encoder: encoder,
		decoder: decoder,
		validator: func(e *Envelope) error {
			if e == nil {
				return fmt.Errorf("envelope is nil")
			}
			if e.Type == "" {
				return fmt.Errorf("envelope missing required field 'type'")
			}
			return nil
		},
	}
}

// Name 返回编码器名称
func (g *GenericCodec) Name() string {
	return g.name
}

// Encode 编码消息
func (g *GenericCodec) Encode(w io.Writer, e *Envelope) error {
	if w == nil {
		return fmt.Errorf("%s.Encode: writer is nil", g.name)
	}
	if err := g.validator(e); err != nil {
		return fmt.Errorf("%s.Encode: %w", g.name, err)
	}
	
	if err := g.encoder(w, e); err != nil {
		return fmt.Errorf("%s.Encode: failed to encode envelope (Type=%s, Mid=%s): %w", g.name, e.Type, e.Mid, err)
	}
	return nil
}

// Decode 解码消息
func (g *GenericCodec) Decode(r io.Reader, e *Envelope, maxSize int) error {
	if r == nil {
		return fmt.Errorf("%s.Decode: reader is nil", g.name)
	}
	if e == nil {
		return fmt.Errorf("%s.Decode: envelope is nil", g.name)
	}
	
	if err := g.decoder(r, e, maxSize); err != nil {
		return fmt.Errorf("%s.Decode: %w", g.name, err)
	}
	
	if err := g.validator(e); err != nil {
		return fmt.Errorf("%s.Decode: decoded envelope invalid: %w", g.name, err)
	}
	
	return nil
}

// SetValidator 设置自定义验证器
func (g *GenericCodec) SetValidator(validator func(e *Envelope) error) {
	if validator != nil {
		g.validator = validator
	}
}

// CodecWrapper 编码器包装器，用于将现有编码器适配为GenericCodec
type CodecWrapper struct {
	codec MessageCodec
}

// NewCodecWrapper 创建编码器包装器
func NewCodecWrapper(codec MessageCodec) *CodecWrapper {
	return &CodecWrapper{codec: codec}
}

// Name 返回包装的编码器名称
func (w *CodecWrapper) Name() string {
	return w.codec.Name()
}

// Encode 包装编码方法
func (w *CodecWrapper) Encode(writer io.Writer, e *Envelope) error {
	if writer == nil {
		return fmt.Errorf("%s.Encode: writer is nil", w.codec.Name())
	}
	if e == nil {
		return fmt.Errorf("%s.Encode: envelope is nil", w.codec.Name())
	}
	
	if err := w.codec.Encode(writer, e); err != nil {
		return fmt.Errorf("%s.Encode: %w", w.codec.Name(), err)
	}
	return nil
}

// Decode 包装解码方法
func (w *CodecWrapper) Decode(reader io.Reader, e *Envelope, maxSize int) error {
	if reader == nil {
		return fmt.Errorf("%s.Decode: reader is nil", w.codec.Name())
	}
	if e == nil {
		return fmt.Errorf("%s.Decode: envelope is nil", w.codec.Name())
	}
	
	if err := w.codec.Decode(reader, e, maxSize); err != nil {
		return fmt.Errorf("%s.Decode: %w", w.codec.Name(), err)
	}
	return nil
}