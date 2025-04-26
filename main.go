package main

import (
	"github.com/xzwsloser/Go-redis/config"
	"github.com/xzwsloser/Go-redis/lib/logger"
	"github.com/xzwsloser/Go-redis/net/server"
)

func main() {
	logger.Info("redis server listen on: %s:%d", config.GetRedisServerConfig().Address, config.GetRedisServerConfig().Port)
	redisServer := server.NewTcpServer()
	redisServer.Run()
}
