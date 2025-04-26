package dict

import (
	"github.com/xzwsloser/Go-redis/lib/logger"
	"github.com/xzwsloser/Go-redis/lib/wildcard"
	"math"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ConcurrentDict is the map to read by many goroutinues
type ConcurrentDict struct {
	table      []*shard
	count      int32 // atomic.Int
	shardCount int
}

// shard is one of the slot of ConcurrentDict
type shard struct {
	m    map[string]any
	lock sync.RWMutex
}

// computeCap compute the most near 2^n near to the param
func computeCap(param int) int {
	if param < 16 {
		return 16
	}

	n := param - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16

	if n < 0 {
		return math.MaxInt32
	}

	return n + 1
}

func NewConcurrentDict(param int) *ConcurrentDict {
	if param == 1 {
		table := []*shard{
			&shard{
				m: make(map[string]any),
			},
		}

		return &ConcurrentDict{
			table:      table,
			count:      0,
			shardCount: 1,
		}
	}

	size := computeCap(param)
	table := make([]*shard, size)
	for i := 0; i < size; i++ {
		table[i] = &shard{
			m: make(map[string]any),
		}
	}

	d := &ConcurrentDict{
		table:      table,
		count:      0,
		shardCount: size,
	}

	return d
}

// the FNV hash algorithm
const prime32 = uint32(16777619)

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

func (d *ConcurrentDict) spread(key string) uint32 {
	if d == nil {
		panic("dict is nil")
	}
	if d.shardCount == 1 {
		return 0
	}

	hashCode := fnv32(key)
	tableSize := uint32(len(d.table))
	return hashCode & (tableSize - 1)
}

func (d *ConcurrentDict) addCount() {
	if d == nil {
		panic("dict is nil")
	}
	atomic.AddInt32(&d.count, 1)
}

func (d *ConcurrentDict) decreaseCount() {
	if d == nil {
		panic("dict is nil")
	}
	atomic.AddInt32(&d.count, -1)
}

func (s *shard) RandomKey() string {
	if s == nil {
		logger.Error("shard is empty!")
		return ""
	}

	for key := range s.m {
		return key
	}

	return ""
}

func (d *ConcurrentDict) getShard(key string) *shard {
	if d == nil {
		panic("dict is nil")
	}
	index := d.spread(key)
	return d.table[index]
}

func (d *ConcurrentDict) Get(key string) (value any, exists bool) {
	if d == nil {
		panic("dict is nil")
	}
	s := d.getShard(key)
	if s == nil {
		return nil, false
	}

	s.lock.RLock()
	defer s.lock.RUnlock()
	if value, ok := s.m[key]; ok {
		return value, true
	}
	return nil, false
}

func (d *ConcurrentDict) GetWithLock(key string) (value any, exists bool) {
	if d == nil {
		panic("dict is nil")
	}

	s := d.getShard(key)
	if s == nil {
		return nil, false
	}

	if value, ok := s.m[key]; ok {
		return value, true
	}
	return nil, false
}

func (d *ConcurrentDict) Len() int {
	if d == nil {
		panic("dict is nil")
	}
	return (int)(atomic.LoadInt32(&d.count))
}

func (d *ConcurrentDict) Put(key string, value any) (result int) {
	if d == nil {
		panic("dict is nil")
	}
	s := d.getShard(key)
	if s != nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		s.m[key] = value
		d.addCount()
		return 1
	}
	return 0
}

func (d *ConcurrentDict) PutWithLock(key string, value any) (result int) {
	if d == nil {
		panic("dict is nil")
	}

	s := d.getShard(key)
	if s != nil {
		s.m[key] = value
		d.addCount()
		return 1
	}
	return 0
}

func (d *ConcurrentDict) PutIfAbsent(key string, value any) (result int) {
	if d == nil {
		panic("dict is nil")
	}
	s := d.getShard(key)
	if s != nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		if _, exists := s.m[key]; exists {
			return 0
		}
		s.m[key] = value
		d.addCount()
		return 1
	}
	return 0
}

func (d *ConcurrentDict) PutIfAbsentWithLock(key string, value any) (result int) {
	if d == nil {
		panic("dict is nil")
	}

	s := d.getShard(key)
	if s != nil {
		if _, exists := s.m[key]; exists {
			return 0
		}
		s.m[key] = value
		d.addCount()
		return 1
	}
	return 0
}

func (d *ConcurrentDict) PutIfExists(key string, value any) (result int) {
	if d == nil {
		panic("dict is nil")
	}
	s := d.getShard(key)
	if s != nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		if _, exists := s.m[key]; !exists {
			return 0
		}
		s.m[key] = value
		d.addCount()
		return 1
	}
	return 0
}

func (d *ConcurrentDict) PutIfExistsWithLock(key string, value any) (result int) {
	if d == nil {
		panic("dict is nil")
	}

	s := d.getShard(key)
	if s != nil {
		if _, exists := s.m[key]; !exists {
			return 0
		}
		s.m[key] = value
		d.addCount()
		return 1
	}
	return 0
}

func (d *ConcurrentDict) Remove(key string) (value any, result int) {
	if d == nil {
		panic("dict is nil")
	}
	s := d.getShard(key)
	if s != nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		if _, exists := s.m[key]; !exists {
			return nil, 0
		}
		val := s.m[key]
		delete(s.m, key)
		d.decreaseCount()
		return val, 1
	}
	return nil, 0
}

