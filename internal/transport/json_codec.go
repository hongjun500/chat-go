package transport

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// JSONCodec 实现 Codec 接口
type JSONCodec struct{}

func (JSONCodec) ContentType() string { return ApplicationJson }

func (JSONCodec) Encode(w io.Writer, m *Envelope) error {
	enc := json.NewEncoder(w)
	return enc.Encode(m)
}

// Decode 从 r 中读取 JSON 数据并解码到 m。
// 如果 maxSize > 0，则会限制最大读取字节数，防止内存滥用。
// 解码成功后，会检查业务字段 m.Type 是否存在。
// 若检测到 JSON 开头字符不是 '{'，则返回错误。
// 该函数采用流式解码，适合处理长连接和大数据场景。
// Decode JSON reads a JSON payload from r and decodes it into m.
func (JSONCodec) Decode(r io.Reader, m *Envelope, maxSize int) error {
	// 若指定最大消息大小，则限制读取字节数
	if maxSize > 0 {
		r = io.LimitReader(r, int64(maxSize))
	}
	// 缓冲读取器，便于回退读取位置
	bufReader := bufio.NewReader(r)
	// 预读首字节，用于验证是否为 JSON 对象
	firstByte, err := bufReader.Peek(1)
	if err != nil {
		return fmt.Errorf("read first byte: %w", err)
	}
	if firstByte[0] != '{' {
		return fmt.Errorf("payload not object")
	}
	// 创建 JSON 解码器（基于流式读取，避免一次性加载全部数据）
	dec := json.NewDecoder(bufReader)
	if err := dec.Decode(m); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}

	if m.Type == "" {
		// Not strictly required, but helps catch malformed inputs
		return fmt.Errorf("missing field: type")
	}
	return nil
}
