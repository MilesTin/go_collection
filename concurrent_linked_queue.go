package collection

import (
	"sync"
	"sync/atomic"
)

type Node struct {
	next       atomic.Value //存储*Node
	val        int64
	sync.Mutex              // 节点锁
	marked     atomic.Value // 为true表示已经被删除
}

type ConcurrentLinkedQueue struct {
	root       atomic.Value // *Node
	length     int32
	sync.Mutex // 仅修改root节点使用
}

// NewLinkedQueue 返回一个全新的有序链表
func NewInt() *ConcurrentLinkedQueue {
	res := &ConcurrentLinkedQueue{}
	var t *Node
	res.root.Store(t)
	return res
}

func (q *ConcurrentLinkedQueue) Contains(value int) bool {
	for cur := q.root.Load().(*Node); cur != nil; cur = cur.next.Load().(*Node) {
		if v := atomic.LoadInt64(&cur.val); v == int64(value) {
			return true
		} else if v > int64(value) {
			return false
		}
	}
	return false
}

func (q *ConcurrentLinkedQueue) Insert(value int) bool {
l1:
	b, a := q.findBAndA(value)
	if b == nil {
		q.Lock()
		if q.root.Load() != a {
			q.Unlock()
			goto l1
		}
		defer q.Unlock()
	} else {
		b.Lock()
		if b.next.Load().(*Node) != a || b.marked.Load().(bool) {
			b.Unlock()
			goto l1
		}
		defer b.Unlock()
	}
	curNode := Node{
		val: int64(value),
	}
	curNode.next.Store(a)
	curNode.marked.Store(false)
	if b != nil {
		b.next.Store(&curNode)
	} else {
		q.root.Store(&curNode)
	}
	atomic.AddInt32(&q.length, 1)
	return true
}

func (q *ConcurrentLinkedQueue) Delete(value int) bool {
l1:
	b, a := q.findBAndA(value)
	if a != nil {
		a.Lock()
	}
	if b == nil {
		q.Lock()
	} else {
		b.Lock()
	}
	if (b != nil && ((b.next.Load().(*Node) != a) || b.marked.Load().(bool))) || (a != nil && a.marked.Load().(bool)) ||
		(b == nil && q.root.Load() != a) {
		if a != nil {
			a.Unlock()
		}
		if b != nil {
			b.Unlock()
		} else {
			q.Unlock()
		}
		goto l1
	}
	if a != nil {
		defer a.Unlock()
	}
	if b != nil {
		defer b.Unlock()
	} else {
		defer q.Unlock()
	}
	if a == nil || int(atomic.LoadInt64(&a.val)) != value {
		return false
	}
	next := a.next.Load().(*Node)
	if b == nil {
		q.root.Store(next)
	} else {
		b.next.Store(next)
	}

	atomic.AddInt32(&q.length, -1)
	return true
}

func (q *ConcurrentLinkedQueue) Range(f func(value int) bool) {
	for cur := q.root.Load().(*Node); cur != nil; cur = cur.next.Load().(*Node) {
		if f(int(atomic.LoadInt64(&cur.val))) {
			continue
		} else {
			break
		}
	}
}

func (q *ConcurrentLinkedQueue) Len() int {
	return int(atomic.LoadInt32(&q.length))
}

func (q *ConcurrentLinkedQueue) findBAndA(val int) (b, a *Node) {
	cur := q.root.Load().(*Node)
	a = cur
	for cur != nil && atomic.LoadInt64(&cur.val) < int64(val) {
		b = cur
		cur = cur.next.Load().(*Node)
		a = cur
	}
	return
}
