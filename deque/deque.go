package deque

import (
	"time"

	"github.com/let-z-go/intrusive_containers"
	"github.com/let-z-go/toolkit/semaphore"
)

type Deque struct {
	semaphore semaphore.Semaphore
	list      intrusive_containers.List
}

func (self *Deque) Initialize(maxNumberOfNodes int32) {
	self.semaphore.Initialize(0, maxNumberOfNodes, 0)
	self.list.Initialize()
}

func (self *Deque) AppendNode(node *intrusive_containers.ListNode, timeout time.Duration) error {
	return self.semaphore.Up(false, timeout, func() {
		self.list.AppendNode(node)
	})
}

func (self *Deque) PrependNode(node *intrusive_containers.ListNode, timeout time.Duration) error {
	return self.semaphore.Up(false, timeout, func() {
		self.list.PrependNode(node)
	})
}

func (self *Deque) RemoveTail(commitNodeRemoval bool, timeout time.Duration) (*intrusive_containers.ListNode, error) {
	node := (*intrusive_containers.ListNode)(nil)

	return node, self.semaphore.Down(!commitNodeRemoval, timeout, func() {
		node = self.list.GetTail()
		node.Remove()
	})
}

func (self *Deque) RemoveHead(commitNodeRemoval bool, timeout time.Duration) (*intrusive_containers.ListNode, error) {
	node := (*intrusive_containers.ListNode)(nil)

	return node, self.semaphore.Down(!commitNodeRemoval, timeout, func() {
		node = self.list.GetHead()
		node.Remove()
	})
}

func (self *Deque) RemoveAllNodes(commitNodeRemovals bool, timeout time.Duration, nodeList *intrusive_containers.List) error {
	return self.semaphore.DownAll(!commitNodeRemovals, timeout, func() {
		self.list.Append(nodeList)
	})
}

func (self *Deque) CommitNodeRemovals(numberOfNodeRemovals int32) {
	self.semaphore.IncreaseMaxValue(numberOfNodeRemovals)
}
