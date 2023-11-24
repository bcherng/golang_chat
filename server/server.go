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

func (s *TcpChatServer) Broadcast(client *client, message string) error {
	registered := false
	nickname := ""
	for c, n := range s.clients {
		if client == c {
			registered = true
			nickname = n
			break
		}
	}
	if registered != true {
		client.conn.Write([]byte("Register first!\n"))
	} else {
		client.conn.Write([]byte("Broadcasting: " + message + "\n"))
		for c := range s.clients {
			if c != client {
				c.conn.Write([]byte(nickname + ": " + message + "\n"))
			}
		}
	}
	return nil
}

func (s *TcpChatServer) Message(client *client, name string, message string) error {
	registered := false
	nickname := ""
	for c, n := range s.clients {
		if client == c {
			if name == n {
				client.conn.Write([]byte("Cannot message yourself!\n"))
				return nil
			} else {
				registered = true
				nickname = n
				break
			}
		}
	}
	if registered != true {
		client.conn.Write([]byte("Register first!\n"))
	} else {
		for c, n := range s.clients {
			if n == name {
				client.conn.Write([]byte("Messaging " + name + ": " + message + "\n"))
				c.conn.Write([]byte(nickname + ": " + message + "\n"))
			}
		}
	}
	return nil
}

func (s *TcpChatServer) Nickname(client *client, name string) bool {
	for _, nickname := range s.clients {
		if name == "server" || nickname == name {
			client.conn.Write([]byte("Name is taken!\n"))
			return false
		}
	}
	for c := range s.clients {
		if c != client {
			c.conn.Write([]byte("server: User " + name + " has logged on\n"))
		}
	}

	s.clients[client] = name
	client.conn.Write([]byte("Registered nickname " + name + "\n"))
	return true
}

func (s *TcpChatServer) List(client *client) {
	for _, name := range s.clients {
		client.conn.Write([]byte(name + " "))
	}
	client.conn.Write([]byte("\n"))
}

func (s *TcpChatServer) accept(conn net.Conn) *client {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	client := &client{
		conn: conn,
	}
	s.clients[client] = ""
	log.Printf("Accepting connection from %v, total clients %v", conn.RemoteAddr().String(), len(s.clients))
	return client
}

func (s *TcpChatServer) remove(client *client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for c, n := range s.clients {
		if client == c {
			if n != "" {
				for c := range s.clients {
					c.conn.Write([]byte("server: User " + n + " has logged on\n"))
				}
			}
			delete(s.clients, c)
			break
		}
	}
	log.Printf("Closing connection from %v, total clients %v", client.conn.RemoteAddr().String(), len(s.clients))
	client.conn.Close()
}

func invalidCommand(client *client) {
	client.conn.Write([]byte("Invalid command\n"))
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
		command := parts[0]
		switch len(parts) {
		case 1:
			if command == "/LIST" {
				s.List(client)
			} else {
				invalidCommand(client)
			}
			break
		case 2:
			params := strings.SplitN(parts[1], " ", 2)
			if command == "/NICK" && len(params) == 1 {
				name := params[0]
				s.Nickname(client, name)
			} else if command == "/BC" {
				message := params[0]
				s.Broadcast(client, message)
			} else if command == "/MSG" && len(params) == 2 {
				target := params[0]
				message := params[1]
				s.Message(client, target, message)
			} else {
				invalidCommand(client)
			}
		default:

		}
	}
}
