package deque

import (
	"context"

	"github.com/let-z-go/intrusive_containers/list"
	"github.com/let-z-go/toolkit/semaphore"
)

type Deque struct {
	semaphore semaphore.Semaphore
	list      list.List
}

func (self *Deque) Initialize(maxNumberOfNodes int32) {
	self.semaphore.Initialize(0, maxNumberOfNodes, 0)
	self.list.Initialize()
}

func (self *Deque) AppendNode(context_ context.Context, node *list.ListNode) error {
	return self.semaphore.Up(context_, false, func() {
		self.list.AppendNode(node)
	})
}

func (self *Deque) PrependNode(context_ context.Context, node *list.ListNode) error {
	return self.semaphore.Up(context_, false, func() {
		self.list.PrependNode(node)
	})
}

func (self *Deque) RemoveTail(context_ context.Context, commitNodeRemoval bool) (*list.ListNode, error) {
	node := (*list.ListNode)(nil)

	return node, self.semaphore.Down(context_, !commitNodeRemoval, func() {
		node = self.list.GetTail()
		node.Remove()
	})
}

func (self *Deque) RemoveHead(context_ context.Context, commitNodeRemoval bool) (*list.ListNode, error) {
	node := (*list.ListNode)(nil)

	return node, self.semaphore.Down(context_, !commitNodeRemoval, func() {
		node = self.list.GetHead()
		node.Remove()
	})
}

func (self *Deque) RemoveAllNodes(context_ context.Context, commitNodeRemovals bool, list_ *list.List) (int32, error) {
	return self.semaphore.DownAll(context_, !commitNodeRemovals, func() {
		self.list.Append(list_)
	})
}

func (self *Deque) CommitNodeRemovals(numberOfNodes int32) error {
	return self.semaphore.IncreaseMaxValue(numberOfNodes, false, nil)
}

func (self *Deque) DiscardNodeRemovals(list_ *list.List, numberOfNodes int32) error {
	return self.semaphore.IncreaseMaxValue(numberOfNodes, true, func() {
		list_.Prepend(&self.list)
	})
}

func (self *Deque) Close(list_ *list.List) error {
	return self.semaphore.Close(func() {
		if list_ != nil {
			self.list.Append(list_)
		}
	})
}

func (self *Deque) IsClosed() bool {
	return self.semaphore.IsClosed()
}
