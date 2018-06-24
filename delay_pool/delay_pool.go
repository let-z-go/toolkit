package delay_pool

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type DelayPool struct {
	values              []interface{}
	numberOfValues      int
	maxDelay            time.Duration
	usedValueCount      int
	nextValueUsableTime time.Time
}

func (self *DelayPool) Reset(values []interface{}, numberOfValues int, maxTotalDelay time.Duration) {
	if values == nil {
		values = self.values
	}

	if len(values) == 0 {
		panic(fmt.Errorf("toolkit: delay pool reset: values=%#v", values))
	}

	if numberOfValues < 1 {
		numberOfValues = self.numberOfValues
	}

	if numberOfValues < 1 {
		panic(fmt.Errorf("toolkit: delay pool reset: numberOfValues=%#v", numberOfValues))
	}

	if maxTotalDelay < 1 {
		maxTotalDelay = time.Duration(self.numberOfValues) * self.maxDelay
	}

	if maxTotalDelay < 1 {
		panic(fmt.Errorf("toolkit: delay pool reset: maxTotalDelay=%#v", maxTotalDelay))
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := int32(len(values)) - 1; i >= 1; i-- {
		j := random.Int31n(i + 1)
		values[i], values[j] = values[j], values[i]
	}

	self.values = values
	self.numberOfValues = numberOfValues
	self.maxDelay = maxTotalDelay / time.Duration(numberOfValues)
	self.usedValueCount = 0
}

func (self *DelayPool) Collect() {
	self.values = nil
}

func (self *DelayPool) GetValue(context_ context.Context) (interface{}, error) {
	if self.usedValueCount == self.numberOfValues {
		if !self.nextValueUsableTime.IsZero() {
			delay := self.nextValueUsableTime.Sub(time.Now())

			if delay >= time.Millisecond {
				if context_ == nil {
					time.Sleep(delay)
				} else {
					select {
					case <-time.After(delay):
					case <-context_.Done():
						return nil, context_.Err()
					}
				}
			}

			self.nextValueUsableTime = time.Time{}
		}

		return nil, NoMoreValuesError
	}

	var delay time.Duration

	if self.usedValueCount == 0 {
		self.nextValueUsableTime = time.Now()
		delay = time.Duration(0)
	} else {
		delay = self.nextValueUsableTime.Sub(time.Now())
	}

	value := self.values[self.usedValueCount%len(self.values)]
	self.usedValueCount += 1
	self.nextValueUsableTime = self.nextValueUsableTime.Add(self.maxDelay)

	if delay >= time.Millisecond {
		if context_ == nil {
			time.Sleep(delay)
		} else {
			select {
			case <-time.After(delay):
			case <-context_.Done():
				return nil, context_.Err()
			}
		}
	}

	return value, nil
}

func (self *DelayPool) WhenNextValueUsable() time.Time {
	return self.nextValueUsableTime
}

var NoMoreValuesError = errors.New("toolkit: no more values")
