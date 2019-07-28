package lazymap

import (
	"sync"
)

type LazyMap struct {
	base sync.Map
}

func (self *LazyMap) GetOrSetValue(key interface{}, valueFactory func() (interface{}, error)) (interface{}, func(), error) {
retry:
	value, _ := self.base.LoadOrStore(key, &lock{})
	valueClearer := (func())(nil)

	if lock_, ok := value.(*lock); ok {
	retry2:
		lock_.Lock()
		value2, ok := self.base.Load(key)

		if !ok {
			lock_.Unlock()
			goto retry
		}

		if lock2, ok := value2.(*lock); ok {
			if lock2 != lock_ {
				lock_.Unlock()
				lock_ = lock2
				goto retry2
			}

			var err error
			value2, err = valueFactory()

			if err != nil {
				lock_.Unlock()
				return nil, nil, err
			}

			self.base.Store(key, value2)
			once := sync.Once{}

			valueClearer = func() {
				once.Do(func() {
					self.base.Delete(key)
				})
			}
		}

		lock_.Unlock()
		value = value2
	}

	return value, valueClearer, nil
}

type lock struct {
	sync.Mutex
}
