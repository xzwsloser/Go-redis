package database

import (
	"github.com/xzwsloser/Go-redis/datastruct/dict"
	"github.com/xzwsloser/Go-redis/datastruct/lock"
	"github.com/xzwsloser/Go-redis/interface/redis"
)

const (
	DEFAULT_HASH_BUCKETS = 16
	DEFAULT_LOCK_KEYS    = 32
)

// Database is the inner memory database of redis
type Database struct {
	index   int
	data    dict.Dict
	lockMap *lock.Locks
}

func NewDatabase() *Database {
	return &Database{
		data:    dict.NewConcurrentDict(DEFAULT_HASH_BUCKETS),
		lockMap: lock.NewLocks(DEFAULT_LOCK_KEYS),
	}
}

type ExecFunc func(db *Database, cmdLine [][]byte) redis.Reply

func (db *Database) GetEntity(key string) (value any, exists bool) {
	value, exists = db.data.Get(key)
	return
}

func (db *Database) GetEntityWithLock(key string) (value any, exists bool) {
	value, exists = db.data.GetWithLock(key)
	return
}

func (db *Database) PutEntity(key string, value any) (result int) {
	result = db.data.Put(key, value)
	return
}

func (db *Database) PutEntityWithLock(key string, value any) (result int) {
	result = db.data.PutWithLock(key, value)
	return
}

func (db *Database) PutEntityIfAbsent(key string, value any) (result int) {
	result = db.data.PutIfAbsent(key, value)
	return
}

func (db *Database) PutEntityIfAbsentWithLock(key string, value any) (result int) {
	result = db.data.PutIfAbsentWithLock(key, value)
	return
}

func (db *Database) PutEntityIfExists(key string, value any) (result int) {
	result = db.data.PutIfExists(key, value)
	return
}

func (db *Database) PutEntityIfExistsWithLock(key string, value any) (result int) {
	result = db.data.PutIfExistsWithLock(key, value)
	return
}

func (db *Database) RemoveEntity(key string) (value any, result int) {
	value, result = db.data.Remove(key)
	return
}

func (db *Database) RemoveEntityWithLock(key string) (value any, result int) {
	value, result = db.data.RemoveWithLock(key)
	return
}

func (db *Database) LockSingleKey(key string) {
	db.lockMap.Locks([]string{key})
}

func (db *Database) Locks(keys []string) {
	db.lockMap.Locks(keys)
}

func (db *Database) UnlockSingleKey(key string) {
	db.lockMap.Unlocks([]string{key})
}

func (db *Database) Unlocks(keys []string) {
	db.lockMap.Unlocks(keys)
}
