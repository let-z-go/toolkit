package delay_pool

import (
	"fmt"
	"math/rand"
	"time"
)

type DelayPool struct {
	values             []interface{}
	numberOfItems      int
	maxDelay           time.Duration
	usedItemCount      int
	nextItemUsableTime time.Time
}

func (self *DelayPool) Reset(values []interface{}, numberOfItems int, maxTotalDelay time.Duration) {
	if len(values) == 0 {
		values = self.values
	}

	if len(values) == 0 {
		panic(fmt.Errorf("toolkit: delay pool reset: values=%#v", values))
	}

	if numberOfItems < 1 {
		numberOfItems = self.numberOfItems
	}

	if numberOfItems < 1 {
		panic(fmt.Errorf("toolkit: delay pool reset: numberOfItems=%#v", numberOfItems))
	}

	if maxTotalDelay < 1 {
		maxTotalDelay = time.Duration(self.numberOfItems) * self.maxDelay
	}

	if maxTotalDelay < 1 {
		panic(fmt.Errorf("toolkit: delay pool reset: maxTotalDelay=%#v", maxTotalDelay))
	}

	for i := int32(len(values)) - 1; i >= 1; i-- {
		j := rand.Int31n(i + 1)
		values[i], values[j] = values[j], values[i]
	}

	self.values = values
	self.numberOfItems = numberOfItems
	self.maxDelay = maxTotalDelay / time.Duration(numberOfItems)
	self.usedItemCount = 0
}

func (self *DelayPool) GetValue() (interface{}, bool) {
	if self.usedItemCount == self.numberOfItems {
		if !self.nextItemUsableTime.IsZero() {
			delay := self.nextItemUsableTime.Sub(time.Now())

			if delay >= time.Millisecond {
				time.Sleep(delay)
			}

			self.nextItemUsableTime = time.Time{}
		}

		return nil, false
	}

	var delay time.Duration

	if self.usedItemCount == 0 {
		self.nextItemUsableTime = time.Now()
		delay = time.Duration(0)
	} else {
		delay = self.nextItemUsableTime.Sub(time.Now())
	}

	value := self.values[self.usedItemCount%len(self.values)]
	self.usedItemCount += 1
	self.nextItemUsableTime = self.nextItemUsableTime.Add(self.maxDelay)

	if delay >= time.Millisecond {
		time.Sleep(delay)
	}

	return value, true
}
