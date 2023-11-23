package server

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type client struct {
	conn   net.Conn
	writer *bufio.Writer
	reader *bufio.Reader
}

type tcpChatServer struct {
	listener net.Listener
	clients  map[*client]string
	mutex    *sync.Mutex
}

func NewTcpChatServer() *tcpChatServer {
	return &tcpChatServer{
		clients: make(map[*client]string),
		mutex:   &sync.Mutex{},
	}
}

func (s *tcpChatServer) Listen(address string) error {
	l, err := net.Listen("tcp", address)
	if err == nil {
		s.listener = l
	}
	log.Printf("Listening on %v", address)
	return err
}

func (s *tcpChatServer) Close() {
	s.listener.Close()
}

func (s *tcpChatServer) Start() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Print(err)
		} else {
			client := s.accept(conn)
			go s.serve(client)
		}
	}
}

func (s *tcpChatServer) Broadcast(message string) error {
	for _, client := range s.clients {
		client.writer.WriteString(message)
	}
	return nil
}

func (s *tcpChatServer) Message(name string, message string) error {
	for _, client := range s.clients {
		if client.name == name {
			client.writer.WriteString(message)
		}
	}
	return nil
}

func (s *tcpChatServer) Nickname(name string) bool {
	for _, client := range s.clients {
		if client.name == name {
			return false
		}
	}
	s.clients = append(s.clients)
	return true
}

func (s *tcpChatServer) List() []string {
	var names []string
	for _, client := range s.clients {
		names = append(names, client.name)
	}
	return names
}

func (s *tcpChatServer) accept(conn net.Conn) *client {
	log.Printf("Accepting connection from %v, total clients %v", conn.RemoteAddr().String(), len(s.clients))
	s.mutex.Lock()
	defer s.mutex.Unlock()
	client := &client{
		conn:   conn,
		writer: *bufio.NewWriter(conn),
		reader: *bufio.NewReader(conn),
	}
	return client
}

func (s *tcpChatServer) remove(client *client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for i, check := range s.clients {
		if check == client {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
		}
	}
	log.Printf("Closing connection from %v, total clients %v", client.conn.RemoteAddr().String())
	client.conn.Close()
}

func (s *tcpChatServer) serve(client *client) {

	client.reader = bufio.Reader(client.conn)
	client.writer = bufio.Writer(client.conn)
	defer s.remove(client)
	for {
		message, error := client.reader.Read()
		if error != nil && error != io.EOF {
			log.Printf("Read Error %v", err)
		}
		if message != nil {
			message = strings.TrimSpace(message)
			parts := strings.SplitN(message, " ", 2)
			if len(parts) != 2 {
				client.writer.write("Invalid Message format %s\n", message)
				continue
			}
			command := parts[0]
			content := parts[1]

			switch command {
			case "/BC":
				go s.Broadcast(content)
			case "/MSG":
				go s.Message(content)
			case "/NICK":
				go s.Nickname(content)
			default:

			}
		}
	}
}
func main() {
	var s server.ChatServer
	s = server.NewServer()
	s.Listen(":3333")
	s.Start()
}
