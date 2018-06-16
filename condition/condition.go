package condition

import (
	"context"
	"sync"
	"unsafe"

	"github.com/let-z-go/intrusive_containers/list"
)

type Condition struct {
	lock        sync.Locker
	waiterList1 list.List
	waiterList2 list.List
}

func (self *Condition) Initialize(lock sync.Locker) {
	self.lock = lock
	self.waiterList1.Initialize()
	self.waiterList2.Initialize()
}

func (self *Condition) WaitFor(context_ context.Context) error {
	waiter := conditionWaiter{}
	waiter.event = make(chan struct{}, 1)
	self.waiterList1.AppendNode(&waiter.listNode)
	self.lock.Unlock()
	var e error

	if context_ == nil {
		<-waiter.event
		e = nil
	} else {
		select {
		case <-waiter.event:
			e = nil
		case <-context_.Done():
			e = context_.Err()
		}
	}

	self.lock.Lock()
	waiter.listNode.Remove()
	return e
}

func (self *Condition) Signal() {
	if self.waiterList1.IsEmpty() {
		return
	}

	waiter := (*conditionWaiter)(self.waiterList1.GetHead().GetContainer(unsafe.Offsetof(conditionWaiter{}.listNode)))
	waiter.event <- struct{}{}
	waiter.listNode.Remove()
	self.waiterList2.AppendNode(&waiter.listNode)
}

func (self *Condition) Broadcast() {
	getNode := self.waiterList1.GetNodes()

	for listNode := getNode(); listNode != nil; listNode = getNode() {
		waiter := (*conditionWaiter)(listNode.GetContainer(unsafe.Offsetof(conditionWaiter{}.listNode)))
		waiter.event <- struct{}{}
	}

	self.waiterList1.Append(&self.waiterList2)
}

type conditionWaiter struct {
	listNode list.ListNode
	event    chan struct{}
}
