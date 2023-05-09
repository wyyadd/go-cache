package cache

import (
	"sync"
	"sync/atomic"

	cmap "github.com/orcaman/concurrent-map"
)

type CacheMap interface {
	Get(k string) (interface{}, bool)
	Set(k string, x interface{})
	Delete(k string)
	Range(f func(k string, v any))
	Count() int
	Flush()
}

type RwmMap struct {
	items map[string]interface{}
	mu    sync.RWMutex
}

func NewRwmMap() CacheMap {
	return &RwmMap{items: map[string]interface{}{}}
}

func (m *RwmMap) Get(k string) (interface{}, bool) {
	m.mu.RLock()
	item, found := m.items[k]
	m.mu.RUnlock()
	if !found {
		return nil, false
	}
	return item, true
}

func (m *RwmMap) Set(k string, x interface{}) {
	m.mu.Lock()
	m.items[k] = x
	m.mu.Unlock()
}

func (m *RwmMap) Delete(k string) {
	m.mu.Lock()
	delete(m.items, k)
	m.mu.Unlock()
}

func (m *RwmMap) Range(f func(k string, v any)) {
	for k, v := range m.items {
		f(k, v)
	}
}

func (m *RwmMap) Count() int {
	return len(m.items)
}

func (m *RwmMap) Flush() {
	m.mu.Lock()
	m.items = map[string]interface{}{}
	m.mu.Unlock()
}

type SyncMap struct {
	items sync.Map
	count atomic.Int32
}

func NewSyncMap() CacheMap {
	return &SyncMap{items: sync.Map{}}
}

func (m *SyncMap) Get(k string) (interface{}, bool) {
	item, found := m.items.Load(k)
	if !found {
		return nil, false
	}
	return item, true
}

func (m *SyncMap) Set(k string, x interface{}) {
	m.items.Store(k, x)
	m.count.Add(1)
}

func (m *SyncMap) Delete(k string) {
	m.items.Delete(k)
	m.count.Add(-1)
}

func (m *SyncMap) Range(f func(k string, v any)) {
	m.items.Range(func(key, value any) bool {
		f(key.(string), value)
		return true
	})
}

func (m *SyncMap) Count() int {
	return int(m.count.Load())
}

func (m *SyncMap) Flush() {
	m.items = sync.Map{}
	m.count = atomic.Int32{}
}

type ConcurrentMap struct {
	items cmap.ConcurrentMap
}

func NewConcurrentMap() CacheMap {
	return &ConcurrentMap{items: cmap.New()}
}

func (m *ConcurrentMap) Get(k string) (interface{}, bool) {
	item, found := m.items.Get(k)
	if !found {
		return nil, false
	}
	return item, true
}

func (m *ConcurrentMap) Set(k string, x interface{}) {
	m.items.Set(k, x)
}

func (m *ConcurrentMap) Delete(k string) {
	m.items.Remove(k)
}

func (m *ConcurrentMap) Range(f func(k string, v any)) {
	for tuple := range m.items.IterBuffered() {
		f(tuple.Key, tuple.Val)
	}
}

func (m *ConcurrentMap) Count() int {
	return m.items.Count()
}

func (m *ConcurrentMap) Flush() {
	m.items.Clear()
}
