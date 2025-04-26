package database

import (
	"errors"
	"github.com/xzwsloser/Go-redis/datastruct/dict"
	"github.com/xzwsloser/Go-redis/datastruct/lock"
	"github.com/xzwsloser/Go-redis/interface/database"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/lib/logger"
	"github.com/xzwsloser/Go-redis/lib/timeheap"
	"github.com/xzwsloser/Go-redis/lib/utils"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"strings"
	"time"
)

const (
	DEFAULT_HASH_BUCKETS  = 16
	DEFAULT_LOCK_KEYS     = 32
	DEFAULT_TICK_INTERVAL = time.Millisecond * 100
	EXPIRE_PREFIX         = "expire:"
)

type CmdLine = [][]byte

// Database is the inner memory database of redis
type Database struct {
	index int
	data  dict.Dict
	// key(string) -> expireTime(time.Time)
	ttlMap dict.Dict
	// key(string) -> versionCode
	versionMap dict.Dict
	lockMap    *lock.Locks
	addAof     func(cmdLine [][]byte)
	timeHeap   *timeheap.TimeHeap
}

func NewDatabase(idx int) *Database {
	db := &Database{
		data:       dict.NewConcurrentDict(DEFAULT_HASH_BUCKETS),
		ttlMap:     dict.NewConcurrentDict(DEFAULT_HASH_BUCKETS),
		versionMap: dict.NewConcurrentDict(DEFAULT_HASH_BUCKETS),
		lockMap:    lock.NewLocks(DEFAULT_LOCK_KEYS),
		addAof:     func(cmdLine [][]byte) {},
		index:      idx,
		timeHeap:   timeheap.NewTimeHeap(DEFAULT_TICK_INTERVAL),
	}
	db.timeHeap.Start()
	return db
}

// ExecFunc the core method to invoke by the command
type ExecFunc func(db *Database, cmdLine [][]byte) redis.Reply

// PreFunc get the key of the command and lock the keys,return write keys and read keys
type PreFunc func(cmdLine [][]byte) ([]string, []string)

// UndoFunc get the undo logs of the current command
type UndoFunc func(db *Database, args [][]byte) []CmdLine

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

func (db *Database) RWLocks(wks []string, rks []string) {
	db.lockMap.RWLocks(wks, rks)
}

func (db *Database) RWUnlocks(wks []string, rks []string) {
	db.lockMap.RWUnlocks(wks, rks)
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

func (db *Database) TTLCmd(key string) [][]byte {
	expireKey := EXPIRE_PREFIX + key
	value, exists := db.ttlMap.GetWithLock(expireKey)
	if !exists {
		return nil
	}
	expireAt := time.UnixMilli(value.(int64))
	return utils.ExpireCmd(key, expireAt)
}

func (db *Database) Expire(key string, expireAt time.Time) redis.Reply {
	expireKey := EXPIRE_PREFIX + key
	db.LockSingleKey(expireKey)
	defer db.UnlockSingleKey(expireKey)
	if _, ok := db.ttlMap.GetWithLock(expireKey); ok {
		db.timeHeap.RemoveTask(expireKey)
	}
	db.ttlMap.PutWithLock(expireKey, expireAt)
	db.timeHeap.AddTask(expireAt, expireKey, func() {
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

func (db *Database) GetVersion(key string) uint32 {
	versionCode, exists := db.versionMap.GetWithLock(key)
	if !exists {
		return 0
	}
	return versionCode.(uint32)
}

func (db *Database) AddVersion(keys ...string) {
	if keys == nil {
		return
	}

	for _, key := range keys {
		versionCode := db.GetVersion(key)
		db.versionMap.PutWithLock(key, versionCode+1)
	}
}

func (db *Database) Exec(conn redis.Conn, cmdLine [][]byte) redis.Reply {
	cmdName := strings.ToLower(string(cmdLine[0]))
	if cmdName == "multi" {
		if len(cmdLine) != 1 {
			return protocol.NewErrReply("args of multi err")
		}
		return StartMulti(conn)
	} else if cmdName == "exec" {
		if len(cmdLine) != 1 {
			return protocol.NewErrReply("args of the exec err")
		}
		return ExecMulti(db, conn)
	} else if cmdName == "watch" {
		if len(cmdLine) < 2 {
			return protocol.NewErrReply("args of the watch err")
		}
		return Watch(db, conn, cmdLine[1:])
	} else if cmdName == "discard" {
		if len(cmdLine) != 1 {
			return protocol.NewErrReply("args of discard err")
		}
		return DiscardMulti(db, conn)
	}

	if conn != nil && conn.InitMulti() {
		return EnqueueCmd(conn, cmdLine)
	}

	if validCommand(cmdLine) != nil {
		return protocol.NewErrReply("in valid command")
	}

	return db.execNormalCommand(conn, cmdLine)
}

func (db *Database) execNormalCommand(conn redis.Conn, cmdLine [][]byte) redis.Reply {
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmd, ok := commandTable[cmdName]
	if !ok {
		return protocol.NewErrReply(COMMAND_NOT_FIND)
	}

	prepare := cmd.prepare
	if prepare != nil {
		wks, rks := prepare(cmdLine[1:])
		db.RWLocks(wks, rks)
		defer db.RWUnlocks(wks, rks)
		db.AddVersion(wks...)
	}

	reply := cmd.exector(db, cmdLine[1:])
	if reply == nil {
		return protocol.NewErrReply(EMPTY_REPLY)
	}
	return reply
}

func (db *Database) execWithLock(conn redis.Conn, cmdLine [][]byte) redis.Reply {
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmd, ok := commandTable[cmdName]
	if !ok {
		return protocol.NewErrReply(COMMAND_NOT_FIND)
	}

	reply := cmd.exector(db, cmdLine[1:])
	if reply == nil {
		return protocol.NewErrReply(EMPTY_REPLY)
	}
	return reply
}

func validCommand(commandLine [][]byte) error {
	commandName := strings.ToLower(string(commandLine[0]))
	commandInfo, exists := commandTable[commandName]
	if !exists {
		return errors.New(COMMAND_NOT_FIND)
	}

	if commandInfo.arity < 0 {
		if len(commandLine) < -commandInfo.arity {
			return errors.New(ARGS_OF_COMMAND_ERR)
		}
	} else {
		if len(commandLine) != commandInfo.arity {
			return errors.New(ARGS_OF_COMMAND_ERR)
		}
	}
	return nil
}
