package semaphore

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

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

func (self *Semaphore) Down(context_ context.Context, decreaseMaxValue bool, callback func()) error {
	return self.doDown(context_, false, decreaseMaxValue, callback)
}

func (self *Semaphore) DownAll(context_ context.Context, decreaseMaxValue bool, callback func()) error {
	return self.doDown(context_, true, decreaseMaxValue, callback)
}

func (self *Semaphore) Up(context_ context.Context, increaseMinValue bool, callback func()) error {
	return self.doUp(context_, false, increaseMinValue, callback)
}

func (self *Semaphore) UpAll(context_ context.Context, increaseMinValue bool, callback func()) error {
	return self.doUp(context_, true, increaseMinValue, callback)
}

func (self *Semaphore) IncreaseMaxValue(increment int32, callback func()) error {
	if self.IsClosed() {
		return SemaphoreClosedError
	}

	if increment < 1 {
		return nil
	}

	self.lock.Lock()
	maxValue := self.maxValue
	self.maxValue += increment

	if self.value == maxValue {
		self.notifyUpWaiter()
	}

	if callback != nil {
		callback()
	}

	self.lock.Unlock()
	return nil
}

func (self *Semaphore) DecreaseMinValue(decrement int32, callback func()) error {
	if self.IsClosed() {
		return SemaphoreClosedError
	}

	if decrement < 1 {
		return nil
	}

	self.lock.Lock()
	minValue := self.minValue
	self.minValue += decrement

	if self.value == minValue {
		self.notifyDownWaiter()
	}

	if callback != nil {
		callback()
	}

	self.lock.Unlock()
	return nil
}

func (self *Semaphore) Close(callback func()) error {
	if !atomic.CompareAndSwapInt32(&self.isClosed, 0, -1) {
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

func (self *Semaphore) IsClosed() bool {
	return atomic.LoadInt32(&self.isClosed) != 0
}

func (self *Semaphore) doDown(context_ context.Context, all bool, decreaseMaxValue bool, callback func()) error {
	if self.IsClosed() {
		return SemaphoreClosedError
	}

	self.lock.Lock()

	if self.IsClosed() {
		self.lock.Unlock()
		return SemaphoreClosedError
	}

	if self.value == self.minValue {
		self.downWaiterCountX2 += 2

		for {
			if e := self.downCondition.WaitFor(context_); e != nil {
				self.lock.Unlock()
				return e
			}

			if self.IsClosed() {
				self.lock.Unlock()
				return SemaphoreClosedError
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

func (self *Semaphore) doUp(context_ context.Context, all bool, increaseMinValue bool, callback func()) error {
	if self.IsClosed() {
		return SemaphoreClosedError
	}

	self.lock.Lock()

	if self.IsClosed() {
		self.lock.Unlock()
		return SemaphoreClosedError
	}

	if self.value == self.maxValue {
		self.upWaiterCountX2 += 2

		for {
			if e := self.upCondition.WaitFor(context_); e != nil {
				self.lock.Unlock()
				return e
			}

			if self.IsClosed() {
				self.lock.Unlock()
				return SemaphoreClosedError
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
