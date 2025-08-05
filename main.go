package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	fmt.Println("chat-go server start at listening port :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}(conn)

	clientAddr := conn.RemoteAddr().String()
	fmt.Println("new client connection from:", clientAddr)
	reader := bufio.NewScanner(conn)
	for reader.Scan() {
		msg := reader.Text()
		fmt.Printf("[%s] : %s\n", clientAddr, msg)
	}
}
