package database

import (
	"github.com/xzwsloser/Go-redis/interface/database"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/lib/utils"
	"github.com/xzwsloser/Go-redis/lib/wildcard"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"strconv"
	"time"
)

func init() {
	RegisterCommand("KEYS", execKeys, nil, nil, -2)
	RegisterCommand("DEL", execKeys, writeKeys, nil, -2)
	RegisterCommand("PERSISTER", execPersister, writeKeys, nil, -2)
	RegisterCommand("PEXPIREAT", execPExpireAt, writeFirstKey, nil, 3)
	RegisterCommand("EXPIREAT", execExpireAt, writeFirstKey, nil, 3)
	RegisterCommand("EXPIRE", execExpire, writeFirstKey, nil, 3)
	RegisterCommand("PEXPIRE", execPExpire, writeFirstKey, nil, 3)
}

// execKeys: keys *
func execKeys(db *Database, cmdLine [][]byte) redis.Reply {
	patternStr := string(cmdLine[0])
	pattern, err := wildcard.CompilePattern(patternStr)
	if err != nil {
		return protocol.NewErrReply("in valid pattern str")
	}
	keys := make([]string, 0)
	db.ForEach(func(key string, value *database.DataEntity) bool {
		if pattern.IsMatch(key) {
			keys = append(keys, key)
		}
		return true
	})

	args := make([][]byte, len(keys))
	for i, key := range keys {
		args[i] = []byte(strconv.Itoa(i) + ") " + key)
	}
	return protocol.NewMultiReply(args)
}

// execDel: DEL k1 , k2 , k3 ...
func execDel(db *Database, cmdLine [][]byte) redis.Reply {
	keysToDel := make([]string, len(cmdLine))
	for i, arg := range cmdLine {
		keysToDel[i] = string(arg)
	}
	var r int
	for _, key := range keysToDel {
		_, result := db.RemoveEntityWithLock(key)
		r += result
	}
	db.addAof(utils.CmdLine2("DEL", cmdLine))
	return protocol.NewIntReply(int64(r))
}

// execPersister: Persister k1 , k2 , k3
func execPersister(db *Database, cmdLine [][]byte) redis.Reply {
	keysToPersister := make([]string, len(cmdLine))
	for i, key := range cmdLine {
		keysToPersister[i] = string(key)
	}
	for _, key := range keysToPersister {
		db.Persister(key)
	}
	db.addAof(utils.CmdLine2("PERSISTER", cmdLine))
	return protocol.NewOkReply()
}

// execPExpireAt: PEXPIREAT key timestamp
func execPExpireAt(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	timeStamp, err := strconv.ParseInt(string(cmdLine[1]), 10, 64)
	if err != nil {
		return protocol.NewErrReply("in valid timestamp format")
	}
	expireAt := time.UnixMilli(timeStamp)
	db.Expire(key, expireAt)
	db.addAof(utils.CmdLine2("PEXPIREAT", cmdLine))
	return protocol.NewIntReply(1)
}

// execExpireAt: EXPIREAT key timestamp
func execExpireAt(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	timeStampStr := string(cmdLine[1])
	timeStamp, err := strconv.ParseInt(timeStampStr, 10, 64)
	if err != nil {
		return protocol.NewErrReply("in valid timestamp format")
	}
	expireAt := time.Unix(timeStamp, 0)
	db.Expire(key, expireAt)
	db.addAof(utils.ExpireCmd(key, expireAt))
	return protocol.NewIntReply(1)
}

// execExpire: EXPIRE key timeDuration
func execExpire(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	td, err := strconv.ParseInt(string(cmdLine[1]), 10, 64)
	if err != nil {
		return protocol.NewErrReply("in valid time duration format")
	}
	expireAt := time.Now().Add(time.Duration(td) * time.Second)
	db.Expire(key, expireAt)
	db.addAof(utils.ExpireCmd(key, expireAt))
	return protocol.NewIntReply(1)
}

// execPExpire: PEXPIRE key timestamp
func execPExpire(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	tdm, err := strconv.ParseInt(string(cmdLine[1]), 10, 64)
	if err != nil {
		return protocol.NewErrReply("in valid time duration format")
	}
	expireAt := time.Now().Add(time.Duration(tdm) * time.Millisecond)
	db.Expire(key, expireAt)
	db.addAof(utils.ExpireCmd(key, expireAt))
	return protocol.NewIntReply(1)
}
