package condition

import (
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
		c.WaitFor(-1)
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
		c.WaitFor(-1)
		m.Unlock()
		wg.Done()
	}()

	go func() {
		m.Lock()
		c.WaitFor(-1)
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
		e := c.WaitFor(time.Second / 20)

		if e != ConditionTimedOutError {
			t.Errorf("%#v != %#v", e, ConditionTimedOutError)
		}

		m.Unlock()
		wg.Done()
	}()

	wg.Wait()
}
