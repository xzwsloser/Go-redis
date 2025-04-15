package sortedset

import (
	"log"
	"testing"
)

func TestSkipListInsert(t *testing.T) {
	skiplist := newSkipList()
	skiplist.insertNode("a", 1.1)
	skiplist.insertNode("b", 2.3)
	skiplist.insertNode("c", 3.2)
	skiplist.insertNode("d", 4.4)
	skiplist.insertNode("e", 5.5)
	skiplist.insertNode("h", 5.7)
	skiplist.insertNode("k", 3.5)
	skiplist.insertNode("m", 1.5)
	skiplist.insertNode("n", 8.7)
	skiplist.insertNode("o", 9.8)
	skiplist.insertNode("p", 8.6)
	skiplist.insertNode("q", 10)
	skiplist.show()
}

func TestRemoveNode(t *testing.T) {
	skiplist := newSkipList()
	skiplist.insertNode("a", 1.1)
	skiplist.insertNode("b", 2.3)
	skiplist.insertNode("c", 3.2)

	skiplist.show()

	remove := skiplist.remove("a", 1.1)
	remove = skiplist.remove("b", 2.3)
	remove = skiplist.remove("c", 3.2)

	if remove {
		log.Println("删除成功...")
	} else {
		t.Error("测试失败...")
	}
}
