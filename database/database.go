package database

import (
	"github.com/xzwsloser/Go-redis/datastruct/dict"
	"github.com/xzwsloser/Go-redis/interface/redis"
)

const (
	DEFAULT_HASH_BUCKETS = 16
)

// Database is the inner memory database of redis
type Database struct {
	index int
	data  dict.Dict
}

func NewDatabase() *Database {
	return &Database{
		data: dict.NewConcurrentDict(DEFAULT_HASH_BUCKETS),
	}
}

type ExecFunc func(db *Database, cmdLine [][]byte) redis.Reply
