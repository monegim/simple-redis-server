package main

import (
	"bufio"
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
		returnString              []byte
		inputCommand              []string
	)
	for {
		message, err := reader.ReadString('\n')
		log.Println("message: ", message)
		if err != nil {
			client.conn.Close()
			log.Println(err)
			return
		}
		if strings.HasPrefix(message, "*") {
			counter = 0
			numberOfElements, err = strconv.Atoi(strings.TrimSpace(message[1:]))
			inputCommand = nil
			if err != nil {
				log.Fatal(err)
			}
		} else if counter <= 2*numberOfElements && counter%2 == 0 {
			inputCommand = append(inputCommand, strings.TrimSpace(message))
		}
		counter++
		//log.Println("counter:", counter)
		if len(inputCommand) == numberOfElements {
			returnString = ExecCommand(inputCommand)
			log.Println("input command:", inputCommand)
			log.Println("output:", string(returnString))
			client.conn.Write(returnString)
		}
	}
}
func ExecCommand(command []string) []byte {
	if len(command) == 1 && command[0] == "ping" {
		return []byte("+PONG\r\n")
	}
	result := "+OK\r\n"
	return []byte(result)
}
func main() {
	server := New(&Config{
		Host: "127.0.0.1",
		Port: "6379",
	})
	log.Printf("Starting server on %s:%s...\n", server.host, server.port)
	server.Run()
}
