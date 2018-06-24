package semaphore

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDownAndUpSemaphore1(t *testing.T) {
	var s Semaphore
	s.Initialize(0, 100, 50)
	var wg sync.WaitGroup
	sc := int32(0)
	fc := int32(0)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			e := s.Down(nil, false, nil)

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
	s.Up(nil, false, nil)
	s.Up(nil, false, nil)
	time.Sleep(time.Second / 10)
	s.Close(nil)
	wg.Wait()

	if sc != 52 {
		t.Errorf("%#v", sc)
	}

	if fc != 48 {
		t.Errorf("%#v", fc)
	}
}

func TestDownAndUpSemaphore2(t *testing.T) {
	var s Semaphore
	s.Initialize(0, 50, 0)
	var wg sync.WaitGroup
	sc := int32(0)
	fc := int32(0)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second/20)

			if e := s.Up(ctx, false, nil); e == nil {
				atomic.AddInt32(&sc, 1)
			} else {
				if e != context.DeadlineExceeded {
					t.Errorf("%#v != %#v", e, context.DeadlineExceeded)
				}

				atomic.AddInt32(&fc, 1)
			}

			cancel()
			wg.Done()
		}(i)
	}

	time.Sleep(time.Second / 10)
	s.Down(nil, false, nil)
	s.Down(nil, false, nil)
	time.Sleep(time.Second / 10)
	wg.Wait()

	if sc != 50 {
		t.Errorf("%#v", sc)
	}

	if fc != 50 {
		t.Errorf("%#v", fc)
	}
}
