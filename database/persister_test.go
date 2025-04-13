package database

import (
	"github.com/xzwsloser/Go-redis/aof"
	"testing"
)

func TestPersister(t *testing.T) {
	persister := aof.NewPersister()
	db := NewDatabase()
	db.addAof = func(cmdLine [][]byte) {
		if persister.AppendOnly {
			persister.SaveCmdLine(db.index, cmdLine)
		}
	}

	var commandSet = [][]byte{
		[]byte("k1"),
		[]byte("v1"),
	}
	execSet(db, commandSet)

	db.index = 2
	var commandSet1 = [][]byte{
		[]byte("k2"),
		[]byte("0"),
	}
	execSet(db, commandSet1)

	var commandIncr = [][]byte{
		[]byte("k2"),
	}
	execIncr(db, commandIncr)
}
