package pub

import (
	"github.com/xzwsloser/Go-redis/datastruct/dict"
	"github.com/xzwsloser/Go-redis/datastruct/lock"
)

type Hub struct {
	// subs: channel -> list(clients)
	subs dict.Dict
	// lockers: channel_name -> lock
	lockers *lock.Locks
}

func NewHub() *Hub {
	return &Hub{
		subs:    dict.NewConcurrentDict(16),
		lockers: lock.NewLocks(16),
	}
}
