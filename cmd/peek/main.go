package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"os"
	"unicode/utf8"

	"github.com/hongjun500/chat-go/internal/protocol"
	"github.com/hongjun500/chat-go/internal/transport"
)

func parseCodec(s string) (int, error) {
	switch s {
	case "json", "JSON", "0":
		return protocol.CodecJson, nil
	case "protobuf", "pb", "PB", "1":
		return protocol.CodecProtobuf, nil
	default:
		return -1, fmt.Errorf("unknown codec: %s", s)
	}
}

func main() {
	var (
		addr   = flag.String("addr", "localhost:8080", "server address")
		codecS = flag.String("codec", "json", "codec: json|protobuf")
		max    = flag.Int("max", 1<<20, "max frame size in bytes")
	)
	flag.Parse()

	cc, err := parseCodec(*codecS)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	mc, err := protocol.NewCodec(cc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fc := transport.NewFrameCodec()
	for {
		data, err := fc.ReadFrame(conn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read frame error: %v\n", err)
			os.Exit(1)
		}

		var env protocol.Envelope
		if err := mc.Decode(bytes.NewReader(data), &env, *max); err != nil {
			fmt.Fprintf(os.Stderr, "decode envelope error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Envelope:\n")
		fmt.Printf("  type: %s\n", env.Type)
		fmt.Printf("  ts:   %d\n", env.Ts)
		if len(env.Data) == 0 {
			fmt.Printf("  data: <empty>\n")
		} else if utf8.Valid(env.Data) {
			fmt.Printf("  data(text): %s\n", string(env.Data))
		} else {
			s := base64.StdEncoding.EncodeToString(env.Data)
			if len(s) > 80 {
				s = s[:80] + "..."
			}
			fmt.Printf("  data(base64): %s\n", s)
		}
	}
}
