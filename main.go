package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

type Server struct {
	host string
	port string
}

type Client struct {
	conn net.Conn
}

type Config struct {
	Host string
	Port string
}

func New(config *Config) *Server {
	return &Server{host: config.Host, port: config.Port}
}

func (s *Server) Run() {
	listener, err := net.Listen("tcp", s.host+":"+s.port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		client := &Client{conn: conn}
		go client.handleRequest()
	}
}

func (client *Client) handleRequest() {
	reader := bufio.NewReader(client.conn)
	var (
		counter, numberOfElements int
		inputCommand              string
	)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			client.conn.Close()
			log.Println(err)
			return
		}
		if strings.HasPrefix(message, "*") {
			counter = 0
			inputCommand = ""
			lenMessage := len(message)
			numberOfElements, err = strconv.Atoi(message[1 : lenMessage-2])
			if err != nil {
				log.Fatal(err)
			}
		} else if counter <= 2*numberOfElements {
			inputCommand = inputCommand + message
		}
		counter++
		//resp := NewRESP(message)
		//responseCommand := resp.Command()
		fmt.Printf("message: %s", message)
		fmt.Printf("inputCommand: %s", inputCommand)
		client.conn.Write([]byte("OK"))
	}
}

func main() {
	server := New(&Config{
		Host: "127.0.0.1",
		Port: "6379",
	})
	log.Printf("Starting server on %s:%s...\n", server.host, server.port)
	server.Run()
}
