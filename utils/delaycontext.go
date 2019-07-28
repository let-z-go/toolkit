package utils

import (
	"context"
	"time"
)

func DelayContext(ctx context.Context, delay time.Duration) context.Context {
	if delay < 1 {
		return ctx
	}

	context2, cancel2 := context.WithCancel(context.Background())

	go func() {
		<-ctx.Done()
		time.Sleep(delay)
		cancel2()
	}()

	return context2
}
