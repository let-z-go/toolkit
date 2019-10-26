package lazymap

import (
	"context"
	"sync"
)

type LazyMap struct {
	base sync.Map
}

func (lm *LazyMap) GetOrSetValue(ctx context.Context, key interface{}, valueFactory ValueFactory) (interface{}, ValueClearer, error) {
	value, ok := lm.base.Load(key)

	if !ok {
		value, _ = lm.base.LoadOrStore(key, new(lock).Init())
	}

Fallback:
	lock_, ok := value.(*lock)

	if !ok {
		return value, nil, nil
	}

Fallback2:
	if err := lock_.Acquire(ctx); err != nil {
		return nil, nil, err
	}

	value, ok = lm.base.Load(key)

	if !ok {
		lock_.Release()
		value, _ = lm.base.LoadOrStore(key, new(lock).Init())
		goto Fallback
	}

	lock2, ok := value.(*lock)

	if !ok {
		lock_.Release()
		return value, nil, nil
	}

	if lock_ != lock2 {
		lock_.Release()
		lock_ = lock2
		goto Fallback2
	}

	var err error
	value, err = valueFactory(ctx)

	if err != nil {
		lock_.Release()
		return nil, nil, err
	}

	lm.base.Store(key, value)
	lock_.Release()
	once := sync.Once{}

	valueClearer := func() {
		once.Do(func() {
			lm.base.Delete(key)
		})
	}

	return value, valueClearer, nil
}

type ValueFactory func(ctx context.Context) (value interface{}, err error)
type ValueClearer func()

type lock chan struct{}

func (l *lock) Init() *lock {
	*l = make(chan struct{}, 1)
	return l
}

func (l lock) Acquire(ctx context.Context) error {
	select {
	case l <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l lock) Release() {
	<-l
}
