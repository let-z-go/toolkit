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
	waiter.event = make(chan struct{}, 1)
	self.listOfWaiters.AppendNode(&waiter.listNode)
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
	ok := e == nil || waiter.listNode.IsReset()

	if !ok {
		waiter.listNode.Remove()
	}

	return ok, e
}

func (self *Condition) Signal() {
	self.checkUninitialized()

	if self.listOfWaiters.IsEmpty() {
		return
	}

	waiter := (*conditionWaiter)(self.listOfWaiters.GetHead().GetContainer(unsafe.Offsetof(conditionWaiter{}.listNode)))
	waiter.listNode.Remove()
	waiter.listNode.Reset()
	waiter.event <- struct{}{}
}

func (self *Condition) Broadcast() {
	self.checkUninitialized()
	getNode := self.listOfWaiters.GetNodes()

	for listNode, nextListNode := getNode(), getNode(); listNode != nil; listNode, nextListNode = nextListNode, getNode() {
		listNode.Reset()
		waiter := (*conditionWaiter)(listNode.GetContainer(unsafe.Offsetof(conditionWaiter{}.listNode)))
		waiter.event <- struct{}{}
	}

	self.listOfWaiters.Initialize()
}

func (self *Condition) checkUninitialized() {
	if self.lock == nil {
		panic(errors.New("toolkit: condition uninitialized"))
	}
}

type conditionWaiter struct {
	listNode list.ListNode
	event    chan struct{}
}
