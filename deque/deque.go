package deque

import (
	"context"
	"errors"

	"github.com/let-z-go/intrusive"

	"github.com/let-z-go/toolkit/semaphore"
)

type Deque struct {
	semaphore semaphore.Semaphore
	list      intrusive.List
}

func (d *Deque) Init(capacity int) *Deque {
	d.semaphore.Init(0, capacity, 0)
	d.list.Init()
	return d
}

func (d *Deque) Close(list *List) error {
	_, err := d.semaphore.Close(func(numberOfNodes int) {
		if list != nil {
			list.Underlying.AppendNodes(&d.list)
			list.Length += numberOfNodes
		}
	})

	return err
}

func (d *Deque) AppendNode(ctx context.Context, node *intrusive.ListNode) error {
	return convertSemaphoreError(d.semaphore.Up(ctx, false, func(int) {
		d.list.AppendNode(node)
	}))
}

func (d *Deque) PrependNode(ctx context.Context, node *intrusive.ListNode) error {
	return convertSemaphoreError(d.semaphore.Up(ctx, false, func(int) {
		d.list.PrependNode(node)
	}))
}

func (d *Deque) RemoveTail(ctx context.Context, withoutCommitment bool) (*intrusive.ListNode, error) {
	node := (*intrusive.ListNode)(nil)

	return node, convertSemaphoreError(d.semaphore.Down(ctx, withoutCommitment, func(int) {
		node = d.list.Tail()
		node.Remove()
	}))
}

func (d *Deque) RemoveHead(ctx context.Context, withoutCommitment bool) (*intrusive.ListNode, error) {
	node := (*intrusive.ListNode)(nil)

	return node, convertSemaphoreError(d.semaphore.Down(ctx, withoutCommitment, func(int) {
		node = d.list.Head()
		node.Remove()
	}))
}

func (d *Deque) CommitNodeRemoval() error {
	return convertSemaphoreError(d.semaphore.IncreaseMaxValue(1, false, nil))
}

func (d *Deque) DiscardNodeRemoval(node *intrusive.ListNode, prependNode bool) error {
	return convertSemaphoreError(d.semaphore.IncreaseMaxValue(1, true, func() {
		if prependNode {
			d.list.PrependNode(node)
		} else {
			d.list.AppendNode(node)
		}
	}))
}

func (d *Deque) RemoveNodes(ctx context.Context, withoutCommitment bool, list *List) error {
	_, err := d.semaphore.DownAll(ctx, withoutCommitment, func(numberOfNodes int) {
		list.Underlying.AppendNodes(&d.list)
		list.Length += numberOfNodes
	})

	return convertSemaphoreError(err)
}

func (d *Deque) CommitNodesRemoval(numberOfNodes int) error {
	return convertSemaphoreError(d.semaphore.IncreaseMaxValue(numberOfNodes, false, nil))
}

func (d *Deque) DiscardNodesRemoval(list *List, prependNodes bool) error {
	return convertSemaphoreError(d.semaphore.IncreaseMaxValue(list.Length, true, func() {
		if prependNodes {
			d.list.PrependNodes(&list.Underlying)
		} else {
			d.list.AppendNodes(&list.Underlying)
		}
	}))
}

func (d *Deque) Shrink(capacityDecrement int, removeFront bool, list *List) error {
	_, err := d.semaphore.DecreaseMaxValue(capacityDecrement, func(lengthDelta int) {
		numberOfNodes := -lengthDelta

		if numberOfNodes >= 1 {
			var firstNode, lastNode *intrusive.ListNode

			if removeFront {
				firstNode = d.list.Head()
				lastNode = firstNode

				for i := 1; i < numberOfNodes; i++ {
					lastNode = lastNode.Next()
				}
			} else {
				lastNode = d.list.Tail()
				firstNode = lastNode

				for i := 1; i < numberOfNodes; i++ {
					firstNode = firstNode.Prev()
				}
			}

			intrusive.RemoveListRange(firstNode, lastNode)
			list.Underlying.AppendRange(firstNode, lastNode)
			list.Length += numberOfNodes
		}
	})

	return convertSemaphoreError(err)
}

func (d *Deque) IsClosed() bool {
	return d.semaphore.IsClosed()
}

func (d *Deque) Capacity() int {
	return d.semaphore.MaxValue()
}

func (d *Deque) Length() int {
	return d.semaphore.Value()
}

type List struct {
	Underlying intrusive.List
	Length     int
}

func NewList() *List {
	var l List
	l.Underlying.Init()
	return &l
}

func (l *List) Reset() *List {
	l.Underlying.Init()
	l.Length = 0
	return l
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
