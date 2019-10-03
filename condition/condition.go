package condition

import (
	"context"
	"sync"
	"unsafe"

	"github.com/let-z-go/intrusives/list"
)

type Condition struct {
	lock          sync.Locker
	listOfWaiters list.List
}

func (self *Condition) Init(lock sync.Locker) *Condition {
	self.lock = lock
	self.listOfWaiters.Init()
	return self
}

func (self *Condition) WaitFor(ctx context.Context) (bool, error) {
	var waiter conditionWaiter
	waiter.Event = make(chan struct{})
	self.listOfWaiters.AppendNode(&waiter.ListNode)
	self.lock.Unlock()
	var err error

	select {
	case <-waiter.Event:
		err = nil
	case <-ctx.Done():
		err = ctx.Err()
	}

	self.lock.Lock()
	ok := err == nil || waiter.ListNode.IsReset()

	if !ok {
		waiter.ListNode.Remove()
	}

	return ok, err
}

func (self *Condition) Signal() {
	if self.listOfWaiters.IsEmpty() {
		return
	}

	waiter := (*conditionWaiter)(self.listOfWaiters.Head().GetContainer(unsafe.Offsetof(conditionWaiter{}.ListNode)))
	waiter.ListNode.Remove()
	waiter.ListNode.Reset()
	close(waiter.Event)
}

func (self *Condition) Broadcast() {
	getNode := self.listOfWaiters.GetNodesSafely()

	for listNode := getNode(); listNode != nil; listNode = getNode() {
		listNode.Reset()
		waiter := (*conditionWaiter)(listNode.GetContainer(unsafe.Offsetof(conditionWaiter{}.ListNode)))
		close(waiter.Event)
	}

	self.listOfWaiters.Init()
}

type conditionWaiter struct {
	ListNode list.ListNode
	Event    chan struct{}
}
