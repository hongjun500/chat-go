package config

import (
	"os"
	"strconv"
)

type Config struct {
	TCPAddr   string
	OutBuffer int
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() *Config {
	addr := getEnv("CHAT_TCP_ADDR", ":8080")
	outBufStr := getEnv("CHAT_OUTBUF", "256")
	outBuf, err := strconv.Atoi(outBufStr)
	if err != nil || outBuf <= 0 {
		outBuf = 256
	}
	return &Config{
		TCPAddr:   addr,
		OutBuffer: outBuf,
	}
}
