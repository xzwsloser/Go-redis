package list

type listNode struct {
	data any
	prev *listNode
	next *listNode
}

type LinkedList struct {
	head *listNode
	tail *listNode
	size int
}

func newListNode(data any) *listNode {
	return &listNode{
		data: data,
		prev: nil,
		next: nil,
	}
}

func NewLinkedList() *LinkedList {
	list := &LinkedList{
		head: nil,
		tail: nil,
		size: 0,
	}
	return list
}

func (list *LinkedList) Empty() bool {
	return list.size == 0
}

func (list *LinkedList) findNode(idx int) *listNode {
	if list.Empty() || idx < 0 || idx >= list.size {
		return nil
	}

	var res *listNode
	if idx < list.size/2 {
		ptr := list.head
		for i := 0; i < idx; i++ {
			ptr = ptr.next
		}
		res = ptr
	} else {
		ptr := list.tail
		for i := 0; i < (list.size - idx - 1); i++ {
			ptr = ptr.prev
		}
		res = ptr
	}
	return res
}

func (list *LinkedList) Get(idx int) (value any) {
	node := list.findNode(idx)
	if node == nil {
		return nil
	}
	return node.data
}

func (list *LinkedList) InsertHead(value any) (l int) {
	pnew := newListNode(value)
	if list.Empty() {
		list.head = pnew
		list.tail = pnew
		pnew.prev = pnew
		pnew.next = pnew
		list.size++
		return
	}
	pnew.next = list.head
	list.head.prev = pnew
	list.head = pnew
	pnew.prev = list.tail
	list.tail.next = pnew
	list.size++
	return list.size
}

func (list *LinkedList) InsertTail(value any) (l int) {
	pnew := newListNode(value)
	if list.Empty() {
		list.head = pnew
		list.tail = pnew
		pnew.prev = pnew
		pnew.next = pnew
		list.size++
		l = list.size
		return
	}
	list.tail.next = pnew
	pnew.prev = list.tail
	list.tail = pnew
	pnew.next = list.head
	list.head.prev = pnew
	list.size++
	l = list.size
	return
}

func (list *LinkedList) RemoveHead() (value any) {
	if list.Empty() {
		return nil
	}

	if list.size == 1 {
		value = list.head.data
		list.head = nil
		list.tail = nil
		list.size--
		return
	}

	value = list.head.data
	list.tail.next = list.head.next
	list.head.next.prev = list.tail
	list.head = list.head.next
	list.size--
	return
}

func (list *LinkedList) RemoveTail() (value any) {
	if list.Empty() {
		return nil
	}

	if list.size == 1 {
		value = list.tail.data
		list.head = nil
		list.tail = nil
		list.size--
		return
	}

	value = list.tail.data
	list.tail.prev.next = list.head
	list.head.prev = list.tail.prev
	list.tail = list.tail.prev
	list.size--
	return
}

func (list *LinkedList) RemoveByValue(value any, reversed bool) (result int) {
	if list.Empty() {
		return 0
	}
	var ptr *listNode
	if reversed {
		ptr = list.tail
	} else {
		ptr = list.head
	}

	for i := 0; i < list.size; i++ {
		if ptr.data == value {
			ptr.prev.next = ptr.next
			ptr.next.prev = ptr.prev
			ptr = nil
			list.size--
			return 1
		} else {
			if reversed {
				ptr = ptr.prev
			} else {
				ptr = ptr.next
			}
		}
	}
	return 0
}

func (list *LinkedList) RemoveByCond(condition func(int, any) bool) (result int) {
	if list.Empty() {
		return 0
	}
	ptr := list.head
	for i := 0; i < list.size; i++ {
		if condition(i, ptr.data) {
			ptr.prev.next = ptr.next
			ptr.next.prev = ptr.prev
			list.size--
			return 1
		}
		ptr = ptr.next
	}
	return 0
}

func (list *LinkedList) Len() int {
	return list.size
}

func (list *LinkedList) Set(idx int, value any) {
	node := list.findNode(idx)
	if node == nil {
		return
	}
	node.data = value
}

func (list *LinkedList) ForEach(consumer func(key any) bool) {
	if list.Empty() {
		return
	}
	ptr := list.head
	for i := 0; i < list.size; i++ {
		if consumer(ptr.data) {
			ptr = ptr.next
		} else {
			break
		}
	}
}

func (list *LinkedList) FindRangeValue(start int, stop int) (values []any) {
	if start < 0 {
		start = 0
	}

	if stop >= list.size {
		stop = list.size - 1
	}

	if start > stop || start >= list.size {
		return nil
	}

	ptr := list.head
	for i := 0; i < start; i++ {
		ptr = ptr.next
	}

	values = make([]any, stop-start+1)
	for j := start; j <= stop; j++ {
		values[j-start] = ptr.data
		ptr = ptr.next
	}
	return values
}
