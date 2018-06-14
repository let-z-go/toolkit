package condition

import (
	"errors"
	"sync"
	"time"
	"unsafe"

	"github.com/let-z-go/intrusive_containers"
)

type Condition struct {
	lock        sync.Locker
	waiterList1 intrusive_containers.List
	waiterList2 intrusive_containers.List
}

func (self *Condition) Initialize(lock sync.Locker) {
	self.lock = lock
	self.waiterList1.Initialize()
	self.waiterList2.Initialize()
}

func (self *Condition) WaitFor(timeout time.Duration) error {
	waiter := conditionWaiter{}
	waiter.event = make(chan byte, 1)
	self.waiterList1.AppendNode(&waiter.listNode)
	self.lock.Unlock()
	var e error

	if timeout < 0 {
		<-waiter.event
		e = nil
	} else {
		select {
		case <-waiter.event:
			e = nil
		case <-time.After(timeout):
			e = ConditionTimedOutError
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
	waiter.event <- 0
	waiter.listNode.Remove()
	self.waiterList2.AppendNode(&waiter.listNode)
}

func (self *Condition) Broadcast() {
	getNode := self.waiterList1.GetNodes()

	for listNode := getNode(); listNode != nil; listNode = getNode() {
		waiter := (*conditionWaiter)(listNode.GetContainer(unsafe.Offsetof(conditionWaiter{}.listNode)))
		waiter.event <- 0
	}

	self.waiterList1.Append(&self.waiterList2)
}

type conditionWaiter struct {
	listNode intrusive_containers.ListNode
	event    chan byte
}

var ConditionTimedOutError = errors.New("toolkit: condition timed out")
