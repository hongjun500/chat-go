package config

import (
	"os"
	"strconv"
)

type Config struct {
	TCPAddr   string
	OutBuffer int
	WSAddr    string
	HTTPAddr  string
	LogLevel  string
	// TCP advanced
	TCPCodec     string // json|protobuf
	WSCodec      string // json|protobuf
	ReadTimeout  int    // seconds
	WriteTimeout int    // seconds
	MaxFrameSize int    // bytes
	// Redis Stream
	RedisAddr   string
	RedisDB     int
	RedisStream string
	RedisGroup  string
	RedisEnable bool
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
	wsAddr := getEnv("CHAT_WS_ADDR", ":8081")
	httpAddr := getEnv("CHAT_HTTP_ADDR", ":8082")
	logLevel := getEnv("CHAT_LOG_LEVEL", "info")
	tcpCodec := getEnv("CHAT_TCP_CODEC", "json")
	wsCodec := getEnv("CHAT_WS_CODEC", "json")
	rtStr := getEnv("CHAT_TCP_READ_TIMEOUT", "60")
	wtStr := getEnv("CHAT_TCP_WRITE_TIMEOUT", "15")
	mfsStr := getEnv("CHAT_TCP_MAX_FRAME", "1048576")
	rt, _ := strconv.Atoi(rtStr)
	wt, _ := strconv.Atoi(wtStr)
	mfs, _ := strconv.Atoi(mfsStr)
	redisAddr := getEnv("CHAT_REDIS_ADDR", "localhost:6379")
	redisDBStr := getEnv("CHAT_REDIS_DB", "0")
	redisDB, _ := strconv.Atoi(redisDBStr)
	redisStream := getEnv("CHAT_REDIS_STREAM", "chat_stream")
	redisGroup := getEnv("CHAT_REDIS_GROUP", "chat_group")
	redisEnable := getEnv("CHAT_REDIS_ENABLE", "true") == "true"

	return &Config{
		TCPAddr:      addr,
		OutBuffer:    outBuf,
		WSAddr:       wsAddr,
		HTTPAddr:     httpAddr,
		LogLevel:     logLevel,
		TCPCodec:     tcpCodec,
		WSCodec:      wsCodec,
		ReadTimeout:  rt,
		WriteTimeout: wt,
		MaxFrameSize: mfs,
		RedisAddr:    redisAddr,
		RedisDB:      redisDB,
		RedisStream:  redisStream,
		RedisGroup:   redisGroup,
		RedisEnable:  redisEnable,
	}
}
