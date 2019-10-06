package timerpool

import (
	"sync"
	"time"
)

func GetTimer(duration time.Duration) *time.Timer {
	if value := timerPool.Get(); value != nil {
		timer := value.(*time.Timer)
		timer.Reset(duration)
		return timer
	}

	return time.NewTimer(duration)
}

func PutTimer(timer *time.Timer) {
	timerPool.Put(timer)
}

func StopAndPutTimer(timer *time.Timer) {
	timer.Stop()
	PutTimer(timer)
}

var timerPool = sync.Pool{}
