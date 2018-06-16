package condition

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestWaitAndSignalCondition1(t *testing.T) {
	m := sync.Mutex{}
	c := Condition{}
	c.Initialize(&m)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		m.Lock()
		c.WaitFor(nil)
		m.Unlock()
		wg.Done()
	}()

	time.Sleep(time.Second / 20)
	m.Lock()
	c.Signal()
	m.Unlock()
	wg.Wait()
}

func TestWaitAndSignalCondition2(t *testing.T) {
	m := sync.Mutex{}
	c := Condition{}
	c.Initialize(&m)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		m.Lock()
		c.WaitFor(nil)
		m.Unlock()
		wg.Done()
	}()

	go func() {
		m.Lock()
		c.WaitFor(nil)
		m.Unlock()
		wg.Done()
	}()

	time.Sleep(time.Second / 20)
	m.Lock()
	c.Broadcast()
	m.Unlock()
	wg.Wait()
}

func TestWaitAndSignalCondition3(t *testing.T) {
	m := sync.Mutex{}
	c := Condition{}
	c.Initialize(&m)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		m.Lock()
		ctx, _ := context.WithTimeout(context.Background(), time.Second/20)

		if e := c.WaitFor(ctx); e != context.DeadlineExceeded {
			t.Errorf("%#v != %#v", e, context.DeadlineExceeded)
		}

		m.Unlock()
		wg.Done()
	}()

	wg.Wait()
}
