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
	lock            sync.Mutex
	minValue        int
	maxValue        int
	value           int
	upWaiterCount   uint
	downWaiterCount uint
	upCondition     condition.Condition
	downCondition   condition.Condition
	isClosed        int32
}

func (self *Semaphore) Init(minValue int, maxValue int, value int) *Semaphore {
	utils.Assert(value >= minValue && value <= maxValue, func() string {
		return fmt.Sprintf("toolkit/semaphore: invalid argument: value=%#v, minValue=%#v, maxValue=%#v", value, minValue, maxValue)
	})

	self.minValue = minValue
	self.maxValue = maxValue
	self.value = value
	self.upCondition.Init(&self.lock)
	self.downCondition.Init(&self.lock)
	return self
}

func (self *Semaphore) Close(callback func()) (int, error) {
	if !atomic.CompareAndSwapInt32(&self.isClosed, 0, 1) {
		return 0, ErrSemaphoreClosed
	}

	self.lock.Lock()

	if callback != nil {
		callback()
	}

	self.upCondition.Broadcast()
	self.downCondition.Broadcast()
	self.lock.Unlock()
	return self.value, nil
}

func (self *Semaphore) Up(ctx context.Context, increaseMinValue bool, callback func()) error {
	_, err := self.doUp(ctx, false, increaseMinValue, callback)
	return err
}

func (self *Semaphore) UpAll(ctx context.Context, increaseMinValue bool, callback func()) (int, error) {
	return self.doUp(ctx, true, increaseMinValue, callback)
}

func (self *Semaphore) Down(ctx context.Context, decreaseMaxValue bool, callback func()) error {
	_, err := self.doDown(ctx, false, decreaseMaxValue, callback)
	return err
}

func (self *Semaphore) DownAll(ctx context.Context, decreaseMaxValue bool, callback func()) (int, error) {
	return self.doDown(ctx, true, decreaseMaxValue, callback)
}

func (self *Semaphore) IncreaseMaxValue(increment int, increaseValue bool, callback func()) error {
	if self.IsClosed() {
		return ErrSemaphoreClosed
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

func (self *Semaphore) DecreaseMinValue(decrement int, decreaseValue bool, callback func()) error {
	if self.IsClosed() {
		return ErrSemaphoreClosed
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

func (self *Semaphore) GetMinValue() int {
	self.lock.Lock()
	minValue := self.minValue
	self.lock.Unlock()
	return minValue
}

func (self *Semaphore) GetMaxValue() int {
	self.lock.Lock()
	maxValue := self.maxValue
	self.lock.Unlock()
	return maxValue
}

func (self *Semaphore) GetValue() int {
	self.lock.Lock()
	value := self.value
	self.lock.Unlock()
	return value
}

func (self *Semaphore) IsClosed() bool {
	return atomic.LoadInt32(&self.isClosed) == 1
}

func (self *Semaphore) doUp(ctx context.Context, maximizeIncrement bool, increaseMinValue bool, callback func()) (int, error) {
	if self.IsClosed() {
		return 0, ErrSemaphoreClosed
	}

	self.lock.Lock()

	if self.IsClosed() {
		self.lock.Unlock()
		return 0, ErrSemaphoreClosed
	}

	if self.value == self.maxValue || self.upWaiterCount >= 1 {
		self.upWaiterCount += 1

		for {
			if ok, err := self.upCondition.WaitFor(ctx); err != nil {
				self.upWaiterCount -= 1

				if ok {
					self.upWaiterCount &^= flagWaiterNotified

					if self.value < self.maxValue {
						self.notifyUpWaiter()
					}
				}

				self.lock.Unlock()
				return 0, err
			}

			if self.IsClosed() {
				self.lock.Unlock()
				return 0, ErrSemaphoreClosed
			}

			self.upWaiterCount &^= flagWaiterNotified

			if self.value < self.maxValue {
				break
			}
		}

		self.upWaiterCount -= 1
	}

	if callback != nil {
		callback()
	}

	var increment int

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

func (self *Semaphore) doDown(ctx context.Context, maximizeDecrement bool, decreaseMaxValue bool, callback func()) (int, error) {
	if self.IsClosed() {
		return 0, ErrSemaphoreClosed
	}

	self.lock.Lock()

	if self.IsClosed() {
		self.lock.Unlock()
		return 0, ErrSemaphoreClosed
	}

	if self.value == self.minValue || self.downWaiterCount >= 1 {
		self.downWaiterCount += 1

		for {
			if ok, err := self.downCondition.WaitFor(ctx); err != nil {
				self.downWaiterCount -= 1

				if ok {
					self.downWaiterCount &^= flagWaiterNotified

					if self.value > self.minValue {
						self.notifyDownWaiter()
					}
				}

				self.lock.Unlock()
				return 0, err
			}

			if self.IsClosed() {
				self.lock.Unlock()
				return 0, ErrSemaphoreClosed
			}

			self.downWaiterCount &^= flagWaiterNotified

			if self.value > self.minValue {
				break
			}
		}

		self.downWaiterCount -= 1
	}

	if callback != nil {
		callback()
	}

	var decrement int

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
	if self.upWaiterCount&flagWaiterNotified == 0 && self.upWaiterCount >= 1 {
		self.upCondition.Signal()
		self.upWaiterCount |= flagWaiterNotified
	}
}

func (self *Semaphore) notifyDownWaiter() {
	if self.downWaiterCount&flagWaiterNotified == 0 && self.downWaiterCount >= 1 {
		self.downCondition.Signal()
		self.downWaiterCount |= flagWaiterNotified
	}
}

var ErrSemaphoreClosed = errors.New("toolkit/semaphore: semaphore closed")

const flagWaiterNotified = (^uint(0) >> 1) + 1
