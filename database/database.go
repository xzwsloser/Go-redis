package database

import (
	"github.com/xzwsloser/Go-redis/datastruct/dict"
	"github.com/xzwsloser/Go-redis/datastruct/lock"
	"github.com/xzwsloser/Go-redis/interface/database"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/lib/logger"
	"github.com/xzwsloser/Go-redis/lib/timeheap"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"time"
)

const (
	DEFAULT_HASH_BUCKETS  = 16
	DEFAULT_LOCK_KEYS     = 32
	DEFAULT_TICK_INTERVAL = time.Millisecond * 500
	EXPIRE_PREFIX         = "expire:"
)

// Database is the inner memory database of redis
type Database struct {
	index int
	data  dict.Dict
	// key(string) -> expireTime(time.Time)
	ttlMap   dict.Dict
	lockMap  *lock.Locks
	addAof   func(cmdLine [][]byte)
	timeHeap *timeheap.TimeHeap
}

func NewDatabase(idx int) *Database {
	db := &Database{
		data:     dict.NewConcurrentDict(DEFAULT_HASH_BUCKETS),
		ttlMap:   dict.NewConcurrentDict(DEFAULT_LOCK_KEYS),
		lockMap:  lock.NewLocks(DEFAULT_LOCK_KEYS),
		addAof:   func(cmdLine [][]byte) {},
		index:    idx,
		timeHeap: timeheap.NewTimeHeap(DEFAULT_TICK_INTERVAL),
	}
	db.timeHeap.Start()
	return db
}

type ExecFunc func(db *Database, cmdLine [][]byte) redis.Reply

func (db *Database) GetEntity(key string) (entity *database.DataEntity, exists bool) {
	value, exists := db.data.Get(key)
	if !exists {
		return nil, false
	}

	entity, ok := value.(*database.DataEntity)
	if !ok {
		return nil, false
	}

	return entity, true
}

func (db *Database) GetEntityWithLock(key string) (entity *database.DataEntity, exists bool) {
	value, exists := db.data.GetWithLock(key)
	if !exists {
		return nil, false
	}

	entity, ok := value.(*database.DataEntity)
	if !ok {
		return nil, false
	}

	return entity, true
}

func (db *Database) PutEntity(key string, entity *database.DataEntity) (result int) {
	result = db.data.Put(key, entity)
	return
}

func (db *Database) PutEntityWithLock(key string, entity *database.DataEntity) (result int) {
	result = db.data.PutWithLock(key, entity)
	return
}

func (db *Database) PutEntityIfAbsent(key string, entity *database.DataEntity) (result int) {
	result = db.data.PutIfAbsent(key, entity)
	return
}

func (db *Database) PutEntityIfAbsentWithLock(key string, entity *database.DataEntity) (result int) {
	result = db.data.PutIfAbsentWithLock(key, entity)
	return
}

func (db *Database) PutEntityIfExists(key string, entity *database.DataEntity) (result int) {
	result = db.data.PutIfExists(key, entity)
	return
}

func (db *Database) PutEntityIfExistsWithLock(key string, entity *database.DataEntity) (result int) {
	result = db.data.PutIfExistsWithLock(key, entity)
	return
}

func (db *Database) RemoveEntity(key string) (entity *database.DataEntity, result int) {
	value, result := db.data.Remove(key)
	entity, ok := value.(*database.DataEntity)
	if !ok {
		return nil, 0
	}
	return
}

func (db *Database) RemoveEntityWithLock(key string) (entity *database.DataEntity, result int) {
	value, result := db.data.RemoveWithLock(key)
	entity, ok := value.(*database.DataEntity)
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

func (db *Database) ForEach(consumer func(key string, value *database.DataEntity) bool) {
	db.data.ForEach(func(key string, value any) bool {
		entity, ok := value.(*database.DataEntity)
		if !ok {
			logger.Warn("failed to transfer data type")
			return true
		}
		return consumer(key, entity)
	})
}

func (db *Database) Persister(key string) {
	expireKey := EXPIRE_PREFIX + key
	db.ttlMap.Remove(expireKey)
	db.timeHeap.RemoveTask(expireKey)
}

func (db *Database) Expire(key string, expireTime time.Duration) redis.Reply {
	expireKey := EXPIRE_PREFIX + key
	db.LockSingleKey(expireKey)
	defer db.UnlockSingleKey(expireKey)
	expire := time.Now().Add(expireTime)
	if _, ok := db.ttlMap.GetWithLock(expireKey); ok {
		db.timeHeap.RemoveTask(expireKey)
	}
	db.ttlMap.PutWithLock(expireKey, expire)
	db.timeHeap.AddTask(expire, expireKey, func() {
		db.Locks([]string{key, expireKey})
		defer db.Unlocks([]string{key, expireKey})
		if value, ok := db.ttlMap.Get(expireKey); ok {
			timeToExpire := value.(time.Time)
			if time.Now().After(timeToExpire) {
				db.data.RemoveWithLock(key)
			}
		}
		db.ttlMap.RemoveWithLock(expireKey)
	})
	return protocol.NewOkReply()
}
