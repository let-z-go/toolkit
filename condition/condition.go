package condition

import (
	"context"
	"sync"
	"unsafe"

	"github.com/let-z-go/intrusive"
)

type Condition struct {
	lock          sync.Locker
	listOfWaiters intrusive.List
}

func (c *Condition) Init(lock sync.Locker) *Condition {
	c.lock = lock
	c.listOfWaiters.Init()
	return c
}

func (c *Condition) WaitFor(ctx context.Context) (bool, error) {
	var waiter conditionWaiter
	waiter.Event = make(chan struct{})
	c.listOfWaiters.AppendNode(&waiter.ListNode)
	c.lock.Unlock()
	var err error

	select {
	case <-waiter.Event:
		err = nil
	case <-ctx.Done():
		err = ctx.Err()
	}

	c.lock.Lock()
	ok := err == nil || waiter.ListNode.IsReset()

	if !ok {
		waiter.ListNode.Remove()
	}

	return ok, err
}

func (c *Condition) Signal() {
	if c.listOfWaiters.IsEmpty() {
		return
	}

	waiter := (*conditionWaiter)(c.listOfWaiters.Head().GetContainer(unsafe.Offsetof(conditionWaiter{}.ListNode)))
	waiter.ListNode.Remove()
	waiter.ListNode.Reset()
	close(waiter.Event)
}

func (c *Condition) Broadcast() {
	getNode := c.listOfWaiters.GetNodesSafely()

	for listNode := getNode(); listNode != nil; listNode = getNode() {
		listNode.Reset()
		waiter := (*conditionWaiter)(listNode.GetContainer(unsafe.Offsetof(conditionWaiter{}.ListNode)))
		close(waiter.Event)
	}

	c.listOfWaiters.Init()
}

type conditionWaiter struct {
	ListNode intrusive.ListNode
	Event    chan struct{}
}
