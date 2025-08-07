package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/ws", handleWS)
	// 启动广播协程
	go broadcaster()

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("chat-go server start at listening port :8080")
	}
}
