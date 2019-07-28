package utils

import (
	"context"
	"time"
)

func DelayContext(context_ context.Context, delay time.Duration) context.Context {
	if delay < 1 {
		return context_
	}

	context2, cancel2 := context.WithCancel(context.Background())

	go func() {
		<-context_.Done()
		time.Sleep(delay)
		cancel2()
	}()

	return context2
}
