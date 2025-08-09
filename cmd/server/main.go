package main

import (
	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/transport"
	"log"
)

func main() {
	hub := chat.NewHub()
	go hub.Run()

	go func() {
		if err := transport.StartTcp(":8080", hub); err != nil {
			log.Fatalf("tcp server exit: %v", err)
		}
	}()

	log.Println("chat-go server start, listen and serve at localhost:8080")
	select {}
}
