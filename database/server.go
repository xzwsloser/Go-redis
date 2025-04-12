package database

import (
	"github.com/xzwsloser/Go-redis/interface/redis"
	"sync/atomic"
)

// RedisServer is the inner server to exec command  like the httpServer
type RedisServer struct {
	dbSet []*atomic.Value
}

func (r *RedisServer) Exec(conn redis.Conn, cmdLine [][]byte) redis.Reply {
	//TODO implement me
	panic("implement me")
}

func (r *RedisServer) Close() {
	//TODO implement me
	panic("implement me")
}

func (r *RedisServer) AfterClientClose(conn redis.Conn) {
	//TODO implement me
	panic("implement me")
}
