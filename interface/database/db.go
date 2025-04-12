package database

import "github.com/xzwsloser/Go-redis/interface/redis"

type CmdLine = [][]byte

type DB interface {
	Exec(conn redis.Conn, cmdLine [][]byte) redis.Reply
	Close()
	AfterClientClose(conn redis.Conn)
}
