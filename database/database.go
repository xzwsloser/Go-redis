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

type DataEntity struct {
	Data any
}

func NewDatabase() *Database {
	return &Database{
		data:    dict.NewConcurrentDict(DEFAULT_HASH_BUCKETS),
		lockMap: lock.NewLocks(DEFAULT_LOCK_KEYS),
	}
}

type ExecFunc func(db *Database, cmdLine [][]byte) redis.Reply

func (db *Database) GetEntity(key string) (entity *DataEntity, exists bool) {
	value, exists := db.data.Get(key)
	if !exists {
		return nil, false
	}

	entity, ok := value.(*DataEntity)
	if !ok {
		return nil, false
	}

	return entity, true
}

func (db *Database) GetEntityWithLock(key string) (entity *DataEntity, exists bool) {
	value, exists := db.data.GetWithLock(key)
	if !exists {
		return nil, false
	}

	entity, ok := value.(*DataEntity)
	if !ok {
		return nil, false
	}

	return entity, true
}

func (db *Database) PutEntity(key string, entity *DataEntity) (result int) {
	result = db.data.Put(key, entity)
	return
}

func (db *Database) PutEntityWithLock(key string, entity *DataEntity) (result int) {
	result = db.data.PutWithLock(key, entity)
	return
}

func (db *Database) PutEntityIfAbsent(key string, entity *DataEntity) (result int) {
	result = db.data.PutIfAbsent(key, entity)
	return
}

func (db *Database) PutEntityIfAbsentWithLock(key string, entity *DataEntity) (result int) {
	result = db.data.PutIfAbsentWithLock(key, entity)
	return
}

func (db *Database) PutEntityIfExists(key string, entity *DataEntity) (result int) {
	result = db.data.PutIfExists(key, entity)
	return
}

func (db *Database) PutEntityIfExistsWithLock(key string, entity *DataEntity) (result int) {
	result = db.data.PutIfExistsWithLock(key, entity)
	return
}

func (db *Database) RemoveEntity(key string) (entity *DataEntity, result int) {
	value, result := db.data.Remove(key)
	entity, ok := value.(*DataEntity)
	if !ok {
		return nil, 0
	}
	return
}

func (db *Database) RemoveEntityWithLock(key string) (entity *DataEntity, result int) {
	value, result := db.data.RemoveWithLock(key)
	entity, ok := value.(*DataEntity)
	if !ok {
		return nil, 0
	}
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
