package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
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
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			client.conn.Close()
			log.Println(err)
			return
		}
		mesType := GetType(message)
		fmt.Printf("message type: %s\n", mesType)
		client.conn.Write([]byte(fmt.Sprintf("message incomming: %s", message)))
	}
}

type respType string

const (
	RESP_SIMPLE_STRING respType = "RESP_SIMPLE_STRING"
	RESP_ERROR         respType = "RESP_ERROR"
	RESP_INTEGER       respType = "RESP_INTEGER"
	RESP_BULK_STRING   respType = "RESP_BULK_STRING"
	RESP_ARRAYS        respType = "RESP_ARRAYS"
)

func GetType(s string) respType {
	c := s[0]
	switch c {
	case '+':
		return RESP_SIMPLE_STRING
	case '-':
		return RESP_ERROR
	case ':':
		return RESP_INTEGER
	case '$':
		return RESP_BULK_STRING
	case '*':
		return RESP_ARRAYS
	default:
		return ""
	}
}

func main() {
	server := New(&Config{
		Host: "127.0.0.1",
		Port: "6379",
	})
	server.Run()

}
