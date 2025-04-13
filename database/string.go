package database

import (
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/lib/utils"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"strconv"
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

	TODO:
	GETEX
	SETEX
	PSETEX
*/

const (
	KEY_NOT_EXISTS_WRAN   = "key not exists"
	KEY_TYPE_CANNOT_TRANS = "the key of type is not int"
	ARGS_NUMBER_ERR_WARN  = "args of the command is err"
)

func init() {
	RegisterCommand("get", execGet, 2)
	RegisterCommand("set", execSet, 3)
	RegisterCommand("setnx", execSetNx, 3)
	RegisterCommand("getset", execGetSet, 3)
	RegisterCommand("del", execDel, -2)
	RegisterCommand("incr", execIncr, 2)
	RegisterCommand("decr", execDecr, 2)
	RegisterCommand("slen", execSLen, 2)
	RegisterCommand("mget", execMGet, -2)
	RegisterCommand("mset", execMSet, -3)
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
	value, exists := db.getAsString(key)
	if !exists {
		return protocol.NewErrReply(KEY_NOT_EXISTS_WRAN)
	}
	return protocol.NewBulkReply(value)
}

// SET  eg SET "k1" "v1"
func execSet(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	value := cmdLine[1]
	result := db.PutEntity(key, &DataEntity{
		Data: value,
	})
	db.addAof(utils.CmdLine2("SET", cmdLine))
	return protocol.NewIntReply(int64(result))
}

// SETNX eg SETNX "k1" "v1"
func execSetNx(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	value := string(cmdLine[1])
	result := db.PutEntity(key, &DataEntity{
		Data: []byte(value),
	})
	db.addAof(utils.CmdLine2("SETNX", cmdLine))
	return protocol.NewIntReply(int64(result))
}

// GETSET  eg GetSet "k1" "v1"
func execGetSet(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	value := string(cmdLine[1])
	db.LockSingleKey(key)
	defer db.UnlockSingleKey(key)
	result, exists := db.getAsStringWithLock(key)
	if !exists {
		return protocol.NewErrReply(KEY_NOT_EXISTS_WRAN)
	}

	_ = db.PutEntityIfExistsWithLock(key, &DataEntity{
		Data: []byte(value),
	})
	db.addAof(utils.CmdLine2("SET", cmdLine))
	return protocol.NewBulkReply([]byte(result))
}

// DEL  eg DEL a1 a2 a3...
func execDel(db *Database, cmdLine [][]byte) redis.Reply {
	keys := bytesToString(cmdLine)
	for _, key := range keys {
		db.RemoveEntity(key)
	}
	db.addAof(utils.CmdLine2("DEL", cmdLine))
	return protocol.NewOkReply()
}

// Incr eg Incr a1
func execIncr(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	db.LockSingleKey(key)
	defer db.UnlockSingleKey(key)
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
	_ = db.PutEntity(key, &DataEntity{
		Data: []byte(valueStr),
	})
	db.addAof(utils.CmdLine2("Incr", cmdLine))
	return protocol.NewIntReply(1)
}

// DECR  e.g DECR v1
func execDecr(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	db.LockSingleKey(key)
	defer db.UnlockSingleKey(key)
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
	_ = db.PutEntity(key, &DataEntity{
		Data: []byte(valueStr),
	})
	db.addAof(utils.CmdLine2("Decr", cmdLine))
	return protocol.NewIntReply(1)
}

// LEN  e.g LEN a1
func execSLen(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	value, exists := db.getAsString(key)
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
		value, exists := db.getAsString(key)
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
		result += db.PutEntity(key, &DataEntity{
			Data: []byte(values[index]),
		})
	}
	db.addAof(utils.CmdLine2("MSET", cmdLine))
	return protocol.NewIntReply(int64(result))
}
