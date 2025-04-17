package main

import server2 "github.com/xzwsloser/Go-redis/net/server"

func main() {
	server := server2.NewTcpServer()
	server.Run()
}
