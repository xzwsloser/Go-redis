package sortedset

import (
	"fmt"
	"math/bits"
	"math/rand"
	"strconv"
)

/**
@Author: loser
@Description: 实现跳表
*/

const (
	maxLevel = 16
)

// Element 表示节点成员
type Element struct {
	Member string
	Score  float64
}

type Node struct {
	Element
	prev  *Node
	level []*Level
}

type Level struct {
	span int64
	next *Node
}

type skipList struct {
	header *Node // 头节点
	tail   *Node // 尾节点
	length int64 // 跳表长度
	level  int16 // 最大层数
}

func newNode(level int, member string, score float64) *Node {
	node := &Node{
		Element: Element{
			Member: member,
			Score:  score,
		},
		level: make([]*Level, int16(level)),
	}

	for i := 0; i < level; i++ {
		node.level[i] = new(Level)
	}
	return node
}

func newSkipList() *skipList {
	return &skipList{
		header: newNode(maxLevel, "", 0),
		tail:   nil,
		length: 0,
		level:  1,
	}
}

// randomLevel 产生随机层数
func randomLevel() int16 {
	total := uint64(1)<<uint64(maxLevel) - 1 // 1 << 16 - 1
	k := rand.Uint64() % total
	return maxLevel - int16(bits.Len64(k+1)) + 1
}

// insertNode 向跳跃表中插入数据
func (skiplist *skipList) insertNode(member string, score float64) *Node {
	// update 记录每一层下降位置的节点
	update := make([]*Node, maxLevel)
	// rank 表示每一层下降节点的累加值
	rank := make([]int64, maxLevel)

	node := skiplist.header
	// 开始每一层每一层遍历
	for i := skiplist.level - 1; i >= 0; i-- {
		// 最高的一个层次,累积排名为 0
		if i == skiplist.level-1 {
			rank[i] = 0
		} else {
			// 节点的值相等,只是层数而已
			rank[i] = rank[i+1]
		}

		// 寻找插入节点的位置,之后一路下降即可
		if node.level[i] != nil {
			for node.level[i].next != nil &&
				(node.level[i].next.Score < score ||
					(node.level[i].next.Score == score &&
						node.level[i].next.Member < member)) {
				rank[i] += node.level[i].span
				node = node.level[i].next
			}
		}
		update[i] = node
	}

	// 找到了目标节点的位置
	level := randomLevel()
	// 如果这一个节点的层数量大于最大层数量,那么就需要增加第一个节点的参数
	if level > skiplist.level {
		for i := skiplist.level; i < level; i++ {
			rank[i] = 0
			update[i] = skiplist.header
			update[i].level[i].span = skiplist.length
		}
		skiplist.level = level
	}

	// 开始插入节点
	pnew := newNode(int(level), member, score)
	for i := 0; i < int(level); i++ {
		// 注意层次和节点之间的关系! 层次 next 指针指向节点
		pnew.level[i].next = update[i].level[i].next
		update[i].level[i].next = pnew
		pnew.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = rank[0] - rank[i] + 1
	}

	// 上层的节点 span 增加
	for i := level; i < skiplist.level; i++ {
		update[i].level[i].span++
	}

	// 最后处理当前节点和头节点之间的关系
	if update[0] == skiplist.header {
		pnew.prev = nil
	} else {
		pnew.prev = update[0]
	}

	// 处理当前节点和后面节点的关系
	if pnew.level[0].next == nil {
		skiplist.tail = pnew
	} else {
		pnew.level[0].next.prev = pnew
	}
	skiplist.length++
	return pnew
}

func (skiplist *skipList) remove(member string, score float64) bool {
	update := make([]*Node, maxLevel)
	node := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		if node.level[i] != nil {
			for node.level[i].next != nil &&
				(node.level[i].next.Score < score ||
					(node.level[i].next.Score == score &&
						node.level[i].next.Member < member)) {
				node = node.level[i].next
			}
		}
		update[i] = node
	}

	node = node.level[0].next
	if node != nil && node.Score == score && node.Member == member {
		skiplist.removeNode(node, update)
		return true
	}
	return false
}

func (skiplist *skipList) removeNode(node *Node, update []*Node) {
	for i := int16(0); i < skiplist.level; i++ {
		if update[i].level[i].next == node {
			update[i].level[i].span += node.level[i].span - 1
			update[i].level[i].next = node.level[i].next
		} else {
			update[i].level[i].span--
		}
	}

	// 处理删除节点和后面节点的关系
	if node.level[0].next != nil {
		node.level[0].next.prev = node.prev
	} else {
		skiplist.tail = node.prev
	}

	if skiplist.level > 1 && skiplist.header.level[skiplist.level-1].next == nil {
		skiplist.level--
	}
	skiplist.length--
}

func (skiplist *skipList) show() {
	for i := skiplist.level - 1; i >= 0; i-- {
		fmt.Print("h")
		tmp := skiplist.header
		for tmp.level[i].next != nil {
			for j := 0; j < int(tmp.level[i].span); j++ {
				fmt.Print(" ")
			}
			fmt.Print(strconv.FormatFloat(tmp.level[i].next.Score, 'f', 2, 64))
			tmp = tmp.level[i].next
		}
		fmt.Println("")
	}
}
