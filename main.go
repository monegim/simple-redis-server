package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	host string
	port string
}

type Client struct {
	conn net.Conn
}

type Store struct {
	data map[string]Fields
	lock *sync.RWMutex
}
type Fields struct {
	Value  any
	Expiry time.Time
}
type Config struct {
	Host string
	Port string
}

type Commands []string
type ServerResponse []byte

func ToClient(s string, t RespType) []byte {
	switch t {
	case RESP_SIMPLE_STRING:
		return []byte("+" + s + "\r\n")
	case RESP_ERROR:
		return []byte("-" + s + "\r\n")
	default:
		return []byte("$-1\r\n")
	}
}
func (c *Commands) Ping() []byte {
	return ToClient("ping", RESP_SIMPLE_STRING)
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

var (
	err                       error
	store                     = Store{data: map[string]Fields{}}
	counter, numberOfElements int
	returnString              []byte
	commands                  Commands
)

func (client *Client) handleRequest() {
	reader := bufio.NewReader(client.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			client.conn.Close()
			log.Println(err)
			return
		}
		if strings.HasPrefix(message, "*") {
			counter = 0
			numberOfElements, err = strconv.Atoi(strings.TrimSpace(message[1:]))
			commands.Clear()
			if err != nil {
				log.Fatal(err)
			}
		} else if counter <= 2*numberOfElements && counter%2 == 0 {
			log.Println("message:", message)
			commands.AddCommand(strings.TrimSuffix(message, "\r\n"))
			//inputCommand = append(inputCommand, )
		}
		counter++
		if commands.Length() == numberOfElements {
			returnString = commands.ExecCommand()
			client.conn.Write(returnString)
		}
	}
}
func (c *Commands) Clear() {
	*c = nil
}
func (c *Commands) AddCommand(command string) {
	*c = append(*c, command)
}
func (c *Commands) Length() int {
	return len(*c)
}

func (c *Commands) Nil() []byte {
	return []byte("$-1\r\n")
}

func (c *Commands) Echo() []byte {
	return []byte("+" + strings.Join((*c)[1:], " ") + "\r\n")
}

func (c *Commands) OK() []byte {
	return []byte("+OK\r\n")
}

func (c *Commands) Set() error {
	//TODO: Support for EX, NX, XX
	key := (*c)[1]
	value := (*c)[2]
	store.lock.Lock()
	defer store.lock.Unlock()
	store.data[key] = Fields{
		Value: value,
	}
	return nil
}

func (c *Commands) GET() ([]byte, error) {
	key := (*c)[1]
	store.lock.RLock()
	defer store.lock.RUnlock()
	value, ok := store.data[key]
	if !ok {
		return c.Nil(), nil
	}
	return ToClient(value.Value.(string), RESP_SIMPLE_STRING), nil
}
func (c *Commands) ExecCommand() []byte {
	if c.Length() == 0 {
		return c.Nil()
	}
	if c.Length() == 1 && (*c)[0] == "ping" {
		return c.Ping()
	}
	switch (*c)[0] {
	case "echo":
		return c.Echo()
	case "set":

		err = c.Set()
		if err != nil {
			return []byte("-Notset\r\n")
		}
		log.Printf("sore: %#v", store)
		return c.OK()
	case "get":
		v, err := c.GET()
		if err != nil {
			return ToClient(err.Error(), RESP_ERROR)
		}
		return v
	default:
		return c.Nil()
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
