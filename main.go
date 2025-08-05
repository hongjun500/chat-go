package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

var (
	clients    = make(map[net.Conn]string) // 所有客户端
	clientsMux sync.Mutex                  // 并发所
	messages   = make(chan string)         // 消息通道
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
		clientsMux.Lock()
		clients[conn] = conn.RemoteAddr().String()
		clientsMux.Unlock()
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		clientsMux.Lock()
		delete(clients, conn)
		clientsMux.Unlock()
		err := conn.Close()
		if err != nil {
			return
		}
	}()

	clientAddr := conn.RemoteAddr().String()
	fmt.Println("new client connection from:", clientAddr)
	welcome := fmt.Sprintf("welcome to the chat-go server at %s", clientAddr)
	messages <- welcome

	reader := bufio.NewScanner(conn)
	for reader.Scan() {
		msg := reader.Text()
		fmt.Printf("[%s] : %s\n", clientAddr, msg)
		messages <- msg
	}
}

func broadcast() {
	for msg := range messages {
		clientsMux.Lock()
		for conn := range clients {
			fmt.Fprintln(conn, msg) // 向每个连接输出消息
		}
		clientsMux.Unlock()
	}
}
