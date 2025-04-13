package dict

type Consumer func(key string, value any) bool

// Dict is the data structure of dict
type Dict interface {
	Get(key string) (value any, exists bool)
	GetWithLock(key string) (value any, exists bool)
	Len() int
	Put(key string, value any) (result int)
	PutWithLock(key string, value any) (result int)
	PutIfAbsent(key string, value any) (result int)
	PutIfAbsentWithLock(key string, value any) (result int)
	PutIfExists(key string, value any) (result int)
	PutIfExistsWithLock(key string, value any) (result int)
	Remove(key string) (value any, result int)
	RemoveWithLock(key string) (value any, result int)
	ForEach(consumer Consumer)
	Keys() []string
	RandomKeys(limit int) []string
	RandomDistinctKeys(limit int) []string
	Clear()
	DictScan(cursor int, count int, pattern string) ([][]byte, int)
}
