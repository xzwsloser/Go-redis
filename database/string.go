package database

import (
	"github.com/xzwsloser/Go-redis/interface/redis"
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

func (db *Database) getAsString(key string) (value string, exists bool) {
	val, exists := db.GetEntity(key)
	if !exists {
		return "", false
	}
	value, ok := val.(string)
	if !ok {
		return "", false
	}
	return
}

func (db *Database) getAsStringWithLock(key string) (value string, exists bool) {
	val, exists := db.GetEntityWithLock(key)
	if !exists {
		return "", false
	}
	value, ok := val.(string)
	if !ok {
		return "", false
	}
	return
}

func (db *Database) getAsInt(key string) (value int, exists bool) {
	valStr, exists := db.GetEntity(key)
	if !exists {
		return 0, false
	}

	var result int
	exists = true
	switch valStr.(type) {
	case int:
		result = valStr.(int)
	case int32:
		result = int(valStr.(int32))
	case int64:
		result = int(valStr.(int64))
	case string:
		result, err := strconv.Atoi(valStr.(string))
		if err != nil {
			return result, false
		}
	default:
		exists = false
	}

	return result, exists
}

func (db *Database) getAsIntWithLock(key string) (value int, exists bool) {
	valStr, exists := db.GetEntityWithLock(key)
	if !exists {
		return 0, false
	}

	var result int
	exists = true
	switch valStr.(type) {
	case int:
		result = valStr.(int)
	case int32:
		result = int(valStr.(int32))
	case int64:
		result = int(valStr.(int64))
	case string:
		result, err := strconv.Atoi(valStr.(string))
		if err != nil {
			return result, false
		}
	default:
		exists = false
	}

	return result, exists
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
	value, exists := db.GetEntity(key)
	var valueStr string
	if !exists {
		return protocol.NewErrReply(KEY_NOT_EXISTS_WRAN)
	}

	switch value.(type) {
	case string:
		valueStr = value.(string)
	case int64:
		valueStr = strconv.FormatInt(value.(int64), 10)
	case int32:
		valueStr = strconv.FormatInt(int64(value.(int32)), 10)
	case int:
		valueStr = strconv.Itoa(value.(int))
	case float64:
		valueStr = strconv.FormatFloat(value.(float64), 'f', 2, 64)
	default:
		valueStr = ""
	}
	return protocol.NewBulkReply([]byte(valueStr))
}

// SET  eg SET "k1" "v1"
func execSet(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	value := string(cmdLine[1])
	result := db.PutEntity(key, value)
	return protocol.NewIntReply(int64(result))
}

// SETNX eg SETNX "k1" "v1"
func execSetNx(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	value := string(cmdLine[1])
	result := db.PutEntity(key, value)
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

	_ = db.PutEntityIfExistsWithLock(key, value)
	return protocol.NewBulkReply([]byte(result))
}

// DEL  eg DEL a1 a2 a3...
func execDel(db *Database, cmdLine [][]byte) redis.Reply {
	keys := bytesToString(cmdLine)
	for _, key := range keys {
		db.RemoveEntity(key)
	}
	return protocol.NewOkReply()
}

// Incr eg Incr a1
func execIncr(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	db.LockSingleKey(key)
	defer db.UnlockSingleKey(key)
	value, exists := db.getAsIntWithLock(key)
	if !exists {
		return protocol.NewErrReply(KEY_NOT_EXISTS_WRAN)
	}
	value++
	_ = db.PutEntityWithLock(key, value)
	return protocol.NewIntReply(1)
}

// DECR  e.g DECR v1
func execDecr(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	db.LockSingleKey(key)
	defer db.UnlockSingleKey(key)
	value, exists := db.getAsIntWithLock(key)
	if !exists {
		return protocol.NewErrReply(KEY_NOT_EXISTS_WRAN)
	}

	value--
	_ = db.PutEntityWithLock(key, value)
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
		resultMap[key] = value
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
		result += db.PutEntity(key, values[index])
	}

	return protocol.NewIntReply(int64(result))
}
