package main

import "server"

func main() {

	var s server.TcpChatServer
	s = *server.NewTcpChatServer()
	s.Listen(":6666")
	s.Start()
}
