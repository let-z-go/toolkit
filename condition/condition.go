package condition

import (
	"context"
	"errors"
	"sync"
	"unsafe"

	"github.com/let-z-go/intrusive_containers/list"
)

type Condition struct {
	lock           sync.Locker
	listOfWaiters1 list.List
	listOfWaiters2 list.List
}

func (self *Condition) Initialize(lock sync.Locker) {
	if self.lock != nil {
		panic(errors.New("toolkit: condition initialization"))
	}

	self.lock = lock
	self.listOfWaiters1.Initialize()
	self.listOfWaiters2.Initialize()
}

func (self *Condition) WaitFor(context_ context.Context) error {
	self.checkUninitialized()
	var waiter conditionWaiter
	waiter.event = make(chan struct{}, 1)
	self.listOfWaiters1.AppendNode(&waiter.listNode)
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
	self.checkUninitialized()

	if self.listOfWaiters1.IsEmpty() {
		return
	}

	waiter := (*conditionWaiter)(self.listOfWaiters1.GetHead().GetContainer(unsafe.Offsetof(conditionWaiter{}.listNode)))
	waiter.event <- struct{}{}
	waiter.listNode.Remove()
	self.listOfWaiters2.AppendNode(&waiter.listNode)
}

func (self *Condition) Broadcast() {
	self.checkUninitialized()
	getNode := self.listOfWaiters1.GetNodes()

	for listNode := getNode(); listNode != nil; listNode = getNode() {
		waiter := (*conditionWaiter)(listNode.GetContainer(unsafe.Offsetof(conditionWaiter{}.listNode)))
		waiter.event <- struct{}{}
	}

	self.listOfWaiters1.Append(&self.listOfWaiters2)
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
