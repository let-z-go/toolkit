package deque

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/let-z-go/intrusive"
)

type Foo struct {
	bar      int
	listNode intrusive.ListNode
}

func TestDeque1(t *testing.T) {
	d := new(Deque).Init(2)
	f := int32(0)

	go func() {
		d.AppendNode(context.Background(), &(&Foo{bar: 1}).listNode)
		atomic.AddInt32(&f, 1)
		d.AppendNode(context.Background(), &(&Foo{bar: 2}).listNode)
		atomic.AddInt32(&f, 1)
		d.AppendNode(context.Background(), &(&Foo{bar: 3}).listNode)
		atomic.AddInt32(&f, 1)
	}()

	time.Sleep(time.Second / 20)

	if f2 := atomic.LoadInt32(&f); f2 != 2 {
		t.Errorf("%#v", f2)
	}

	d.CommitNodesRemoval(1)
	time.Sleep(time.Second / 20)

	if f2 := atomic.LoadInt32(&f); f2 != 3 {
		t.Errorf("%#v", f2)
	}

	var ln *intrusive.ListNode
	ln, _ = d.RemoveHead(context.Background(), false)

	if f := (*Foo)(ln.GetContainer(unsafe.Offsetof(Foo{}.listNode))); f.bar != 1 {
		t.Errorf("%#v", f.bar)
	}

	ln, _ = d.RemoveHead(context.Background(), false)

	if f := (*Foo)(ln.GetContainer(unsafe.Offsetof(Foo{}.listNode))); f.bar != 2 {
		t.Errorf("%#v", f.bar)
	}

	ln, _ = d.RemoveHead(context.Background(), false)

	if f := (*Foo)(ln.GetContainer(unsafe.Offsetof(Foo{}.listNode))); f.bar != 3 {
		t.Errorf("%#v", f.bar)
	}

}

func TestDeque2(t *testing.T) {
	d := new(Deque).Init(3)
	var wg sync.WaitGroup
	ec := int32(0)

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(i int) {
			if err := d.AppendNode(context.Background(), &(&Foo{bar: i}).listNode); err != nil {
				if err != ErrDequeClosed {
					t.Errorf("%#v != %#v", err, ErrDequeClosed)
				}

				atomic.AddInt32(&ec, 1)
			}

			wg.Done()
		}(i)
	}

	time.Sleep(time.Second / 20)
	d.Close(nil)
	wg.Wait()

	if ec != 7 {
		t.Errorf("%#v", ec)
	}
}

func TestShrinkDeque(t *testing.T) {
	d := new(Deque).Init(8)

	for i := 0; i < 6; i++ {
		d.AppendNode(context.Background(), &(&Foo{bar: i}).listNode)
	}

	l := NewList()
	d.Shrink(4, false, l)
	if l.Length != 2 {
		t.Fatal(l.Length)
	}
	if d.Length() != 4 {
		t.Fatal(d.Length())
	}
	getNode := l.Underlying.GetNodes()

	for i, ln := 0, getNode(); ln != nil; i, ln = i+1, getNode() {
		if f := (*Foo)(ln.GetContainer(unsafe.Offsetof(Foo{}.listNode))); f.bar != i+4 {
			t.Errorf("%#v", f.bar)
		}
	}

	l = NewList()
	d.Shrink(2, true, l)
	if l.Length != 2 {
		t.Fatal(l.Length)
	}
	if d.Length() != 2 {
		t.Fatal(d.Length())
	}
	getNode = l.Underlying.GetNodes()

	for i, ln := 0, getNode(); ln != nil; i, ln = i+1, getNode() {
		if f := (*Foo)(ln.GetContainer(unsafe.Offsetof(Foo{}.listNode))); f.bar != i {
			t.Errorf("%#v", f.bar)
		}
	}
}
