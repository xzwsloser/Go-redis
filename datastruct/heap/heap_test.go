package heap

import (
	"log"
	"testing"
)

func TestHeapSort(t *testing.T) {
	arr := []any{5, 3, 8, 1, 2, 11, 9, 7, 10, 22}
	heap := NewHeap(arr, func(a any, b any) bool {
		at := a.(int)
		bt := b.(int)
		return at > bt
	})

	heap.Push(31)
	heap.Push(30)
	heap.Push(12)
	heap.Push(13)

	for heap.Len() > 0 {
		log.Print(heap.Top().(int), " ")
		heap.Pop()
	}
	log.Println()
}
