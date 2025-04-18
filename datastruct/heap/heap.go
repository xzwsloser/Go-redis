package heap

// Heap is the implement of priority queue
type Heap struct {
	arr     []any
	compare func(a any, b any) bool
}

func NewHeap(heap []any, compare func(a any, b any) bool) *Heap {
	heapInit(heap, compare)
	return &Heap{
		arr:     heap,
		compare: compare,
	}
}

func heapInit(heap []any, compare func(a any, b any) bool) {
	n := len(heap)
	for i := (n - 1) / 2; i >= 0; i-- {
		adjustHeap(heap, i, n, compare)
	}
}

func adjustHeap(heap []any, cur int, length int, compare func(a any, b any) bool) {
	if cur >= length {
		return
	}
	temp := heap[cur]
	for i := 2*cur + 1; i < length; i = i*2 + 1 {
		if i+1 < length && compare(heap[i], heap[i+1]) {
			i++
		}

		if compare(temp, heap[i]) {
			heap[cur] = heap[i]
			cur = i
		} else {
			break
		}
	}
	heap[cur] = temp
}

func (heap *Heap) Push(value any) {
	heap.arr = append(heap.arr, value)
	child := len(heap.arr) - 1
	parent := 0
	for ; child > 0; child = parent {
		parent = (child - 1) / 2
		if heap.compare(heap.arr[parent], heap.arr[child]) {
			heap.arr[child], heap.arr[parent] = heap.arr[parent], heap.arr[child]
		}
	}
}

func (heap *Heap) Top() any {
	if heap.arr == nil || len(heap.arr) == 0 {
		return nil
	}
	return heap.arr[0]
}

func (heap *Heap) Pop() any {
	n := len(heap.arr)
	heap.arr[0], heap.arr[n-1] = heap.arr[n-1], heap.arr[0]
	temp := heap.arr[n-1]
	adjustHeap(heap.arr, 0, n-1, heap.compare)
	heap.arr = heap.arr[:n-1]
	return temp
}

func (heap *Heap) Len() int {
	return len(heap.arr)
}
