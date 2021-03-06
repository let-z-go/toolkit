package lazymap

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLazyMap(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup
	var lm LazyMap
	c := int32(0)

	for i := 1; i <= 100; i++ {
		wg.Add(1)

		go func(i int) {
			for i := 0; i < 100; i++ {
				v, vc, _ := lm.GetOrSetValue(context.Background(), "k", func(context.Context) (interface{}, error) {
					time.Sleep(time.Duration(rand.Intn(3)) * time.Millisecond)
					return "v", nil
				})

				if v != "v" {
					t.Errorf("%#v!=\"v\"", v)
				}

				if vc != nil {
					time.Sleep(time.Duration(rand.Intn(3)) * time.Millisecond)
					vc()
					atomic.AddInt32(&c, 1)
				}
			}
			wg.Done()
		}(i)
	}

	wg.Wait()

	if c == 0 {
		t.Errorf("%#v==0", c)
	}
}
