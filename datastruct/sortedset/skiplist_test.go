package sortedset

import (
	"fmt"
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
}

func TestRemoveNode(t *testing.T) {
	skiplist := newSkipList()
	skiplist.insertNode("a", 1.1)
	skiplist.insertNode("b", 2.3)
	skiplist.insertNode("c", 3.2)

	remove := skiplist.remove("a", 1.1)
	remove = skiplist.remove("b", 2.3)
	remove = skiplist.remove("c", 3.2)

	if remove {
		log.Println("删除成功...")
	} else {
		t.Error("测试失败...")
	}
}

func TestHasInRange(t *testing.T) {
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

	// 注意排除区间值
	min := &ScoreBorder{
		Value:   10,
		Exclude: true, // 表示 (
	}

	max := &ScoreBorder{
		Value: 111.0,
	}
	inRange := skiplist.hasInRange(min, max)
	if inRange {
		log.Print("测试成功...")
	} else {
		log.Print("测试失败")
	}
}

func TestFirstInRange(t *testing.T) {
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

	// 注意排除区间值
	min := &ScoreBorder{
		Value:   0.2,
		Exclude: false,
	}

	max := &ScoreBorder{
		Value: 12,
	}

	first := skiplist.getFirstNodeInRange(min, max)
	if first != nil {
		log.Println(first.Element)
	}
}

func TestFindLastInRange(t *testing.T) {
	skiplist := newSkipList()
	skiplist.insertNode("a", 1)
	skiplist.insertNode("b", 2)
	skiplist.insertNode("c", 3)
	skiplist.insertNode("d", 4)
	skiplist.insertNode("e", 5)
	skiplist.insertNode("f", 6)
	skiplist.insertNode("g", 7)
	skiplist.insertNode("h", 8)
	skiplist.insertNode("i", 9)
	skiplist.insertNode("j", 10)

	min := &ScoreBorder{
		Value: 1.1,
	}

	max := &ScoreBorder{
		Value:   10.1,
		Exclude: false,
	}

	inRange := skiplist.getLastNodeInRange(min, max)
	if inRange == nil {
		log.Print("测试失败...")
	} else {
		log.Print(inRange.Element)
	}
}

func TestRemoveByRange(t *testing.T) {
	skiplist := newSkipList()
	skiplist.insertNode("a", 1)
	skiplist.insertNode("b", 2)
	skiplist.insertNode("c", 3)
	skiplist.insertNode("d", 4)
	skiplist.insertNode("e", 5)
	skiplist.insertNode("f", 6)
	skiplist.insertNode("g", 7)
	skiplist.insertNode("h", 8)
	skiplist.insertNode("i", 9)
	skiplist.insertNode("j", 10)

	min := &ScoreBorder{
		Value:   1,
		Exclude: true,
	}

	max := &ScoreBorder{
		Value:   9,
		Exclude: true,
	}

	removed := skiplist.removeByRange(min, max, 100)
	for _, v := range removed {
		fmt.Println(v.Member, v.Score)
	}
}

func TestRemoveByRank(t *testing.T) {
	skiplist := newSkipList()
	skiplist.insertNode("a", 1)
	skiplist.insertNode("b", 2)
	skiplist.insertNode("c", 3)
	skiplist.insertNode("d", 4)
	skiplist.insertNode("e", 5)
	skiplist.insertNode("f", 6)
	skiplist.insertNode("g", 7)
	skiplist.insertNode("h", 8)
	skiplist.insertNode("i", 9)
	skiplist.insertNode("j", 10)

	//rank := skiplist.getByRank(1)
	//fmt.Println(rank.Element, rank.Score)

	removed := skiplist.removeByRank(4, 8)
	for _, v := range removed {
		fmt.Println(v.Member, v.Score)
	}
}