func (d *ConcurrentDict) RemoveWithLock(key string) (value any, result int) {
	if d == nil {
		panic("dict is nil")
	}

	s := d.getShard(key)
	if s != nil {
		if _, exists := s.m[key]; !exists {
			return nil, 0
		}
		val := s.m[key]
		delete(s.m, key)
		d.decreaseCount()
		return val, 1
	}
	return nil, 0

}

func (d *ConcurrentDict) ForEach(consumer Consumer) {
	if d == nil {
		panic("dict is nil")
	}

	for _, s := range d.table {
		s.lock.RLock()
		f := func() bool {
			defer s.lock.RUnlock()
			for key, value := range s.m {
				if !consumer(key, value) {
					return false
				}
			}
			return true
		}

		if !f() {
			break
		}
	}
}

func (d *ConcurrentDict) Keys() []string {
	if d == nil {
		panic("dict is nil")
	}
	keys := make([]string, d.Len())
	i := 0
	for _, s := range d.table {
		s.lock.RLock()
		for key, _ := range s.m {
			keys[i] = key
			i++
		}
		s.lock.Unlock()
	}
	return keys
}

func (d *ConcurrentDict) RandomKeys(limit int) []string {
	if d == nil {
		panic("dict is nil")
	}

	nR := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]string, limit)
	for i := 0; i < limit; {
		index := nR.Intn(d.shardCount)
		s := d.table[index]
		if s != nil {
			s.lock.RLock()
			key := s.RandomKey()
			if key != "" {
				result[i] = key
				i++
			}
			s.lock.RUnlock()
		}
	}
	return result
}

func (d *ConcurrentDict) RandomDistinctKeys(limit int) []string {
	if d == nil {
		panic("dict is nil")
	}

	if limit > d.Len() {
		return d.Keys()
	}
	memo := make(map[string]struct{})
	nR := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(memo) < limit {
		index := nR.Intn(d.shardCount)
		s := d.table[index]
		if s != nil {
			s.lock.RLock()
			key := s.RandomKey()
			if key != "" {
				if _, exists := memo[key]; !exists {
					memo[key] = struct{}{}
				}
			}
			s.lock.Unlock()
		}
	}

	arr := make([]string, limit)
	i := 0
	for key, _ := range memo {
		arr[i] = key
		i++
	}
	return arr
}

func (d *ConcurrentDict) Clear() {
	*d = *NewConcurrentDict(d.shardCount)
}

func stringsToBytes(keys []string) [][]byte {
	result := make([][]byte, len(keys))
	for i, v := range keys {
		result[i] = []byte(v)
	}
	return result
}

// DictScan is scan the k-v pairs from the position of cursor
func (d *ConcurrentDict) DictScan(cursor int, count int, pattern string) ([][]byte, int) {
	if d == nil {
		panic("dict is nil")
	}

	size := d.Len()
	result := make([][]byte, 0)

	if pattern == "*" && count >= size {
		return stringsToBytes(d.Keys()), 0
	}

	matchKey, err := wildcard.CompilePattern(pattern)
	if err != nil {
		return result, -1
	}

	shardCount := len(d.table)
	shardIndex := cursor

	for shardIndex < shardCount {
		shard := d.table[shardIndex]
		shard.lock.RLock()
		if len(result)+len(shard.m) > count && shardIndex > cursor {
			shard.lock.RUnlock()
			return result, shardIndex
		}

		for key := range shard.m {
			if pattern == "*" || matchKey.IsMatch(key) {
				result = append(result, []byte(key))
			}
		}
		shard.lock.RUnlock()
		shardIndex++
	}

	return result, 0
}

func (dict *ConcurrentDict) toLockIndices(keys []string, reverse bool) []uint32 {
	indexMap := make(map[uint32]struct{})
	for _, key := range keys {
		index := dict.spread(key)
		indexMap[index] = struct{}{}
	}
	indices := make([]uint32, 0, len(indexMap))
	for index := range indexMap {
		indices = append(indices, index)
	}
	sort.Slice(indices, func(i, j int) bool {
		if !reverse {
			return indices[i] < indices[j]
		}
		return indices[i] > indices[j]
	})
	return indices
}

// RWLocks locks write keys and read keys together. allow duplicate keys
func (dict *ConcurrentDict) RWLocks(writeKeys []string, readKeys []string) {
	keys := append(writeKeys, readKeys...)
	indices := dict.toLockIndices(keys, false)
	writeIndexSet := make(map[uint32]struct{})
	for _, wKey := range writeKeys {
		idx := dict.spread(wKey)
		writeIndexSet[idx] = struct{}{}
	}
	for _, index := range indices {
		_, w := writeIndexSet[index]
		mu := &dict.table[index].lock
		if w {
			mu.Lock()
		} else {
			mu.RLock()
		}
	}
}

// RWUnLocks unlocks write keys and read keys together. allow duplicate keys
func (dict *ConcurrentDict) RWUnLocks(writeKeys []string, readKeys []string) {
	keys := append(writeKeys, readKeys...)
	indices := dict.toLockIndices(keys, true)
	writeIndexSet := make(map[uint32]struct{})
	for _, wKey := range writeKeys {
		idx := dict.spread(wKey)
		writeIndexSet[idx] = struct{}{}
	}
	for _, index := range indices {
		_, w := writeIndexSet[index]
		mu := &dict.table[index].lock
		if w {
			mu.Unlock()
		} else {
			mu.RUnlock()
		}
	}
}
