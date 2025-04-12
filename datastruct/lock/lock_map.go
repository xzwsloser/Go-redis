package lock

import (
	"sort"
	"sync"
)

const (
	prime32 = uint32(16777619)
)

// Locks the lock map when operate many different keys
type Locks struct {
	table []*sync.RWMutex
}

// the FNV hash algorithm
func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

func NewLocks(tableSize int) *Locks {
	table := make([]*sync.RWMutex, tableSize)
	for i := 0; i < tableSize; i++ {
		table[i] = &sync.RWMutex{}
	}

	return &Locks{
		table: table,
	}
}

func (ls *Locks) spread(key string) int {
	if ls == nil {
		panic("lock map is nil")
	}
	hashCode := fnv32(key)
	return int(hashCode) & (len(ls.table) - 1)
}

func (ls *Locks) toIndexSlices(keys []string, reverse bool) []int {
	if ls == nil {
		panic("lock map is nil")
	}

	indexMap := make(map[uint32]struct{})
	for _, key := range keys {
		index := ls.spread(key)
		indexMap[uint32(index)] = struct{}{}
	}

	result := make([]int, 0, len(indexMap))
	for key, _ := range indexMap {
		result = append(result, int(key))
	}

	sort.Slice(result, func(i, j int) bool {
		if reverse {
			return result[i] > result[j]
		}
		return result[i] < result[j]
	})
	return result
}

func (ls *Locks) Locks(keys []string) {
	if ls == nil {
		panic("locks is empty")
	}
	indexs := ls.toIndexSlices(keys, false)
	for _, index := range indexs {
		ls.table[index].Lock()
	}
}

func (ls *Locks) RLocks(keys []string) {
	if ls == nil {
		panic("locks is nil")
	}
	indexs := ls.toIndexSlices(keys, false)
	for _, index := range indexs {
		ls.table[index].RLock()
	}
}

func (ls *Locks) Unlocks(keys []string) {
	if ls == nil {
		panic("locks is nil")
	}
	indexs := ls.toIndexSlices(keys, true)
	for _, index := range indexs {
		ls.table[index].Unlock()
	}
}

func (ls *Locks) RUnlocks(keys []string) {
	if ls == nil {
		panic("locks is nil")
	}
	indexs := ls.toIndexSlices(keys, true)
	for _, index := range indexs {
		ls.table[index].RUnlock()
	}
}

func (ls *Locks) RWLocks(writeKeys []string, readKeys []string) {
	if ls == nil {
		panic("locks is nil")
	}
	keys := append(writeKeys, readKeys...)
	indexs := ls.toIndexSlices(keys, false)
	writeIndexs := make(map[int]struct{})
	for _, key := range writeKeys {
		index := ls.spread(key)
		writeIndexs[index] = struct{}{}
	}

	for _, index := range indexs {
		lock := ls.table[index]
		_, w := writeIndexs[index]
		if w {
			lock.Lock()
		} else {
			lock.Unlock()
		}
	}
}

func (ls *Locks) RWUnlocks(writeKeys []string, readKeys []string) {
	if ls == nil {
		panic("locks is nil")
	}
	keys := append(writeKeys, readKeys...)
	indexs := ls.toIndexSlices(keys, true)
	writeIndexs := make(map[int]struct{})
	for _, key := range writeKeys {
		index := ls.spread(key)
		writeIndexs[index] = struct{}{}
	}

	for _, index := range indexs {
		lock := ls.table[index]
		_, w := writeIndexs[index]
		if w {
			lock.Unlock()
		} else {
			lock.RUnlock()
		}
	}
}
