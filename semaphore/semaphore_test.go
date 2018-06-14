package semaphore

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDownAndUpSemaphore1(t *testing.T) {
	s := Semaphore{}
	s.Initialize(0, 100, 50)
	wg := sync.WaitGroup{}
	sc := int32(0)
	fc := int32(0)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			e := s.Down(false, -1, nil)

			if e == nil {
				atomic.AddInt32(&sc, 1)
			} else {
				if e != SemaphoreClosedError {
					t.Errorf("%#v != %#v", e, SemaphoreClosedError)
				}

				atomic.AddInt32(&fc, 1)
			}

			wg.Done()
		}(i)
	}

	time.Sleep(time.Second / 10)
	s.Up(false, -1, nil)
	s.Up(false, -1, nil)
	time.Sleep(time.Second / 10)
	s.Close(nil, nil)
	wg.Wait()

	if sc != 52 {
		t.Errorf("%#v != 52", sc)
	}

	if fc != 48 {
		t.Errorf("%#v != 48", fc)
	}
}

func TestDownAndUpSemaphore2(t *testing.T) {
	s := Semaphore{}
	s.Initialize(0, 100, 50)
	wg := sync.WaitGroup{}
	sc := int32(0)
	fc := int32(0)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			e := s.Down(false, time.Second/20, nil)

			if e == nil {
				atomic.AddInt32(&sc, 1)
			} else {
				if e != SemaphoreTimedOutError {
					t.Errorf("%#v != %#v", e, SemaphoreTimedOutError)
				}

				atomic.AddInt32(&fc, 1)
			}

			wg.Done()
		}(i)
	}

	time.Sleep(time.Second / 10)
	s.Up(false, -1, nil)
	s.Up(false, -1, nil)
	time.Sleep(time.Second / 10)
	wg.Wait()

	if sc != 50 {
		t.Errorf("%#v != 50", sc)
	}

	if fc != 50 {
		t.Errorf("%#v != 50", fc)
	}
}
