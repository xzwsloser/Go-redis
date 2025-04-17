package sortedset

import (
	"log"
	"testing"
)

func TestCountInRange(t *testing.T) {
	ss := NewSortedSet()
	ss.Put("a", 1.1)
	ss.Put("b", 2.2)
	ss.Put("c", 3.3)
	ss.Put("d", 4.4)
	ss.Put("e", 5.5)
	ss.Put("f", 6.6)
	min := &ScoreBorder{
		Value:   1.1,
		Exclude: true,
	}

	max := &ScoreBorder{
		Value: 4.0,
	}
	inRange := ss.CountInRange(min, max)
	log.Println("数量: ", inRange)
}
