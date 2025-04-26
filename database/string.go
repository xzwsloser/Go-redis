package database

import (
	"github.com/xzwsloser/Go-redis/interface/database"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/lib/utils"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"strconv"
	"time"
)

/*
	GET
	SET
	SETNX
	MGET
	MSET
	GETSET
	DEL
	INCR
	DECR
	LEN
*/

const (
	KEY_NOT_EXISTS_WRAN   = "key not exists"
	KEY_TYPE_CANNOT_TRANS = "the key of type is not int"
	ARGS_NUMBER_ERR_WARN  = "args of the command is err"
)

func init() {
	RegisterCommand("GET", execGet, readFirstKey, nil, 2)
	RegisterCommand("SET", execSet, writeFirstKey, rollbackFirstKey, 3)
	RegisterCommand("SETNX", execSetNx, writeFirstKey, rollbackFirstKey, 3)
	RegisterCommand("GETSET", execGetSet, writeFirstKey, rollbackFirstKey, 3)
	RegisterCommand("INCR", execIncr, writeFirstKey, rollbackFirstKey, 2)
	RegisterCommand("DECR", execDecr, writeFirstKey, rollbackFirstKey, 2)
	RegisterCommand("SLEN", execSLen, readFirstKey, nil, 2)
	RegisterCommand("MGET", execMGet, readKeys, nil, -2)
	RegisterCommand("MSET", execMSet, prepareMSet, undoMSet, -3)
	RegisterCommand("SETEX", execSetEx, writeFirstKey, rollbackFirstKey, 4)
}

func (db *Database) getAsString(key string) (value []byte, exists bool) {
	entity, exists := db.GetEntity(key)
	if !exists {
		return nil, false
	}

	value, ok := entity.Data.([]byte)
	if !ok {
		return nil, true
	}
	return value, true
}

func (db *Database) getAsStringWithLock(key string) (value []byte, exists bool) {
	entity, exists := db.GetEntityWithLock(key)
	if !exists {
		return nil, false
	}

	value, ok := entity.Data.([]byte)
	if !ok {
		return nil, false
	}
	return value, true
}

func bytesToString(bytes [][]byte) (key []string) {
	key = make([]string, len(bytes))
	for i, v := range bytes {
		key[i] = string(v)
	}
	return key
}

func stringsToBytes(keys []string) (bytes [][]byte) {
	bytes = make([][]byte, len(keys))
	for i, key := range keys {
		bytes[i] = []byte(key)
	}
	return
}

// GET  eg GET "Hello"
func execGet(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	value, exists := db.getAsStringWithLock(key)
	if !exists {
		return protocol.NewErrReply(KEY_NOT_EXISTS_WRAN)
	}
	return protocol.NewBulkReply(value)
}

// SET  eg SET "k1" "v1"
func execSet(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	value := cmdLine[1]
	result := db.PutEntityWithLock(key, &database.DataEntity{
		Data: value,
	})
	db.addAof(utils.CmdLine2("SET", cmdLine))
	db.Persister(key)
	return protocol.NewIntReply(int64(result))
}

// SETNX eg SETNX "k1" "v1"
func execSetNx(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	value := string(cmdLine[1])
	result := db.PutEntityWithLock(key, &database.DataEntity{
		Data: []byte(value),
	})
	db.addAof(utils.CmdLine2("SETNX", cmdLine))
	db.Persister(key)
	return protocol.NewIntReply(int64(result))
}

// GETSET  eg GetSet "k1" "v1"
func execGetSet(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	value := string(cmdLine[1])
	result, exists := db.getAsStringWithLock(key)
	if !exists {
		return protocol.NewErrReply(KEY_NOT_EXISTS_WRAN)
	}

	_ = db.PutEntityIfExistsWithLock(key, &database.DataEntity{
		Data: []byte(value),
	})
	db.addAof(utils.CmdLine2("SET", cmdLine))
	return protocol.NewBulkReply([]byte(result))
}

// DEL  eg DEL a1 a2 a3...
//func execDel(db *Database, cmdLine [][]byte) redis.Reply {
//	keys := bytesToString(cmdLine)
//	for _, key := range keys {
//		db.RemoveEntityWithLock(key)
//	}
//	db.addAof(utils.CmdLine2("DEL", cmdLine))
//	return protocol.NewOkReply()
//}

