package utils

import (
	"strconv"
	"time"
)

var (
	PEXPIREAT = []byte("PEXPIREAT")
)

type CmdLine = [][]byte

func CmdLine1(cmdName string, args ...string) CmdLine {
	var result = make([][]byte, len(args)+1)
	result[0] = []byte(cmdName)
	for i, arg := range args {
		result[i+1] = []byte(arg)
	}
	return result
}

func CmdLine2(cmdName string, args [][]byte) CmdLine {
	var result = make([][]byte, len(args)+1)
	result[0] = []byte(cmdName)
	for i, arg := range args {
		result[i+1] = arg
	}
	return result
}

func ExpireCmd(key string, expireAt time.Time) CmdLine {
	args := make([][]byte, 3)
	args[0] = PEXPIREAT
	args[1] = []byte(key)
	timeStamp := expireAt.UnixNano() / 1000000
	timeStampStr := strconv.FormatInt(timeStamp, 10)
	args[2] = []byte(timeStampStr)
	return args
}

func Equals(v1 any, v2 any) bool {
	b1, ok1 := v1.([]byte)
	b2, ok2 := v2.([]byte)
	if ok1 && ok2 {
		return ByteEquals(b1, b2)
	}
	return v1 == v2
}

func ByteEquals(b1 []byte, b2 []byte) bool {
	if b1 == nil && b2 == nil {
		return true
	} else if b1 == nil || b2 == nil {
		return false
	}

	if len(b1) != len(b2) {
		return false
	}

	for i := 0; i < len(b1); i++ {
		if b1[i] != b2[i] {
			return false
		}
	}
	return true
}
