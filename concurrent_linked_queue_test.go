package collection

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConcurrentLinkedQueue_Contains(t *testing.T) {

}

func TestConcurrentLinkedQueue_Delete(t *testing.T) {

}

func TestConcurrentLinkedQueue_Insert(t *testing.T) {
	q := NewLinkedQueue()
	Convey("concurrent insert", t, func() {
		count := 10000
		var wg sync.WaitGroup
		var succ int32
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func() {
				if q.Insert(rand.Int31n(1000)) {
					atomic.AddInt32(&succ, 1)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		So(q.Len(), ShouldEqual, succ)
	})
}

func BenchmarkConcurrentLinkedQueue_Insert(b *testing.B) {
	q := NewLinkedQueue()
	var count int32 = 1000000
	// first half to delete
	eles := make([]int32, 0, count)
	for i := 0; i < int(count); i++ {
		eles = append(eles, rand.Int31n(10000))
	}
	b.ResetTimer()
	var index int32 = -1
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			q.Insert(eles[atomic.AddInt32(&index, 1)])
		}
	})
}

func TestConcurrentLinkedQueue_Len(t *testing.T) {
}

func TestConcurrentLinkedQueue_Range(t *testing.T) {
}

func TestConcurrentLinkedQueue_findBAndA(t *testing.T) {

}

func TestNewLinkedQueue(t *testing.T) {
	Convey("NewLinkedQueue", t, func() {
		Convey("all success", func() {
			q := NewLinkedQueue()
			So(q, ShouldNotBeNil)
			So(q.Len(), ShouldEqual, 0)
		})
	})
}

func TestLinkedQueue(t *testing.T) {
	q := NewLinkedQueue()
	var count int32 = 100000
	// first half to delete
	eles := make([]int32, 0, count)
	for i := 0; i < int(count); i++ {
		eles = append(eles, rand.Int31n(10000))
	}

	Convey("concurrent test", t, func() {
		var wg sync.WaitGroup
		Convey("concurrent insert", func() {
			var succ int32
			wg.Add(int(count))
			for i := 0; i < int(count); i++ {
				index := i
				go func() {
					if q.Insert(eles[index]) {
						atomic.AddInt32(&succ, 1)
					}
					wg.Done()
				}()
			}
			wg.Wait()
			So(q.Len(), ShouldEqual, succ)
		})
		Convey("concurrent delete", func() {
			prevLen := q.Len()
			var fail int32
			wg.Add(int(count / 2))
			for i := 0; i < int(count)/2; i++ {
				index := i
				go func() {
					if q.Delete(eles[index]) {
						atomic.AddInt32(&fail, 1)
					}
					wg.Done()
				}()
			}
			wg.Wait()
			So(q.Len(), ShouldEqual, prevLen-fail)
		})
	})
}
