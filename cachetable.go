package mycache2go

import (
	"log"
	"sync"
	"time"
)

type CacheTable struct {
	sync.RWMutex

	// data related
	name  string
	items map[interface{}]*CacheItem

	// time related
	cleaner         *time.Timer
	cleanUpInterval time.Duration

	// log
	logger *log.Logger

	// callbacks
	doAfterAddItem     []func(interface{}, interface{})
	doBeforeDeleteItem []func(interface{}, interface{})
	loadDataFunc       func(interface{}, ...interface{}) *CacheItem
}

func NewCacheTable(name string) *CacheTable {
	return &CacheTable{
		name:  name,
		items: make(map[interface{}]*CacheItem),
	}
}

func (table *CacheTable) SetLogger(logger *log.Logger) {
	table.Lock()
	defer table.Unlock()
	table.logger = logger
}

func (table *CacheTable) AddDoAfterAddItem(f func(interface{}, interface{})) {
	table.Lock()
	defer table.Unlock()
	table.doAfterAddItem = append(table.doAfterAddItem, f)
}

func (table *CacheTable) ClearDoAfterAddItem() {
	table.Lock()
	defer table.Unlock()
	table.doAfterAddItem = nil
}

func (table *CacheTable) AddDoBeforeDeleteItem(f func(interface{}, interface{})) {
	table.Lock()
	defer table.Unlock()
	table.doBeforeDeleteItem = append(table.doBeforeDeleteItem, f)
}

func (table *CacheTable) ClearDoBeforeDeleteItem() {
	table.Lock()
	defer table.Unlock()
	table.doBeforeDeleteItem = nil
}

func (table *CacheTable) SetLoadDataFunc(f func(interface{}, ...interface{}) *CacheItem) {
	table.Lock()
	defer table.Unlock()
	table.loadDataFunc = f
}

func (table *CacheTable) ClearLoadDataFunc() {
	table.Lock()
	defer table.Unlock()
	table.loadDataFunc = nil
}

func (table *CacheTable) deleteInternal(k interface{}) {
	// call backs
	beforeFuncs := table.doBeforeDeleteItem
	item := table.items[k]
	table.Unlock()
	for _, beforeFunc := range beforeFuncs {
		beforeFunc(k, item.value) // key and value are immutable
	}
	table.Lock()
	delete(table.items, k)
}

func (table *CacheTable) RemoveItem(k interface{}) {
	table.Lock()
	table.deleteInternal(k)
	table.Unlock()
}

func (table *CacheTable) checkExpire() {
	// STW
	table.Lock()
	if table.cleaner != nil {
		table.cleaner.Stop()
	}
	now := time.Now()
	smallestSpan := 0 * time.Second
	for key, item := range table.items {
		item.RLock()
		accessedOn := item.accessedOn
		lifeSpan := item.lifeSpan
		item.RUnlock()

		if now.Sub(accessedOn) >= lifeSpan { // delete
			// delete one item
			table.deleteInternal(key)
		} else { // record most dangerour item
			dangerousTime := lifeSpan - now.Sub(accessedOn)
			if smallestSpan == 0 || smallestSpan > dangerousTime {
				smallestSpan = dangerousTime
			}
		}
	}
	if smallestSpan > 0 {
		if table.cleanUpInterval == 0 || table.cleanUpInterval > smallestSpan {
			table.cleanUpInterval = smallestSpan
			table.cleaner = time.AfterFunc(smallestSpan, table.checkExpire)
		}
	}
	table.Unlock()
}

func (table *CacheTable) AddItem(k, v interface{}, span time.Duration) {
	item := NewCacheItem(k, v, span)
	table.Lock()
	table.items[k] = item
	addFuncs := table.doAfterAddItem
	intervalNow := table.cleanUpInterval
	table.Unlock()

	// do add callbacks
	for _, addFunc := range addFuncs {
		addFunc(k, v)
	}

	// check interval
	if span > 0 && (intervalNow == 0 || span < intervalNow) {
		// do lost check
		table.checkExpire()
	}
}

func (table *CacheTable) Data(k interface{}) (*CacheItem, error) {
	table.RLock()
	r, ok := table.items[k]
	loadFunc := table.loadDataFunc
	table.RUnlock()
	if ok {
		r.KeepAlive()
		return r, nil
	}
	// load data
	if loadFunc != nil {
		item := loadFunc(k)
		table.AddItem(item.key, item.value, item.lifeSpan)
		return item, nil
	}
	return nil, ErrorNotFoundKeyAndCanNotLoad
}
