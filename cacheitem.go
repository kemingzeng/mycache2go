package mycache2go

import (
	"sync"
	"time"
)

type CacheItem struct {
	// mutex
	sync.RWMutex

	// data related
	key   interface{}
	value interface{}

	// time related
	lifeSpan   time.Duration
	createdOn  time.Time
	accessedOn time.Time

	// statistic
	accessedCount uint64

	// callbacks
	doWhenExpire []func(interface{}, interface{})
}

func NewCacheItem(key, value interface{}, span time.Duration) *CacheItem {
	t := time.Now()
	return &CacheItem{
		key:           key,
		value:         value,
		lifeSpan:      span,
		createdOn:     t,
		accessedOn:    t,
		accessedCount: 0,
		doWhenExpire:  nil,
	}
}

func (item *CacheItem) Key() interface{} {
	return item.key
}

func (item *CacheItem) Value() interface{} {
	return item.value
}

func (item *CacheItem) LifeSpan() time.Duration {
	return item.lifeSpan
}

func (item *CacheItem) AddExpireCallbacks(f func(interface{}, interface{})) {
	item.Lock()
	defer item.Unlock()
	item.doWhenExpire = append(item.doWhenExpire, f)
}

func (item *CacheItem) ClearExpireCallbacks() {
	item.Lock()
	defer item.Unlock()
	item.doWhenExpire = nil
}

func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedOn = time.Now()
	item.accessedCount++
}
