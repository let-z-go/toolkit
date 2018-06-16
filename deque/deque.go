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

func (self *Deque) RemoveAllNodes(context_ context.Context, commitNodeRemovals bool, nodeList *list.List) error {
	return self.semaphore.DownAll(context_, !commitNodeRemovals, func() {
		self.list.Append(nodeList)
	})
}

func (self *Deque) CommitNodeRemovals(numberOfNodeRemovals int32) error {
	return self.semaphore.IncreaseMaxValue(numberOfNodeRemovals, nil)
}

func (self *Deque) Close() error {
	return self.semaphore.Close(func() {
		self.list.Initialize()
	})
}
