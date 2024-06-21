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
	conn   net.Conn
	reader *bufio.Reader
}

type Store struct {
	data map[string]Fields
	lock sync.RWMutex
}
type Fields struct {
	Value  any
	Expiry time.Time
}
type Config struct {
	Host string
	Port string
}

// type Commands []string
//type ServerResponse []byte

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

func (c *Command) Ping() []byte {
	return ToClient("PONG", RESP_SIMPLE_STRING)
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
		go client.serve()
	}
}

var (
	err   error
	store = Store{data: map[string]Fields{}}
)

func (client *Client) serve() {
	client.reader = bufio.NewReader(client.conn)
	for {
		cmd, err := client.readCommand()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cmd:", cmd)
		executed := cmd.ExecCommand()
		client.conn.Write([]byte(executed))
	}
}

//	func (c *Commands) Clear() {
//		*c = nil
//	}
//
//	func (c *Commands) AddCommand(command string) {
//		*c = append(*c, command)
//	}
func (c *Command) Length() int {
	return len(c.Args)
}
func (c *Command) Nil() []byte {
	return ToClient("-1", RESP_BULK_STRING)
}
func (c *Command) Echo() []byte {
	return ToClient(strings.Join(c.Args, " "), RESP_SIMPLE_STRING)
}

func (c *Command) OK() []byte {
	return ToClient("OK", RESP_SIMPLE_STRING)
}

func (c *Command) Set() error {
	//TODO: Support for EX, NX, XX
	key := c.Args[0]
	value := c.Args[1]
	store.lock.Lock()
	defer store.lock.Unlock()
	store.data[key] = Fields{
		Value: value,
	}
	return nil
}

func (c *Command) GET() ([]byte, error) {
	key := c.Args[0]
	log.Println("store:", store)
	store.lock.RLock()
	defer store.lock.RUnlock()
	value, ok := store.data[key]
	if !ok {
		return c.Nil(), nil
	}
	return ToClient(value.Value.(string), RESP_SIMPLE_STRING), nil
}
func (c *Command) ExecCommand() []byte {
	//if c.Length() == 0 {
	//	return c.Nil()
	//}
	switch c.Name {
	case "ping":
		return c.Ping()
	case "echo":
		return c.Echo()
	case "set":
		err = c.Set()
		if err != nil {
			return ToClient("Notset", RESP_ERROR)
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

type Command struct {
	Name string
	Args []string
}

func (client *Client) readCommand() (*Command, error) {
	var numberOfElements int
	for {
		line, err := client.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(line, "*") {
			numberOfElements, err = strconv.Atoi(strings.TrimSpace(line[1:]))
			command := &Command{}
			for i := 0; i < 2*numberOfElements; i++ {
				line, err := client.reader.ReadString('\n')
				if err != nil {
					return nil, err
				}
				if i == 1 {
					command.Name = strings.TrimSuffix(line, "\r\n")
				}
				if i%2 == 1 && i > 2 {
					command.Args = append(command.Args, strings.TrimSpace(line))
				}
			}

			return command, nil
		}
	}
}

//func (client *Client) readLine() (string, error) {
//	var line string
//	for {
//		partialLine, isPrefix, err := client.reader.ReadLine()
//		if err != nil {
//			return "", nil
//		}
//		line += string(partialLine)
//		if isPrefix {
//			continue
//		}
//		return line, nil
//	}
//}

func main() {
	server := New(&Config{
		Host: "127.0.0.1",
		Port: "6379",
	})
	log.Printf("Starting server on %s:%s...\n", server.host, server.port)
	server.Run()
}
