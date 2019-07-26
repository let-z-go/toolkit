package condition

import (
	"context"
	"errors"
	"sync"
	"unsafe"

	"github.com/let-z-go/intrusive_containers/list"
)

type Condition struct {
	lock          sync.Locker
	listOfWaiters list.List
}

func (self *Condition) Initialize(lock sync.Locker) *Condition {
	if self.lock != nil {
		panic(errors.New("toolkit: condition already initialized"))
	}

	self.lock = lock
	self.listOfWaiters.Initialize()
	return self
}

func (self *Condition) WaitFor(context_ context.Context) (bool, error) {
	self.checkUninitialized()
	var waiter conditionWaiter
	waiter.Event = make(chan struct{}, 1)
	self.listOfWaiters.AppendNode(&waiter.ListNode)
	self.lock.Unlock()
	var e error

	select {
	case <-waiter.Event:
		e = nil
	case <-context_.Done():
		e = context_.Err()
	}

	self.lock.Lock()
	ok := e == nil || waiter.ListNode.IsReset()

	if !ok {
		waiter.ListNode.Remove()
	}

	return ok, e
}

func (self *Condition) Signal() {
	self.checkUninitialized()

	if self.listOfWaiters.IsEmpty() {
		return
	}

	waiter := (*conditionWaiter)(self.listOfWaiters.GetHead().GetContainer(unsafe.Offsetof(conditionWaiter{}.ListNode)))
	waiter.ListNode.Remove()
	waiter.ListNode.Reset()
	waiter.Event <- struct{}{}
}

func (self *Condition) Broadcast() {
	self.checkUninitialized()
	getNode := self.listOfWaiters.GetNodesSafely()

	for listNode := getNode(); listNode != nil; listNode = getNode() {
		listNode.Reset()
		waiter := (*conditionWaiter)(listNode.GetContainer(unsafe.Offsetof(conditionWaiter{}.ListNode)))
		waiter.Event <- struct{}{}
	}

	self.listOfWaiters.Initialize()
}

func (self *Condition) checkUninitialized() {
	if self.lock == nil {
		panic(errors.New("toolkit: condition uninitialized"))
	}
}

type conditionWaiter struct {
	ListNode list.ListNode
	Event    chan struct{}
}
