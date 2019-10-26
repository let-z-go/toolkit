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

func (s *Semaphore) Init(minValue int, maxValue int, value int) *Semaphore {
	utils.Assert(value >= minValue && value <= maxValue, func() string {
		return fmt.Sprintf("toolkit/semaphore: invalid argument: value=%#v, minValue=%#v, maxValue=%#v", value, minValue, maxValue)
	})

	s.minValue = minValue
	s.maxValue = maxValue
	s.value = value
	s.upCondition.Init(&s.lock)
	s.downCondition.Init(&s.lock)
	return s
}

func (s *Semaphore) Close(callback func()) (int, error) {
	if !atomic.CompareAndSwapInt32(&s.isClosed, 0, 1) {
		return 0, ErrSemaphoreClosed
	}

	s.lock.Lock()

	if callback != nil {
		callback()
	}

	s.upCondition.Broadcast()
	s.downCondition.Broadcast()
	s.lock.Unlock()
	return s.value, nil
}

func (s *Semaphore) Up(ctx context.Context, increaseMinValue bool, callback func()) error {
	_, err := s.doUp(ctx, false, increaseMinValue, callback)
	return err
}

func (s *Semaphore) UpAll(ctx context.Context, increaseMinValue bool, callback func()) (int, error) {
	return s.doUp(ctx, true, increaseMinValue, callback)
}

func (s *Semaphore) Down(ctx context.Context, decreaseMaxValue bool, callback func()) error {
	_, err := s.doDown(ctx, false, decreaseMaxValue, callback)
	return err
}

func (s *Semaphore) DownAll(ctx context.Context, decreaseMaxValue bool, callback func()) (int, error) {
	return s.doDown(ctx, true, decreaseMaxValue, callback)
}

func (s *Semaphore) IncreaseMaxValue(increment int, increaseValue bool, callback func()) error {
	if s.IsClosed() {
		return ErrSemaphoreClosed
	}

	if increment < 1 {
		return nil
	}

	s.lock.Lock()

	if callback != nil {
		callback()
	}

	s.maxValue += increment

	if increaseValue {
		s.value += increment

		if s.value-increment == s.minValue {
			s.notifyDownWaiter()
		}
	} else {
		if s.value == s.maxValue-increment {
			s.notifyUpWaiter()
		}
	}

	s.lock.Unlock()
	return nil
}

func (s *Semaphore) DecreaseMinValue(decrement int, decreaseValue bool, callback func()) error {
	if s.IsClosed() {
		return ErrSemaphoreClosed
	}

	if decrement < 1 {
		return nil
	}

	s.lock.Lock()

	if callback != nil {
		callback()
	}

	s.minValue -= decrement

	if decreaseValue {
		s.value -= decrement

		if s.value+decrement == s.maxValue {
			s.notifyUpWaiter()
		}
	} else {
		if s.value == s.minValue+decrement {
			s.notifyDownWaiter()
		}
	}

	s.lock.Unlock()
	return nil
}

func (s *Semaphore) IsClosed() bool {
	return atomic.LoadInt32(&s.isClosed) == 1
}

func (s *Semaphore) MinValue() int {
	s.lock.Lock()
	minValue := s.minValue
	s.lock.Unlock()
	return minValue
}

func (s *Semaphore) MaxValue() int {
	s.lock.Lock()
	maxValue := s.maxValue
	s.lock.Unlock()
	return maxValue
}

func (s *Semaphore) Value() int {
	s.lock.Lock()
	value := s.value
	s.lock.Unlock()
	return value
}

func (s *Semaphore) doUp(ctx context.Context, maximizeIncrement bool, increaseMinValue bool, callback func()) (int, error) {
	if s.IsClosed() {
		return 0, ErrSemaphoreClosed
	}

	s.lock.Lock()

	if s.IsClosed() {
		s.lock.Unlock()
		return 0, ErrSemaphoreClosed
	}

	if s.value == s.maxValue || s.upWaiterCount >= 1 {
		s.upWaiterCount++

		for {
			if ok, err := s.upCondition.WaitFor(ctx); err != nil {
				s.upWaiterCount--

				if ok {
					s.upWaiterCount &^= flagWaiterNotified

					if s.value < s.maxValue {
						s.notifyUpWaiter()
					}
				}

				s.lock.Unlock()
				return 0, err
			}

			if s.IsClosed() {
				s.lock.Unlock()
				return 0, ErrSemaphoreClosed
			}

			s.upWaiterCount &^= flagWaiterNotified

			if s.value < s.maxValue {
				break
			}
		}

		s.upWaiterCount--
	}

	if callback != nil {
		callback()
	}

	var increment int

	if maximizeIncrement {
		increment = s.maxValue - s.value
		s.value = s.maxValue
	} else {
		increment = 1
		s.value++

		if s.value < s.maxValue {
			s.notifyUpWaiter()
		}
	}

	if increaseMinValue {
		s.minValue += increment
	} else {
		if s.value-increment == s.minValue {
			s.notifyDownWaiter()
		}
	}

	s.lock.Unlock()
	return increment, nil
}

func (s *Semaphore) doDown(ctx context.Context, maximizeDecrement bool, decreaseMaxValue bool, callback func()) (int, error) {
	if s.IsClosed() {
		return 0, ErrSemaphoreClosed
	}

	s.lock.Lock()

	if s.IsClosed() {
		s.lock.Unlock()
		return 0, ErrSemaphoreClosed
	}

	if s.value == s.minValue || s.downWaiterCount >= 1 {
		s.downWaiterCount++

		for {
			if ok, err := s.downCondition.WaitFor(ctx); err != nil {
				s.downWaiterCount--

				if ok {
					s.downWaiterCount &^= flagWaiterNotified

					if s.value > s.minValue {
						s.notifyDownWaiter()
					}
				}

				s.lock.Unlock()
				return 0, err
			}

			if s.IsClosed() {
				s.lock.Unlock()
				return 0, ErrSemaphoreClosed
			}

			s.downWaiterCount &^= flagWaiterNotified

			if s.value > s.minValue {
				break
			}
		}

		s.downWaiterCount--
	}

	if callback != nil {
		callback()
	}

	var decrement int

	if maximizeDecrement {
		decrement = s.value - s.minValue
		s.value = s.minValue
	} else {
		decrement = 1
		s.value--

		if s.value > s.minValue {
			s.notifyDownWaiter()
		}
	}

	if decreaseMaxValue {
		s.maxValue -= decrement
	} else {
		if s.value+decrement == s.maxValue {
			s.notifyUpWaiter()
		}
	}

	s.lock.Unlock()
	return decrement, nil
}

func (s *Semaphore) notifyUpWaiter() {
	if s.upWaiterCount&flagWaiterNotified == 0 && s.upWaiterCount >= 1 {
		s.upCondition.Signal()
		s.upWaiterCount |= flagWaiterNotified
	}
}

func (s *Semaphore) notifyDownWaiter() {
	if s.downWaiterCount&flagWaiterNotified == 0 && s.downWaiterCount >= 1 {
		s.downCondition.Signal()
		s.downWaiterCount |= flagWaiterNotified
	}
}

var ErrSemaphoreClosed = errors.New("toolkit/semaphore: semaphore closed")

const flagWaiterNotified = (^uint(0) >> 1) + 1
