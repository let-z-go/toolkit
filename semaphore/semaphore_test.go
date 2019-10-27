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

func TestSemaphore7(t *testing.T) {
	{
		s := new(Semaphore).Init(0, 10, 9)
		d, _ := s.DecreaseMaxValue(1, nil)
		if d != 0 {
			t.Fatal(d)
		}
		f := int32(0)
		go func() {
			s.Up(context.Background(), false, nil)
			atomic.StoreInt32(&f, 1)
		}()
		time.Sleep(100 * time.Millisecond)
		if f2 := atomic.LoadInt32(&f); f2 != 0 {
			t.Fatal(f2)
		}
		d, _ = s.DecreaseMaxValue(1, nil)
		if d != -1 {
			t.Fatal(d)
		}
		time.Sleep(100 * time.Millisecond)
		if f2 := atomic.LoadInt32(&f); f2 != 0 {
			t.Fatal(f2)
		}
		if v := s.Value(); v != 8 {
			t.Fatal(v)
		}
		s.IncreaseMaxValue(1, false, nil)
		if v := s.MaxValue(); v != 9 {
			t.Fatal(v)
		}
		time.Sleep(100 * time.Millisecond)
		if f2 := atomic.LoadInt32(&f); f2 != 1 {
			t.Fatal(f2)
		}
	}
	{
		s := new(Semaphore).Init(0, 10, 1)
		d, _ := s.IncreaseMinValue(1, nil)
		if d != 0 {
			t.Fatal(d)
		}
		f := int32(0)
		go func() {
			s.Down(context.Background(), false, nil)
			atomic.StoreInt32(&f, 1)
		}()
		time.Sleep(100 * time.Millisecond)
		if f2 := atomic.LoadInt32(&f); f2 != 0 {
			t.Fatal(f2)
		}
		d, _ = s.IncreaseMinValue(1, nil)
		if d != 1 {
			t.Fatal(d)
		}
		time.Sleep(100 * time.Millisecond)
		if f2 := atomic.LoadInt32(&f); f2 != 0 {
			t.Fatal(f2)
		}
		if v := s.Value(); v != 2 {
			t.Fatal(v)
		}
		s.DecreaseMinValue(1, false, nil)
		if v := s.MinValue(); v != 1 {
			t.Fatal(v)
		}
		time.Sleep(100 * time.Millisecond)
		if f2 := atomic.LoadInt32(&f); f2 != 1 {
			t.Fatal(f2)
		}
	}
}
