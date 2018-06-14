package deque

import (
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/let-z-go/intrusive_containers"
)

type Foo struct {
	bar      int
	listNode intrusive_containers.ListNode
}

func TestDeque(t *testing.T) {
	d := Deque{}
	d.Initialize(2)
	f := int32(0)

	go func() {
		d.AppendNode(&(&Foo{bar: 1}).listNode, -1)
		atomic.AddInt32(&f, 1)
		d.AppendNode(&(&Foo{bar: 2}).listNode, -1)
		atomic.AddInt32(&f, 1)
		d.AppendNode(&(&Foo{bar: 3}).listNode, -1)
		atomic.AddInt32(&f, 1)
	}()

	time.Sleep(time.Second / 20)

	if f2 := atomic.LoadInt32(&f); f2 != 2 {
		t.Errorf("%#v", f2)
	}

	d.CommitNodeRemovals(1)
	time.Sleep(time.Second / 20)

	if f2 := atomic.LoadInt32(&f); f2 != 3 {
		t.Errorf("%#v", f2)
	}

	var ln *intrusive_containers.ListNode
	ln, _ = d.RemoveHead(true, -1)

	if f := (*Foo)(ln.GetContainer(unsafe.Offsetof(Foo{}.listNode))); f.bar != 1 {
		t.Errorf("%#v", f.bar)
	}

	ln, _ = d.RemoveHead(true, -1)

	if f := (*Foo)(ln.GetContainer(unsafe.Offsetof(Foo{}.listNode))); f.bar != 2 {
		t.Errorf("%#v", f.bar)
	}

	ln, _ = d.RemoveHead(true, -1)

	if f := (*Foo)(ln.GetContainer(unsafe.Offsetof(Foo{}.listNode))); f.bar != 3 {
		t.Errorf("%#v", f.bar)
	}

}
