package semaphore

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/let-z-go/toolkit/condition"
)

type Semaphore struct {
	minValue          int32
	maxValue          int32
	value             int32
	lock              sync.Mutex
	downWaiterCountX2 uint32
	upWaiterCountX2   uint32
	downCondition     condition.Condition
	upCondition       condition.Condition
	isClosed          int32
	closingError      error
}

func (self *Semaphore) Initialize(minValue int32, maxValue int32, value int32) {
	if value < minValue || value > maxValue {
		panic(invalidSemaphoreValueError{fmt.Sprintf("minValue=%#v, maxValue=%#v, value=%#v", minValue, maxValue, value)})
	}

	self.minValue = minValue
	self.maxValue = maxValue
	self.value = value
	self.downCondition.Initialize(&self.lock)
	self.upCondition.Initialize(&self.lock)
}

func (self *Semaphore) Down(decreaseMaxValue bool, timeout time.Duration, callback func()) error {
	return self.doDown(false, decreaseMaxValue, timeout, callback)
}

func (self *Semaphore) DownAll(decreaseMaxValue bool, timeout time.Duration, callback func()) error {
	return self.doDown(true, decreaseMaxValue, timeout, callback)
}

func (self *Semaphore) Up(increaseMinValue bool, timeout time.Duration, callback func()) error {
	return self.doUp(false, increaseMinValue, timeout, callback)
}

func (self *Semaphore) UpAll(increaseMinValue bool, timeout time.Duration, callback func()) error {
	return self.doUp(true, increaseMinValue, timeout, callback)
}

func (self *Semaphore) IncreaseMaxValue(increment int32) {
	if self.IsClosed() {
		panic(SemaphoreClosedError)
	}

	if increment < 1 {
		return
	}

	maxValue := self.maxValue
	self.maxValue += increment

	if self.value == maxValue {
		self.notifyUpWaiter()
	}
}

func (self *Semaphore) DecreaseMinValue(decrement int32) {
	if self.IsClosed() {
		panic(SemaphoreClosedError)
	}

	if decrement < 1 {
		return
	}

	minValue := self.minValue
	self.minValue += decrement

	if self.value == minValue {
		self.notifyDownWaiter()
	}
}

func (self *Semaphore) Close(closingError error, callback func()) error {
	if !atomic.CompareAndSwapInt32(&self.isClosed, 0, -1) {
		panic(SemaphoreClosedError)
	}

	if closingError == nil {
		closingError = SemaphoreClosedError
	}

	self.lock.Lock()
	self.closingError = closingError

	if callback != nil {
		callback()
	}

	self.upCondition.Broadcast()
	self.downCondition.Broadcast()
	self.lock.Unlock()
	return nil
}

func (self *Semaphore) IsClosed() bool {
	return atomic.LoadInt32(&self.isClosed) != 0
}

func (self *Semaphore) doDown(all bool, decreaseMaxValue bool, timeout time.Duration, callback func()) error {
	if self.IsClosed() {
		panic(SemaphoreClosedError)
	}

	self.lock.Lock()

	if self.closingError != nil {
		return self.closingError
	}

	if self.value == self.minValue {
		self.downWaiterCountX2 += 2

		for {
			if e := self.downCondition.WaitFor(timeout); e == condition.ConditionTimedOutError {
				self.lock.Unlock()
				return SemaphoreTimedOutError
			}

			if self.closingError != nil {
				self.lock.Unlock()
				return self.closingError
			}

			if self.value > self.minValue {
				break
			}
		}

		self.downWaiterCountX2 = (self.downWaiterCountX2 - 2) &^ 1
	}

	var decrement int32

	if all {
		decrement = self.value - self.minValue
	} else {
		decrement = 1
	}

	self.value -= decrement

	if decreaseMaxValue {
		self.maxValue -= decrement
	}

	if callback != nil {
		callback()
	}

	if self.value > self.minValue {
		self.notifyDownWaiter()
	}

	if self.value == self.maxValue-1 {
		self.notifyUpWaiter()
	}

	self.lock.Unlock()
	return nil
}

func (self *Semaphore) doUp(all bool, increaseMinValue bool, timeout time.Duration, callback func()) error {
	if self.IsClosed() {
		panic(SemaphoreClosedError)
	}

	self.lock.Lock()

	if self.closingError != nil {
		return self.closingError
	}

	if self.value == self.maxValue {
		self.upWaiterCountX2 += 2

		for {
			if e := self.upCondition.WaitFor(timeout); e == condition.ConditionTimedOutError {
				self.lock.Unlock()
				return SemaphoreTimedOutError
			}

			if self.closingError != nil {
				self.lock.Unlock()
				return self.closingError
			}

			if self.value < self.maxValue {
				break
			}
		}

		self.upWaiterCountX2 = (self.upWaiterCountX2 - 2) &^ 1
	}

	var increment int32

	if all {
		increment = self.maxValue - self.value
	} else {
		increment = 1
	}

	self.value += increment

	if increaseMinValue {
		self.minValue += increment
	}

	if callback != nil {
		callback()
	}

	if self.value < self.maxValue {
		self.notifyUpWaiter()
	}

	if self.value == self.minValue+1 {
		self.notifyDownWaiter()
	}

	self.lock.Unlock()
	return nil
}

func (self *Semaphore) notifyDownWaiter() {
	if self.downWaiterCountX2 >= 2 && self.downWaiterCountX2&1 == 0 {
		self.downCondition.Signal()
		self.downWaiterCountX2 |= 1
	}
}

func (self *Semaphore) notifyUpWaiter() {
	if self.upWaiterCountX2 >= 2 && self.upWaiterCountX2&1 == 0 {
		self.upCondition.Signal()
		self.upWaiterCountX2 |= 1
	}
}

type invalidSemaphoreValueError struct {
	context string
}

func (self invalidSemaphoreValueError) Error() string {
	result := "toolkit: invalid semaphore value"

	if self.context != "" {
		result += ": " + self.context
	}

	return result
}

var SemaphoreClosedError = errors.New("toolkit: semaphore closed")
var SemaphoreTimedOutError = errors.New("toolkit: semaphore timed out")
