package collection

import (
	"sync"
	"sync/atomic"
)

type Node struct {
	next       atomic.Value //存储*Node
	val        int32
	sync.Mutex              // 节点锁
	marked     atomic.Value // 为true表示已经被删除
}

type ConcurrentLinkedQueue struct {
	root       atomic.Value // *Node
	length     int32
	sync.Mutex // 仅修改root节点使用
}

// NewLinkedQueue 返回一个全新的有序链表
func NewLinkedQueue() *ConcurrentLinkedQueue {
	res := &ConcurrentLinkedQueue{}
	var t *Node
	res.root.Store(t)
	return res
}

func (q *ConcurrentLinkedQueue) Contains(value int32) bool {
	for cur := q.root.Load().(*Node); cur != nil; cur = cur.next.Load().(*Node) {
		if v := atomic.LoadInt32(&cur.val); v == value {
			return true
		} else if v > value {
			return false
		}
	}
	return false
}

func (q *ConcurrentLinkedQueue) Insert(value int32) bool {
	b, a := q.findBAndA(value)
	if b == nil {
		q.Lock()
		defer q.Unlock()
	} else {
		b.Lock()
		defer b.Unlock()
		if b.next.Load().(*Node) != a {
			return false
		}
	}
	curNode := Node{
		val: value,
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

func (q *ConcurrentLinkedQueue) Delete(value int32) bool {
	b, a := q.findBAndA(value)
	if a == nil {
		return true // 无需删除
	}
	a.Lock()
	defer a.Unlock()
	if b == nil {
		q.Lock()
		defer q.Unlock()
	} else {
		b.Lock()
		defer b.Unlock()
	}
	if (b != nil && (b.next.Load().(*Node) != a || b.marked.Load().(bool))) || (atomic.LoadInt32(&a.val) != value) ||
		a.marked.Load().(bool) {
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

func (q *ConcurrentLinkedQueue) Range(f func(value int32) bool) {
	for cur := q.root.Load().(*Node); cur != nil; cur = cur.next.Load().(*Node) {
		f(atomic.LoadInt32(&cur.val))
	}
}

func (q *ConcurrentLinkedQueue) Len() int32 {
	return atomic.LoadInt32(&q.length)
}

func (q *ConcurrentLinkedQueue) findBAndA(val int32) (b, a *Node) {
	cur := q.root.Load().(*Node)
	a = cur
	for cur != nil && atomic.LoadInt32(&cur.val) < val {
		b = cur
		cur = cur.next.Load().(*Node)
		a = cur
	}
	return
}
