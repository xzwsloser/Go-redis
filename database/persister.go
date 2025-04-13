package database

import "github.com/xzwsloser/Go-redis/aof"

func (server *RedisServer) bindPersister(persister *aof.Persister) {
	server.persister = persister
	for i := 0; i < len(server.dbSet); i++ {
		db := server.dbSet[i].Load().(*Database)
		db.addAof = func(cmdLine [][]byte) {
			if persister.AppendOnly {
				persister.SaveCmdLine(db.index, cmdLine)
			}
		}
	}
}
