package background_task

import (
	"context"
	"sync"
)

func Run(bgContext context.Context, bgTask func(context.Context)) func() {
	context_, cancel := context.WithCancel(bgContext)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		bgTask(context_)
		wg.Done()
	}()

	bgTaskCanceller := func() {
		cancel()
		wg.Wait()
	}

	return bgTaskCanceller
}