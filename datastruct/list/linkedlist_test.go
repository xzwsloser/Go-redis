package list

import (
	"log"
	"testing"
)

func TestLinkedListInsertHead(t *testing.T) {
	list := NewLinkedList()
	list.InsertHead(1)
	list.InsertHead(2)
	list.InsertHead(3)
	list.InsertHead(4)
	list.InsertHead(5)
	list.InsertHead(6)
	list.ForEach(func(key any) bool {
		log.Print(key, " ")
		return true
	})
}

func TestLinkedListInsertTail(t *testing.T) {
	list := NewLinkedList()
	list.InsertTail(1)
	list.InsertTail(2)
	list.InsertTail(3)
	list.InsertTail(4)
	list.InsertTail(5)
	list.InsertTail(6)
	list.ForEach(func(key any) bool {
		log.Print(key, " ")
		return true
	})
	log.Println("==")
	log.Print(list.Len())
}

func TestLinkedListRemoveHead(t *testing.T) {
	list := NewLinkedList()
	list.InsertHead(1)
	list.InsertHead(2)
	list.InsertHead(3)
	list.InsertHead(4)
	list.InsertHead(5)
	value := list.Get(4)
	list.Set(4, 1000)
	value = list.Get(4)
	log.Print("====", value, "====")
}

func TestLinkedListRemove(t *testing.T) {
	list := NewLinkedList()
	list.InsertHead(1)
	list.InsertHead(2)
	list.InsertHead(3)
	list.InsertHead(4)
	list.InsertHead(5)
	list.InsertHead(6)
	consumer := func(key any) bool {
		log.Print(key, " ")
		return true
	}
	log.Println("删除之前:")
	list.ForEach(consumer)
	log.Println("\n删除之后")
	list.RemoveHead()
	list.RemoveHead()
	list.RemoveTail()
	list.ForEach(consumer)
}

func TestRemoveByValue(t *testing.T) {
	list := NewLinkedList()
	list.InsertTail("hello1")
	list.InsertTail("hello2")
	list.InsertTail("hello3")
	list.InsertTail("hello4")
	list.InsertTail("hello4444")
	list.InsertTail("hello5555")
	list.InsertTail("hello4")
	list.InsertTail("hello5")
	list.InsertTail("hello6")
	list.InsertTail("hello7")
	consumer := func(key any) bool {
		log.Print(key, " ")
		return true
	}
	log.Println("删除之前:")
	list.ForEach(consumer)
	log.Println("\n删除之后")
	list.RemoveByValue("hello4", false)
	list.ForEach(consumer)
}

func TestRange(t *testing.T) {
	list := NewLinkedList()
	list.InsertTail("hello1")
	list.InsertTail("hello2")
	list.InsertTail("hello3")
	list.InsertTail("hello4")
	list.InsertTail("hello4444")
	list.InsertTail("hello5555")
	list.InsertTail("hello4")
	list.InsertTail("hello5")
	list.InsertTail("hello6")
	list.InsertTail("hello7")
	values := list.FindRangeValue(1, 8)
	for i, v := range values {
		log.Println(i, " == ", v)
	}
	list.RemoveByCond(func(i int, value any) bool {
		return i == 3
	})
	log.Println("=================")
	list.ForEach(func(key any) bool {
		log.Println(key)
		return true
	})
}
