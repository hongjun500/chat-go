package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

// Envelope 与服务端一致（简化必要字段）
  type Envelope struct {
  	Version    string `json:"version"`
  	Type       string `json:"type"`
  	Encoding   string `json:"encoding,omitempty"` // 建议填 "json"
  	Mid        string `json:"mid,omitempty"`
  	Correlation string `json:"correlation_id,omitempty"`
  	From       string `json:"from,omitempty"`
  	To         string `json:"to,omitempty"`
  	Ts         int64  `json:"ts"`
  	Data       []byte `json:"data,omitempty"` // Go 的 json 会自动 base64
  }

  func writeFrame(conn net.Conn, payload []byte) error {
  	var header [4]byte
  	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
  	if _, err := conn.Write(header[:]); err != nil {
  		return err
  	}
  	_, err := conn.Write(payload)
  	return err
  }

  func readFrame(conn net.Conn) ([]byte, error) {
  	var header [4]byte
  	if _, err := conn.Read(header[:]); err != nil {
  		return nil, err
  	}
  	n := int(binary.BigEndian.Uint32(header[:]))
  	buf := make([]byte, n)
  	_, err := ioReadFull(conn, buf)
  	return buf, err
  }

  // 兼容老 Go 版本的 ReadFull（避免额外 import）
  func ioReadFull(conn net.Conn, buf []byte) (int, error) {
  	read := 0
  	for read < len(buf) {
  		n, err := conn.Read(buf[read:])
  		if err != nil {
  			return read, err
  		}
  		read += n
  	}
  	return read, nil
  }

  func main() {
  	addr := "127.0.0.1:8080"
  	conn, err := net.Dial("tcp", addr)
  	if err != nil {
  		panic(err)
  	}
  	defer conn.Close()
  	fmt.Println("connected:", addr)

  	// 读取欢迎 Envelope
  	if buf, err := readFrame(conn); err == nil {
  		var env Envelope
  		if err := json.Unmarshal(buf, &env); err == nil {
  			fmt.Printf("welcome: type=%s ts=%d data=%q\n", env.Type, env.Ts,
  string(env.Data))
  		} else {
  			fmt.Println("decode welcome failed:", err)
  		}
  	} else {
  		fmt.Println("read welcome failed:", err)
  	}

  	// 发送一条 text 消息
  	env := Envelope{
  		Version:  "1.0",
  		Type:     "text",
  		Encoding: "json",
  		Mid:      uuid.NewString(),
  		From:     "go-client",
  		Ts:       time.Now().UnixMilli(),
  		Data:     []byte("Hello from Go client"),
  	}
  	var payload bytes.Buffer
  	if err := json.NewEncoder(&payload).Encode(&env); err != nil {
  		panic(err)
  	}
  	if err := writeFrame(conn, payload.Bytes()); err != nil {
  		panic(err)
  	}
  	fmt.Println("sent one text envelope")
  }