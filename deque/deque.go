package deque

import (
	"context"
	"errors"

	"github.com/let-z-go/intrusives/list"

	"github.com/let-z-go/toolkit/semaphore"
)

type Deque struct {
	semaphore semaphore.Semaphore
	list      list.List
}

func (self *Deque) Init(maxNumberOfNodes int) *Deque {
	self.semaphore.Init(0, maxNumberOfNodes, 0)
	self.list.Init()
	return self
}

func (self *Deque) Close(list_ *list.List) (int, error) {
	n, err := self.semaphore.Close(func() {
		if list_ == nil {
			self.list.Init()
		} else {
			list_.AppendNodes(&self.list)
		}
	})

	return n, convertSemaphoreError(err)
}

func (self *Deque) AppendNode(ctx context.Context, node *list.ListNode) error {
	return convertSemaphoreError(self.semaphore.Up(ctx, false, func() {
		self.list.AppendNode(node)
	}))
}

func (self *Deque) PrependNode(ctx context.Context, node *list.ListNode) error {
	return convertSemaphoreError(self.semaphore.Up(ctx, false, func() {
		self.list.PrependNode(node)
	}))
}

func (self *Deque) RemoveTail(ctx context.Context, withoutCommitment bool) (*list.ListNode, error) {
	node := (*list.ListNode)(nil)

	return node, convertSemaphoreError(self.semaphore.Down(ctx, withoutCommitment, func() {
		node = self.list.GetTail()
		node.Remove()
	}))
}

func (self *Deque) RemoveHead(ctx context.Context, withoutCommitment bool) (*list.ListNode, error) {
	node := (*list.ListNode)(nil)

	return node, convertSemaphoreError(self.semaphore.Down(ctx, withoutCommitment, func() {
		node = self.list.GetHead()
		node.Remove()
	}))
}

func (self *Deque) CommitNodeRemoval() error {
	return convertSemaphoreError(self.semaphore.IncreaseMaxValue(1, false, nil))
}

func (self *Deque) DiscardNodeRemoval(node *list.ListNode, appendNode bool) error {
	return convertSemaphoreError(self.semaphore.IncreaseMaxValue(1, true, func() {
		if appendNode {
			self.list.AppendNode(node)
		} else {
			self.list.PrependNode(node)
		}
	}))
}

func (self *Deque) RemoveNodes(ctx context.Context, withoutCommitment bool, list_ *list.List) (int, error) {
	n, err := self.semaphore.DownAll(ctx, withoutCommitment, func() {
		list_.AppendNodes(&self.list)
	})

	return n, convertSemaphoreError(err)
}

func (self *Deque) CommitNodesRemoval(numberOfNodes int) error {
	return convertSemaphoreError(self.semaphore.IncreaseMaxValue(numberOfNodes, false, nil))
}

func (self *Deque) DiscardNodesRemoval(list_ *list.List, numberOfNodes int, appendNodes bool) error {
	return convertSemaphoreError(self.semaphore.IncreaseMaxValue(numberOfNodes, true, func() {
		if appendNodes {
			self.list.AppendNodes(list_)
		} else {
			self.list.PrependNodes(list_)
		}
	}))
}

func (self *Deque) GetMaxLength() int {
	return self.semaphore.GetMaxValue()
}

func (self *Deque) GetLength() int {
	return self.semaphore.GetValue()
}

func (self *Deque) IsClosed() bool {
	return self.semaphore.IsClosed()
}

var ErrDequeClosed = errors.New("toolkit/deque: deque closed")

func convertSemaphoreError(err error) error {
	if err != nil {
		switch err {
		case semaphore.ErrSemaphoreClosed:
			return ErrDequeClosed
		default:
			return err
		}
	}

	return nil
}
