package server

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type client struct {
	conn net.Conn
}

type TcpChatServer struct {
	listener net.Listener
	clients  map[*client]string
	mutex    *sync.Mutex
}

func NewTcpChatServer() *TcpChatServer {
	return &TcpChatServer{
		clients: make(map[*client]string),
		mutex:   &sync.Mutex{},
	}
}

func (s *TcpChatServer) Listen(address string) error {
	l, err := net.Listen("tcp", address)
	if err == nil {
		s.listener = l
	}
	log.Printf("Listening on %v", address)
	return err
}

func (s *TcpChatServer) Close() {
	s.listener.Close()
}

func (s *TcpChatServer) Start() {
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

func (s *TcpChatServer) Broadcast(message string) error {
	for client, _ := range s.clients {
		client.conn.Write([]byte(message))
	}
	return nil
}

func (s *TcpChatServer) Message(name string, message string) error {
	for client, nickname := range s.clients {
		if nickname == name {
			client.conn.Write([]byte(message))
		}
	}
	return nil
}

func (s *TcpChatServer) Nickname(client *client, name string) bool {
	for _, nickname := range s.clients {
		if nickname == name {
			return false
		}
	}
	s.clients[client] = name
	return true
}

func (s *TcpChatServer) List() []string {
	var names []string
	for _, name := range s.clients {
		names = append(names, name)
	}
	return names
}

func (s *TcpChatServer) accept(conn net.Conn) *client {
	log.Printf("Accepting connection from %v, total clients %v", conn.RemoteAddr().String(), len(s.clients))
	s.mutex.Lock()
	defer s.mutex.Unlock()
	client := &client{
		conn: conn,
	}
	return client
}

func (s *TcpChatServer) remove(client *client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for c, _ := range s.clients {
		if client == c {
			delete(s.clients, c)
			break
		}
	}
	log.Printf("Closing connection from %v, total clients %v", client.conn.RemoteAddr().String())
	client.conn.Close()
}

func (s *TcpChatServer) serve(client *client) {

	defer s.remove(client)
	for {
		buffer := make([]byte, 128)
		n, err := client.conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err)
			return
		}
		message := string(buffer[:n])
		message = strings.TrimSpace(message)
		parts := strings.SplitN(message, " ", 2)
		if len(parts) != 2 {
			client.conn.Write([]byte("Invalid Message format: " + message + "\n"))
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
