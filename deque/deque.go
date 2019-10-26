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

func (d *Deque) Init(maxLength int) *Deque {
	d.semaphore.Init(0, maxLength, 0)
	d.list.Init()
	return d
}

func (d *Deque) Close(list *intrusive.List) (int, error) {
	n, err := d.semaphore.Close(func() {
		if list == nil {
			d.list.Init()
		} else {
			list.AppendNodes(&d.list)
		}
	})

	return n, convertSemaphoreError(err)
}

func (d *Deque) AppendNode(ctx context.Context, node *intrusive.ListNode) error {
	return convertSemaphoreError(d.semaphore.Up(ctx, false, func() {
		d.list.AppendNode(node)
	}))
}

func (d *Deque) PrependNode(ctx context.Context, node *intrusive.ListNode) error {
	return convertSemaphoreError(d.semaphore.Up(ctx, false, func() {
		d.list.PrependNode(node)
	}))
}

func (d *Deque) RemoveTail(ctx context.Context, withoutCommitment bool) (*intrusive.ListNode, error) {
	node := (*intrusive.ListNode)(nil)

	return node, convertSemaphoreError(d.semaphore.Down(ctx, withoutCommitment, func() {
		node = d.list.Tail()
		node.Remove()
	}))
}

func (d *Deque) RemoveHead(ctx context.Context, withoutCommitment bool) (*intrusive.ListNode, error) {
	node := (*intrusive.ListNode)(nil)

	return node, convertSemaphoreError(d.semaphore.Down(ctx, withoutCommitment, func() {
		node = d.list.Head()
		node.Remove()
	}))
}

func (d *Deque) CommitNodeRemoval() error {
	return convertSemaphoreError(d.semaphore.IncreaseMaxValue(1, false, nil))
}

func (d *Deque) DiscardNodeRemoval(node *intrusive.ListNode, appendNode bool) error {
	return convertSemaphoreError(d.semaphore.IncreaseMaxValue(1, true, func() {
		if appendNode {
			d.list.AppendNode(node)
		} else {
			d.list.PrependNode(node)
		}
	}))
}

func (d *Deque) RemoveNodes(ctx context.Context, withoutCommitment bool, list *intrusive.List) (int, error) {
	n, err := d.semaphore.DownAll(ctx, withoutCommitment, func() {
		list.AppendNodes(&d.list)
	})

	return n, convertSemaphoreError(err)
}

func (d *Deque) CommitNodesRemoval(numberOfNodes int) error {
	return convertSemaphoreError(d.semaphore.IncreaseMaxValue(numberOfNodes, false, nil))
}

func (d *Deque) DiscardNodesRemoval(list *intrusive.List, numberOfNodes int, appendNodes bool) error {
	return convertSemaphoreError(d.semaphore.IncreaseMaxValue(numberOfNodes, true, func() {
		if appendNodes {
			d.list.AppendNodes(list)
		} else {
			d.list.PrependNodes(list)
		}
	}))
}

func (d *Deque) IsClosed() bool {
	return d.semaphore.IsClosed()
}

func (d *Deque) MaxLength() int {
	return d.semaphore.MaxValue()
}

func (d *Deque) Length() int {
	return d.semaphore.Value()
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
