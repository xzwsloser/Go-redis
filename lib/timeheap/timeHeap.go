package timeheap

import (
	"github.com/xzwsloser/Go-redis/datastruct/heap"
	"sync"
	"time"
)

// timeNode 时间堆上的节点
type timeNode struct {
	job        func()    // 回调函数
	key        string    // 对应 Redis 中的 key
	expireTime time.Time // 过期的时间
	valid      bool      // 是否有效,也就是是否被删除
}

type TimeHeap struct {
	mu          *sync.Mutex
	interval    time.Duration        // 心搏函数的执行时间间隔
	ticker      *time.Ticker         // 提供周期的信号
	heap        *heap.Heap           // 底层的时间堆对象
	keyToNode   map[string]*timeNode // Redis 中的 key 到 timeNode 的映射关系
	addTaskChan chan *timeNode       // 添加任务的管道
	removeChan  chan string          // 删除任务管道
	stopChan    chan struct{}        // 停止管道
}

func NewTimeHeap(duration time.Duration) *TimeHeap {
	compare := func(a any, b any) bool {
		at := a.(*timeNode)
		bt := b.(*timeNode)
		return at.expireTime.After(bt.expireTime)
	}

	return &TimeHeap{
		mu:          &sync.Mutex{},
		interval:    duration,
		heap:        heap.NewHeap([]any{}, compare),
		keyToNode:   make(map[string]*timeNode),
		addTaskChan: make(chan *timeNode),
		removeChan:  make(chan string),
		stopChan:    make(chan struct{}),
	}
}

// @brief: AddTask 提供给外界的函数,用于添加任务
func (th *TimeHeap) AddTask(expireTime time.Time, key string, job func()) {
	tn := &timeNode{
		expireTime: expireTime,
		key:        key,
		job:        job,
		valid:      true,
	}
	th.addTaskChan <- tn
}

// @brief: RemoveTask 提供给外界的函数,用于删除任务
func (th *TimeHeap) RemoveTask(key string) {
	th.removeChan <- key
}

func (th *TimeHeap) Start() {
	th.ticker = time.NewTicker(th.interval)
	go func() {
		for {
			select {
			case task := <-th.addTaskChan:
				th.add(task)
			case key := <-th.removeChan:
				th.remove(key)
			case <-th.ticker.C:
				th.tick()
			}
		}
	}()
}

func (th *TimeHeap) Stop() {
	th.stopChan <- struct{}{}
}

func (th *TimeHeap) add(tn *timeNode) {
	th.mu.Lock()
	defer th.mu.Unlock()
	if node, ok := th.keyToNode[tn.key]; ok {
		th.keyToNode[tn.key] = tn
		node.valid = false
		th.heap.Push(tn)
		return
	}
	th.keyToNode[tn.key] = tn
	th.heap.Push(tn)
	return
}

func (th *TimeHeap) remove(key string) {
	th.mu.Lock()
	defer th.mu.Unlock()
	if node, ok := th.keyToNode[key]; ok {
		node.valid = false
		delete(th.keyToNode, key)
		return
	}
}

func (th *TimeHeap) tick() {
	th.mu.Lock()
	defer th.mu.Unlock()
	if th.heap.Len() == 0 {
		return
	}
	curNode := th.heap.Top().(*timeNode)
	keys := make([]string, 0)
	for th.heap.Len() > 0 && time.Now().After(curNode.expireTime) {
		th.heap.Pop()
		if curNode.valid {
			keys = append(keys, curNode.key)
			task := curNode.job
			go task()
		}
		if th.heap.Len() > 0 {
			curNode = th.heap.Top().(*timeNode)
		}
	}
	go func() {
		for _, key := range keys {
			th.removeChan <- key
		}
	}()
}
