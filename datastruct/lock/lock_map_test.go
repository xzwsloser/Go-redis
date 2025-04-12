package lock

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestLocks(t *testing.T) {
	var a int = 1
	var b int = 1
	var c int = 1
	locks := NewLocks(6)

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer func() {
				wg.Done()
			}()
			keys := []string{"a", "b", "c"}
			for j := 0; j < 10; j++ {
				locks.Locks(keys)
				a++
				b++
				c++
				locks.Unlocks(keys)
				time.Sleep(time.Millisecond * 10)
			}

		}()
	}
	wg.Wait()
	log.Println("a = ", a)
	log.Println("b = ", b)
	log.Println("c = ", c)

}
