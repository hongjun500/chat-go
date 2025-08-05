package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 输入昵称
	fmt.Print("请输入你的昵称：")
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')
	name = name[:len(name)-1] // 去掉换行

	// 启动后台协程接收消息
	go func() {
		serverReader := bufio.NewScanner(conn)
		for serverReader.Scan() {
			fmt.Println(serverReader.Text())
		}
	}()

	// 主线程循环发送消息
	for {
		text, _ := reader.ReadString('\n')
		text = text[:len(text)-1]
		fmt.Fprintf(conn, "%s\n", fmt.Sprintf("%s", text))
	}
}
