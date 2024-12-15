package stack

import (
	"sync/atomic"
	"unsafe"
)

/*
	Lock-Free Stack
*/

type Node struct {
	val  any
	next *Node
}

type BLFStack struct {
	top *Node
}

func NewBLFStack() *BLFStack {
	return &BLFStack{}
}

func (s *BLFStack) Push(val any) {
	for {
		// old := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&s.top)))
		p := unsafe.Pointer(s.top)
		oldTop := atomic.LoadPointer(&p)
		newTop := &Node{
			val:  val,
			next: (*Node)(oldTop),
		}

		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&s.top)), oldTop, unsafe.Pointer(newTop)) {
			// 添加成功
			return
		}
	}
}

func (s *BLFStack) Pop() (any, bool) {
	for {
		p := unsafe.Pointer(s.top)
		oldTop := atomic.LoadPointer(&p)
		if oldTop == nil {
			return nil, false
		}
		nextNode := (*Node)(oldTop).next
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&s.top)), oldTop, unsafe.Pointer(nextNode)) {
			return (*Node)(oldTop).val, true
		}
	}
}
