package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	fmt.Println("chat-go server start at listening port :8080")

	// 启动广播协程
	go broadcast()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		ClientsMux.Lock()
		Clients[conn] = &Client{
			Conn: conn,
			Name: "sys",
		}
		ClientsMux.Unlock()
		go handleConnection(conn)
	}
}
