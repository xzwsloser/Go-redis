package main

import "github.com/xzwsloser/Go-redis/net/server"

func main() {
	redisServer := server.NewTcpServer()
	redisServer.Run()
}
