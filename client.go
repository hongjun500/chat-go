package main

import "net"

type Client struct {
	Conn net.Conn
	Name string
}
