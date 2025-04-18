package timeheap

import (
	"testing"
	"time"
)

func TestTimeHeap(t *testing.T) {
	th := NewTimeHeap(time.Millisecond * 500)
	th.Start()
	//th.AddTask(time.Second, "k1", func() {
	//	log.Print("第一次执行 1s")
	//})
	//
	//th.AddTask(time.Second*2, "k2", func() {
	//	log.Print("第二次执行 2s")
	//})
	//
	//th.AddTask(time.Second*5, "k3", func() {
	//	log.Print("第三次执行 5s")
	//})
	//
	//th.AddTask(time.Second*6, "k4", func() {
	//	log.Print("第四次执行 6s")
	//})
	//
	//
	//th.AddTask(time.Second*7, "k5", func() {
	//	log.Print("第四次执行 6s")
	//})
	time.Sleep(10 * time.Second)
}
