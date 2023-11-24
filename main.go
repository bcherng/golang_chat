package main

import "server"

func main() {

	var s server.TcpChatServer
	s = *server.NewTcpChatServer()
	s.Listen(":3333")
	s.Start()
}
