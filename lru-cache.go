package cache

import (
	"container/list"
	"sync"
	"time"
)

type LRUCache struct {
	mu       sync.RWMutex
	maxItems int

	expireTime time.Duration
	cleanTime  time.Duration

	cache   map[string]*list.Element
	lruList *list.List
}

type CacheItem struct {
	key      string
	value    interface{}
	expireAt time.Time
}

func (c *CacheItem) isExpired() bool {
	return c.expireAt.Before(time.Now())
}

func NewLRUCache(maxItems int, expireTime time.Duration, cleanTime time.Duration) *LRUCache {
	c := &LRUCache{
		cache:      make(map[string]*list.Element),
		lruList:    list.New(),
		maxItems:   maxItems,
		expireTime: expireTime,
		cleanTime:  cleanTime,
	}
	go c.startGC()
	return c
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	if ele, hit := c.cache[key]; hit && !ele.Value.(*CacheItem).isExpired() {
		c.mu.RUnlock()
		c.mu.Lock()
		c.lruList.MoveToFront(ele)
		c.mu.Unlock()
		return ele.Value.(*CacheItem).value, true
	}
	c.mu.RUnlock()
	return nil, false
}

func (c *LRUCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ele, hit := c.cache[key]; hit {
		c.lruList.MoveToFront(ele)

		ele.Value.(*CacheItem).expireAt = time.Now().Add(c.expireTime)
		ele.Value.(*CacheItem).value = value
		return
	}

	ele := c.lruList.PushFront(&CacheItem{key: key, value: value, expireAt: time.Now().Add(c.expireTime)})
	c.cache[key] = ele

	if c.lruList.Len() > c.maxItems {
		// Remove least recently used item
		ele := c.lruList.Back()
		if ele != nil {
			c.lruList.Remove(ele)
			delete(c.cache, ele.Value.(*CacheItem).key)
		}
	}
}

func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ele, hit := c.cache[key]; hit {
		c.lruList.Remove(ele)
		delete(c.cache, key)
	}
}

func (c *LRUCache) startGC() {
	ticker := time.NewTicker(c.cleanTime)
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for key, ele := range c.cache {
				if ele.Value.(*CacheItem).isExpired() {
					c.lruList.Remove(ele)
					delete(c.cache, key)
				}
			}
			c.mu.Unlock()
		}
	}
}
