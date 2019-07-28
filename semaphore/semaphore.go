package semaphore

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/let-z-go/toolkit/condition"
	"github.com/let-z-go/toolkit/utils"
)

type Semaphore struct {
	lock              sync.Mutex
	minValue          int32
	maxValue          int32
	value             int32
	upWaiterCountX2   uint32
	downWaiterCountX2 uint32
	upCondition       condition.Condition
	downCondition     condition.Condition
	isClosed          int32
}

func (self *Semaphore) Initialize(minValue int32, maxValue int32, value int32) *Semaphore {
	utils.Assert(value >= minValue && value <= maxValue, func() string {
		return fmt.Sprintf("toolkit/semaphore: invalid argument: value=%#v, minValue=%#v, maxValue=%#v", value, minValue, maxValue)
	})

	self.minValue = minValue
	self.maxValue = maxValue
	self.value = value
	self.upCondition.Initialize(&self.lock)
	self.downCondition.Initialize(&self.lock)
	return self
}

func (self *Semaphore) Close(callback func()) error {
	if !atomic.CompareAndSwapInt32(&self.isClosed, 0, 1) {
		return SemaphoreClosedError
	}

	self.lock.Lock()

	if callback != nil {
		callback()
	}

	self.upCondition.Broadcast()
	self.downCondition.Broadcast()
	self.lock.Unlock()
	return nil
}

func (self *Semaphore) Up(context_ context.Context, increaseMinValue bool, callback func()) error {
	_, e := self.doUp(context_, false, increaseMinValue, callback)
	return e
}

func (self *Semaphore) UpAll(context_ context.Context, increaseMinValue bool, callback func()) (int32, error) {
	return self.doUp(context_, true, increaseMinValue, callback)
}

func (self *Semaphore) Down(context_ context.Context, decreaseMaxValue bool, callback func()) error {
	_, e := self.doDown(context_, false, decreaseMaxValue, callback)
	return e
}

func (self *Semaphore) DownAll(context_ context.Context, decreaseMaxValue bool, callback func()) (int32, error) {
	return self.doDown(context_, true, decreaseMaxValue, callback)
}

func (self *Semaphore) IncreaseMaxValue(increment int32, increaseValue bool, callback func()) error {
	if self.IsClosed() {
		return SemaphoreClosedError
	}

	if increment < 1 {
		return nil
	}

	self.lock.Lock()

	if callback != nil {
		callback()
	}

	self.maxValue += increment

	if increaseValue {
		self.value += increment

		if self.value-increment == self.minValue {
			self.notifyDownWaiter()
		}
	} else {
		if self.value == self.maxValue-increment {
			self.notifyUpWaiter()
		}
	}

	self.lock.Unlock()
	return nil
}

func (self *Semaphore) DecreaseMinValue(decrement int32, decreaseValue bool, callback func()) error {
	if self.IsClosed() {
		return SemaphoreClosedError
	}

	if decrement < 1 {
		return nil
	}

	self.lock.Lock()

	if callback != nil {
		callback()
	}

	self.minValue -= decrement

	if decreaseValue {
		self.value -= decrement

		if self.value+decrement == self.maxValue {
			self.notifyUpWaiter()
		}
	} else {
		if self.value == self.minValue+decrement {
			self.notifyDownWaiter()
		}
	}

	self.lock.Unlock()
	return nil
}

func (self *Semaphore) IsClosed() bool {
	return atomic.LoadInt32(&self.isClosed) == 1
}

func (self *Semaphore) doUp(context_ context.Context, maximizeIncrement bool, increaseMinValue bool, callback func()) (int32, error) {
	if self.IsClosed() {
		return 0, SemaphoreClosedError
	}

	self.lock.Lock()

	if self.IsClosed() {
		self.lock.Unlock()
		return 0, SemaphoreClosedError
	}

	if self.value == self.maxValue || self.upWaiterCountX2 >= 2 {
		self.upWaiterCountX2 += 2

		for {
			if ok, e := self.upCondition.WaitFor(context_); e != nil {
				self.upWaiterCountX2 -= 2

				if ok {
					self.upWaiterCountX2 &^= 1

					if self.value < self.maxValue {
						self.notifyUpWaiter()
					}
				}

				self.lock.Unlock()
				return 0, e
			}

			if self.IsClosed() {
				self.lock.Unlock()
				return 0, SemaphoreClosedError
			}

			self.upWaiterCountX2 &^= 1

			if self.value < self.maxValue {
				break
			}
		}

		self.upWaiterCountX2 -= 2
	}

	if callback != nil {
		callback()
	}

	var increment int32

	if maximizeIncrement {
		increment = self.maxValue - self.value
		self.value = self.maxValue
	} else {
		increment = 1
		self.value++

		if self.value < self.maxValue {
			self.notifyUpWaiter()
		}
	}

	if increaseMinValue {
		self.minValue += increment
	} else {
		if self.value-increment == self.minValue {
			self.notifyDownWaiter()
		}
	}

	self.lock.Unlock()
	return increment, nil
}

func (self *Semaphore) doDown(context_ context.Context, maximizeDecrement bool, decreaseMaxValue bool, callback func()) (int32, error) {
	if self.IsClosed() {
		return 0, SemaphoreClosedError
	}

	self.lock.Lock()

	if self.IsClosed() {
		self.lock.Unlock()
		return 0, SemaphoreClosedError
	}

	if self.value == self.minValue || self.downWaiterCountX2 >= 2 {
		self.downWaiterCountX2 += 2

		for {
			if ok, e := self.downCondition.WaitFor(context_); e != nil {
				self.downWaiterCountX2 -= 2

				if ok {
					self.downWaiterCountX2 &^= 1

					if self.value > self.minValue {
						self.notifyDownWaiter()
					}
				}

				self.lock.Unlock()
				return 0, e
			}

			if self.IsClosed() {
				self.lock.Unlock()
				return 0, SemaphoreClosedError
			}

			self.downWaiterCountX2 &^= 1

			if self.value > self.minValue {
				break
			}
		}

		self.downWaiterCountX2 -= 2
	}

	if callback != nil {
		callback()
	}

	var decrement int32

	if maximizeDecrement {
		decrement = self.value - self.minValue
		self.value = self.minValue
	} else {
		decrement = 1
		self.value--

		if self.value > self.minValue {
			self.notifyDownWaiter()
		}
	}

	if decreaseMaxValue {
		self.maxValue -= decrement
	} else {
		if self.value+decrement == self.maxValue {
			self.notifyUpWaiter()
		}
	}

	self.lock.Unlock()
	return decrement, nil
}

func (self *Semaphore) notifyUpWaiter() {
	if self.upWaiterCountX2 >= 2 && self.upWaiterCountX2&1 == 0 {
		self.upCondition.Signal()
		self.upWaiterCountX2 |= 1
	}
}

func (self *Semaphore) notifyDownWaiter() {
	if self.downWaiterCountX2 >= 2 && self.downWaiterCountX2&1 == 0 {
		self.downCondition.Signal()
		self.downWaiterCountX2 |= 1
	}
}

var SemaphoreClosedError = errors.New("toolkit/semaphore: semaphore closed")
