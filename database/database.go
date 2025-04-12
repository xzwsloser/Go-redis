package database

import (
	"github.com/xzwsloser/Go-redis/datastruct/dict"
	"github.com/xzwsloser/Go-redis/interface/redis"
)

// Database is the inner memory database of redis
type Database struct {
	index int
	data  dict.Dict
}

type ExecFunc func(db *Database, cmdLine [][]byte) redis.Reply