// Incr eg Incr a1
func execIncr(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	valueBytes, exists := db.getAsStringWithLock(key)
	if !exists {
		return protocol.NewIntReply(0)
	}
	value, err := strconv.ParseInt(string(valueBytes), 10, 64)
	if err != nil {
		return protocol.NewIntReply(0)
	}
	value++
	valueStr := strconv.FormatInt(value, 10)
	_ = db.PutEntity(key, &database.DataEntity{
		Data: []byte(valueStr),
	})
	db.addAof(utils.CmdLine2("Incr", cmdLine))
	return protocol.NewIntReply(1)
}

// DECR  e.g DECR v1
func execDecr(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	valueBytes, exists := db.getAsStringWithLock(key)
	if !exists {
		return protocol.NewIntReply(0)
	}
	value, err := strconv.Atoi(string(valueBytes))
	if err != nil {
		return protocol.NewIntReply(0)
	}
	value--
	valueStr := strconv.Itoa(value)
	_ = db.PutEntity(key, &database.DataEntity{
		Data: []byte(valueStr),
	})
	db.addAof(utils.CmdLine2("Decr", cmdLine))
	return protocol.NewIntReply(1)
}

// LEN  e.g LEN a1
func execSLen(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	value, exists := db.getAsStringWithLock(key)
	if !exists {
		return protocol.NewErrReply(KEY_NOT_EXISTS_WRAN)
	}
	return protocol.NewIntReply(int64(len(value)))
}

// MGET  MGET k1 k2 k3
func execMGet(db *Database, cmdLine [][]byte) redis.Reply {
	keys := bytesToString(cmdLine)
	resultMap := make(map[string]string)
	for _, key := range keys {
		value, exists := db.getAsStringWithLock(key)
		if !exists {
			return protocol.NewErrReply(KEY_NOT_EXISTS_WRAN)
		}
		resultMap[key] = string(value)
	}

	results := make([]string, len(resultMap))
	i := 0
	for k, v := range resultMap {
		results[i] = k + ":" + v
		i++
	}
	bytes := stringsToBytes(results)
	return protocol.NewMultiReply(bytes)
}

func prepareMSet(args [][]byte) ([]string, []string) {
	if len(args)%2 == 1 {
		return nil, nil
	}
	wks := make([]string, len(args)/2)
	i := 0
	for j := 0; j < len(args); j += 2 {
		wks[i] = string(args[j])
		i++
	}
	return wks, nil
}

func undoMSet(db *Database, args [][]byte) []CmdLine {
	if len(args)%2 == 1 {
		return nil
	}
	keys := make([]string, len(args)/2)
	j := 0
	for i := 0; i < len(keys); i++ {
		keys[i] = string(args[j])
		j += 2
	}
	return rollbackGivenKeys(db, keys...)
}

// MSET MSET k1 v1 k2 v2 ...
func execMSet(db *Database, cmdLine [][]byte) redis.Reply {
	if len(cmdLine)%2 == 1 {
		return protocol.NewErrReply(ARGS_NUMBER_ERR_WARN)
	}

	keys := make([]string, len(cmdLine)/2)
	values := make([]string, len(cmdLine)/2)
	i := 0
	j := 0
	for index, value := range cmdLine {
		if index%2 == 0 {
			keys[i] = string(value)
			i++
		} else {
			values[j] = string(value)
			j++
		}
	}

	result := 0
	for index, key := range keys {
		result += db.PutEntityWithLock(key, &database.DataEntity{
			Data: []byte(values[index]),
		})
	}
	db.addAof(utils.CmdLine2("MSET", cmdLine))
	return protocol.NewIntReply(int64(result))
}

// setEx key value expireTime(s)
func execSetEx(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	value := cmdLine[1]
	expireTimeStr := string(cmdLine[2])
	expireTime, err := strconv.ParseInt(expireTimeStr, 10, 64)
	if err != nil {
		return protocol.NewErrReply("err expireTime format")
	}

	result := db.PutEntityWithLock(key, &database.DataEntity{
		Data: value,
	})

	if expireTime > 0 {
		timeout := time.Duration(expireTime) * time.Second
		expireAt := time.Now().Add(timeout)
		db.Expire(key, time.Now().Add(timeout))
		db.addAof(utils.ExpireCmd(key, expireAt))
	}
	return protocol.NewIntReply(int64(result))
}
