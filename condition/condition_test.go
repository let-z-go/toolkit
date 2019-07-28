package condition

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestCondition1(t *testing.T) {
	var m sync.Mutex
	c := new(Condition).Init(&m)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		m.Lock()

		if ok, _ := c.WaitFor(context.Background()); !ok {
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
	c := new(Condition).Init(&m)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		m.Lock()

		if ok, _ := c.WaitFor(context.Background()); !ok {
			t.Error()
		}

		m.Unlock()
		wg.Done()
	}()

	go func() {
		m.Lock()

		if ok, _ := c.WaitFor(context.Background()); !ok {
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
	c := new(Condition).Init(&m)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		m.Lock()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second/20)

		if _, err := c.WaitFor(ctx); err != context.DeadlineExceeded {
			t.Errorf("%#v != %#v", err, context.DeadlineExceeded)
		}

		cancel()
		m.Unlock()
		wg.Done()
	}()

	wg.Wait()
}
