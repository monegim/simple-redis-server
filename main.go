package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatal(err)
	}
	//defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go handler(conn)
	}
}

func handler(conn net.Conn) {
	fmt.Println("new connection from", conn.RemoteAddr())
}
