package cache

import (
	"testing"
	"time"
)

func TestNewLRUCache(t *testing.T) {
	cache := NewLRUCache(10, 10, 10)
	if cache == nil {
		t.Error("NewLRUCache failed")
	}
}

func TestLRUCache_Get(t *testing.T) {
	cache := NewLRUCache(10, 10, 10)
	cache.Set("key", "value")
	value, ok := cache.Get("key")
	if !ok || value != "value" {
		t.Error("LRUCache Get failed")
	}
}

func TestLRUCache_Delete(t *testing.T) {
	cache := NewLRUCache(10, 10, 10)
	cache.Set("key", "value")
	cache.Delete("key")
	_, ok := cache.Get("key")
	if ok {
		t.Error("LRUCache Delete failed")
	}
}

func TestLRUCache_GC(t *testing.T) {
	cache := NewLRUCache(10, time.Second, 2*time.Second)
	cache.Set("key", "value")
	time.Sleep(3 * time.Second)
	_, ok := cache.Get("key")
	if ok {
		t.Error("LRUCache GC failed")
	}
}
