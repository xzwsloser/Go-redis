package sortedset

import (
	"errors"
	"slices"
)

const (
	DefaultDictSize = 16
)

type SortedSet struct {
	dict     map[string]*Element
	skiplist *skipList
}

func NewSortedSet() *SortedSet {
	return &SortedSet{
		dict:     make(map[string]*Element),
		skiplist: newSkipList(),
	}
}

// @brief: Put: 向 SortedSet 中加入元素
func (s *SortedSet) Put(member string, score float64) (result int) {
	element, ok := s.dict[member]
	s.dict[member] = &Element{
		Member: member,
		Score:  score,
	}
	if ok {
		if element.Score == score {
			return 0
		} else {
			s.skiplist.remove(member, element.Score)
			s.skiplist.insertNode(member, score)
		}
	} else {
		s.skiplist.insertNode(member, score)
	}

	return 1
}

func (s *SortedSet) Get(member string) *Element {
	element, ok := s.dict[member]
	if ok {
		return element
	}
	return nil
}

func (s *SortedSet) Len() int64 {
	return int64(len(s.dict))
}

func (s *SortedSet) Remove(member string) int64 {
	element, ok := s.dict[member]
	if !ok {
		return 0
	}
	result := s.skiplist.remove(element.Member, element.Score)
	if result {
		return 1
	}
	return 0
}

func (s *SortedSet) CountInRange(min Border, max Border) int64 {
	firstNode := s.skiplist.getFirstNodeInRange(min, max)
	if firstNode == nil {
		return -1
	}
	lastNode := s.skiplist.getLastNodeInRange(min, max)
	if lastNode == nil {
		return -1
	}
	rl := s.skiplist.getRank(firstNode.Member, firstNode.Score)
	rr := s.skiplist.getRank(lastNode.Member, lastNode.Score)
	return rr - rl + 1
}

func (s *SortedSet) GetRank(member string, desc bool) (rank int64, err error) {
	element := s.Get(member)
	if element == nil {
		return 0, errors.New("member not exists")
	}
	// 1 2 3 4 5
	// 0 1 2 3 4
	// 4 3 2 1 0
	rank = s.skiplist.getRank(element.Member, element.Score)
	if desc {
		rank = s.Len() - rank
	} else {
		rank--
	}
	return
}

func (s *SortedSet) GetByRange(min Border, max Border, desc bool) []*Element {
	firstNode := s.skiplist.getFirstNodeInRange(min, max)
	lastNode := s.skiplist.getLastNodeInRange(min, max)
	if firstNode == nil || lastNode == nil {
		return nil
	}

	rl := s.skiplist.getRank(firstNode.Member, firstNode.Score)
	rr := s.skiplist.getRank(lastNode.Member, lastNode.Score)
	elements := make([]*Element, rr-rl+1)
	for i := rl; i <= rr; i++ {
		elements[i-rl] = &s.skiplist.getByRank(i).Element
	}

	if !desc {
		slices.Reverse(elements)
	}
	return elements
}

func (s *SortedSet) GetByRankRange(start int64, stop int64) []*Element {
	elements := make([]*Element, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		if i >= 1 && i <= s.Len() {
			elements = append(elements, &s.skiplist.getByRank(i).Element)
		}
	}
	return elements
}

func (s *SortedSet) RemByRankRange(start int64, stop int64) (result int64) {
	elements := s.skiplist.removeByRank(start, stop)
	result = int64(len(elements))
	return result
}

func (s *SortedSet) ForEach(consumer func(score float64, member string) bool) {
	if s.Len() == 0 {
		return
	}
	ptr := s.skiplist.header
	ptr = ptr.level[0].next
	for ptr != nil {
		if !consumer(ptr.Score, ptr.Member) {
			break
		}
		ptr = ptr.level[0].next
	}
}
