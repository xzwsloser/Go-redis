package atomic

import "sync/atomic"

type Boolean uint32

func (b *Boolean) Get() bool {
	if atomic.LoadUint32((*uint32)(b)) == 1 {
		return true
	}

	return false
}

func (b *Boolean) Set(v bool) {
	var val uint32 = 0
	if v {
		val = 1
	}

	atomic.StoreUint32((*uint32)(b), val)
}
