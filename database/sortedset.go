package database

import (
	"github.com/xzwsloser/Go-redis/datastruct/sortedset"
	"github.com/xzwsloser/Go-redis/interface/database"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"strconv"
)

/**
@Author: loser
@Dscription: 实现 ZSET 中的命令
*/

func init() {
	RegisterCommand("ZADD", execZAdd, writeFirstKey, rollbackFirstKey, -4)
	RegisterCommand("ZCARD", execZCard, readFirstKey, nil, 2)
	RegisterCommand("ZCOUNT", execZCount, readFirstKey, nil, 4)
	RegisterCommand("ZINCRBY", execZIncrBy, writeFirstKey, rollbackFirstKey, 4)
	RegisterCommand("ZRANK", execZRank, readFirstKey, nil, 3)
	RegisterCommand("ZSCORE", execZScore, readFirstKey, nil, 3)
	RegisterCommand("ZRANGE", execZRange, readFirstKey, nil, 4)
	RegisterCommand("ZREM", execZRem, writeFirstKey, rollbackFirstKey, -3)
	RegisterCommand("ZRANGEBYSCORE", execZRangeByScore, readFirstKey, nil, 4)
	RegisterCommand("ZREMRANGEBYRANK", execZRemRangeByRank, writeFirstKey, rollbackFirstKey, 4)
}

func (db *Database) getOrInitSortedSet(key string) *sortedset.SortedSet {
	entity, exists := db.GetEntityWithLock(key)
	if !exists {
		ns := sortedset.NewSortedSet()
		db.PutEntityWithLock(key, &database.DataEntity{
			Data: ns,
		})
		return ns
	}
	ss, ok := entity.Data.(*sortedset.SortedSet)
	if !ok {
		ns := sortedset.NewSortedSet()
		db.PutEntityWithLock(key, &database.DataEntity{
			Data: ns,
		})
		return ns
	}
	return ss
}

// ZADD  e.g ZADD key score member ...
func execZAdd(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	pairLens := len(cmdLine) - 1
	if pairLens%2 == 1 {
		return protocol.NewErrReply("the args of command Err")
	}
	elements := make([]*sortedset.Element, pairLens/2)
	j := 0
	for i := 1; i < len(cmdLine); i += 2 {
		score, err := strconv.ParseFloat(string(cmdLine[i]), 64)
		if err != nil {
			return protocol.NewErrReply("the args of command Err")
		}
		elements[j] = &sortedset.Element{
			Member: string(cmdLine[i+1]),
			Score:  score,
		}
		j++
	}

	ss := db.getOrInitSortedSet(key)
	var result int
	for _, element := range elements {
		result += ss.Put(element.Member, element.Score)
	}
	return protocol.NewIntReply(int64(result))
}

// ZCARD key
func execZCard(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	ss := db.getOrInitSortedSet(key)
	return protocol.NewIntReply(ss.Len())
}

// ZCOUNT e.g ZCOUNT key min max
func execZCount(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	minValue, err := strconv.ParseFloat(string(cmdLine[1]), 64)
	maxValue, err := strconv.ParseFloat(string(cmdLine[2]), 64)
	if err != nil {
		return protocol.NewErrReply("args of request Err")
	}
	minBorder := &sortedset.ScoreBorder{
		Value: minValue,
	}

	maxBorder := &sortedset.ScoreBorder{
		Value: maxValue,
	}

	ss := db.getOrInitSortedSet(key)
	result := ss.CountInRange(minBorder, maxBorder)
	return protocol.NewIntReply(result)
}

// ZINCRBY key increment member
func execZIncrBy(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	increment, err := strconv.ParseFloat(string(cmdLine[1]), 64)
	if err != nil {
		return protocol.NewErrReply("invalid increment args")
	}
	member := string(cmdLine[2])
	ss := db.getOrInitSortedSet(key)
	element := ss.Get(member)
	if element == nil {
		return protocol.NewErrReply("no such a member")
	}
	element.Score += increment
	result := ss.Put(element.Member, element.Score)
	return protocol.NewIntReply(int64(result))
}

// ZRank key member
func execZRank(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	member := string(cmdLine[1])
	ss := db.getOrInitSortedSet(key)
	rank, err := ss.GetRank(member, true)
	if err != nil {
		return protocol.NewErrReply("no such a member")
	}
	return protocol.NewIntReply(rank)
}

// ZSCORE key member
func execZScore(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	member := string(cmdLine[1])
	ss := db.getOrInitSortedSet(key)
	element := ss.Get(member)
	if element == nil {
		return protocol.NewErrReply("no such a member")
	}
	result := strconv.FormatFloat(element.Score, 'f', 2, 64)
	return protocol.NewBulkReply([]byte(result))
}

// ZRange key start stop
func execZRange(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	start, err := strconv.ParseInt(string(cmdLine[1]), 10, 64)
	stop, err := strconv.ParseInt(string(cmdLine[2]), 10, 64)
	if err != nil {
		return protocol.NewErrReply("no such a member")
	}

	ss := db.getOrInitSortedSet(key)
	elements := ss.GetByRankRange(start, stop)
	if elements == nil {
		return protocol.NewErrReply("invalid rank range")
	}
	args := make([][]byte, len(elements))
	for i, element := range elements {
		member := element.Member
		score := strconv.FormatFloat(element.Score, 'f', 2, 64)
		args[i] = []byte(member + ":" + score)
	}
	return protocol.NewMultiReply(args)
}

// execZRem: ZREM key member [member ...]
func execZRem(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	members := make([]string, len(cmdLine)-1)
	for i := 1; i < len(cmdLine); i++ {
		members[i-1] = string(cmdLine[i])
	}
	ss := db.getOrInitSortedSet(key)
	var result int64 = 0
	for _, member := range members {
		result += ss.Remove(member)
	}
	return protocol.NewIntReply(result)
}

// ZRangeByScore key min max
func execZRangeByScore(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	minValue, err := strconv.ParseFloat(string(cmdLine[1]), 64)
	maxValue, err := strconv.ParseFloat(string(cmdLine[2]), 64)
	if err != nil {
		return protocol.NewErrReply("in valid args")
	}
	ss := db.getOrInitSortedSet(key)
	min := &sortedset.ScoreBorder{
		Value: minValue,
	}

	max := &sortedset.ScoreBorder{
		Value: maxValue,
	}

	elements := ss.GetByRange(min, max, true)
	if elements == nil {
		return protocol.NewErrReply("invalid rank range")
	}
	args := make([][]byte, len(elements))
	for i, element := range elements {
		member := element.Member
		score := strconv.FormatFloat(element.Score, 'f', 2, 64)
		args[i] = []byte(member + ":" + score)
	}
	return protocol.NewMultiReply(args)
}

func execZRemRangeByRank(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	start, err := strconv.ParseInt(string(cmdLine[1]), 10, 64)
	stop, err := strconv.ParseInt(string(cmdLine[2]), 10, 64)
	if err != nil {
		return protocol.NewErrReply("in valid args")
	}
	ss := db.getOrInitSortedSet(key)
	result := ss.RemByRankRange(start, stop)
	return protocol.NewIntReply(result)
}
