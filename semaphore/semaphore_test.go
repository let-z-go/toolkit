package semaphore

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSemaphore1(t *testing.T) {
	s := new(Semaphore).Init(0, 100, 50)
	var wg sync.WaitGroup
	sc := int32(0)
	fc := int32(0)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			err := s.Down(context.Background(), false, nil)

			if err == nil {
				atomic.AddInt32(&sc, 1)
			} else {
				if err != ErrSemaphoreClosed {
					t.Errorf("%#v != %#v", err, ErrSemaphoreClosed)
				}

				atomic.AddInt32(&fc, 1)
			}

			wg.Done()
		}(i)
	}

	time.Sleep(time.Second / 10)
	s.Up(context.Background(), false, nil)
	s.Up(context.Background(), false, nil)
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

func TestSemaphore2(t *testing.T) {
	s := new(Semaphore).Init(0, 50, 0)
	var wg sync.WaitGroup
	sc := int32(0)
	fc := int32(0)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second/20)

			if err := s.Up(ctx, false, nil); err == nil {
				atomic.AddInt32(&sc, 1)
			} else {
				if err != context.DeadlineExceeded {
					t.Errorf("%#v != %#v", err, context.DeadlineExceeded)
				}

				atomic.AddInt32(&fc, 1)
			}

			cancel()
			wg.Done()
		}(i)
	}

	time.Sleep(time.Second / 10)
	s.Down(context.Background(), false, nil)
	s.Down(context.Background(), false, nil)
	time.Sleep(time.Second / 10)
	wg.Wait()

	if sc != 50 {
		t.Errorf("%#v", sc)
	}

	if fc != 50 {
		t.Errorf("%#v", fc)
	}
}

func TestSemaphore3(t *testing.T) {
	{
		s := new(Semaphore).Init(0, 10, 10)

		go func() {
			time.Sleep(time.Second / 10)
			s.DownAll(context.Background(), false, nil)
		}()

		s.Up(context.Background(), false, nil)
	}

	{
		s := new(Semaphore).Init(0, 10, 0)

		go func() {
			time.Sleep(time.Second / 10)
			s.UpAll(context.Background(), false, nil)
		}()

		s.Down(context.Background(), false, nil)
	}
}

func TestSemaphore4(t *testing.T) {
	{
		s := new(Semaphore).Init(0, 10, 10)

		go func() {
			time.Sleep(time.Second / 10)
			s.DecreaseMinValue(1, true, nil)
		}()

		s.Up(context.Background(), false, nil)
	}

	{
		s := new(Semaphore).Init(0, 10, 0)

		go func() {
			time.Sleep(time.Second / 10)
			s.IncreaseMaxValue(1, true, nil)
		}()

		s.Down(context.Background(), false, nil)
	}
}

func TestSemaphore5(t *testing.T) {
	{
		s := new(Semaphore).Init(0, 10, 10)

		go func() {
			time.Sleep(time.Second / 10)
			s.Down(context.Background(), false, nil)
			s.Up(context.Background(), false, nil)
			time.Sleep(time.Second / 10)
			s.Down(context.Background(), false, nil)
		}()

		s.Up(context.Background(), false, nil)
	}

	{
		s := new(Semaphore).Init(0, 10, 10)

		go func() {
			time.Sleep(time.Second / 10)
			s.Up(context.Background(), false, nil)
			s.Down(context.Background(), false, nil)
			time.Sleep(time.Second / 10)
			s.Up(context.Background(), false, nil)
		}()

		s.Down(context.Background(), false, nil)
	}
}

func TestSemaphore6(t *testing.T) {
	{
		s := new(Semaphore).Init(0, 10, 10)
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(time.Second / 10)
			cancel()
			s.Down(context.Background(), false, nil)
		}()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			go func() {
				s.Up(context.Background(), false, nil)
				wg.Done()
			}()

			if err := s.Up(ctx, false, nil); err == nil {
				s.Down(context.Background(), false, nil)
			}

			wg.Done()
		}()

		wg.Wait()
	}
}
