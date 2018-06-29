package condition

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestCondition1(t *testing.T) {
	var m sync.Mutex
	var c Condition
	c.Initialize(&m)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		m.Lock()

		if ok, _ := c.WaitFor(nil); !ok {
			t.Error()
		}

		m.Unlock()
		wg.Done()
	}()

	time.Sleep(time.Second / 20)
	m.Lock()
	c.Signal()
	m.Unlock()
	wg.Wait()
}

func TestCondition2(t *testing.T) {
	var m sync.Mutex
	var c Condition
	c.Initialize(&m)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		m.Lock()

		if ok, _ := c.WaitFor(nil); !ok {
			t.Error()
		}

		m.Unlock()
		wg.Done()
	}()

	go func() {
		m.Lock()

		if ok, _ := c.WaitFor(nil); !ok {
			t.Error()
		}

		m.Unlock()
		wg.Done()
	}()

	time.Sleep(time.Second / 20)
	m.Lock()
	c.Broadcast()
	m.Unlock()
	wg.Wait()
}

func TestCondition3(t *testing.T) {
	var m sync.Mutex
	var c Condition
	c.Initialize(&m)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		m.Lock()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second/20)

		if _, e := c.WaitFor(ctx); e != context.DeadlineExceeded {
			t.Errorf("%#v != %#v", e, context.DeadlineExceeded)
		}

		cancel()
		m.Unlock()
		wg.Done()
	}()

	wg.Wait()
}
