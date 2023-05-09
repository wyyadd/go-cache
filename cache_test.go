package cache

import (
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

type TestStruct struct {
	Num      int
	Children []*TestStruct
}

func TestCache(t *testing.T) {
	testCache(t, NewRwmMap())
	testCache(t, NewSyncMap())
	testCache(t, NewConcurrentMap())
}

func testCache(t *testing.T, m CacheMap) {
	tc := New(DefaultExpiration, 0, m)

	a, found := tc.Get("a")
	if found || a != nil {
		t.Error("Getting A found value that shouldn't exist:", a)
	}

	b, found := tc.Get("b")
	if found || b != nil {
		t.Error("Getting B found value that shouldn't exist:", b)
	}

	c, found := tc.Get("c")
	if found || c != nil {
		t.Error("Getting C found value that shouldn't exist:", c)
	}

	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", "b", DefaultExpiration)
	tc.Set("c", 3.5, DefaultExpiration)

	x, found := tc.Get("a")
	if !found {
		t.Error("a was not found while getting a2")
	}
	if x == nil {
		t.Error("x for a is nil")
	} else if a2 := x.(int); a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}

	x, found = tc.Get("b")
	if !found {
		t.Error("b was not found while getting b2")
	}
	if x == nil {
		t.Error("x for b is nil")
	} else if b2 := x.(string); b2+"B" != "bB" {
		t.Error("b2 (which should be b) plus B does not equal bB; value:", b2)
	}

	x, found = tc.Get("c")
	if !found {
		t.Error("c was not found while getting c2")
	}
	if x == nil {
		t.Error("x for c is nil")
	} else if c2 := x.(float64); c2+1.2 != 4.7 {
		t.Error("c2 (which should be 3.5) plus 1.2 does not equal 4.7; value:", c2)
	}
}

func TestCacheTimes(t *testing.T) {
	testCacheTimes(t, NewRwmMap())
	testCacheTimes(t, NewSyncMap())
	testCacheTimes(t, NewConcurrentMap())
}

func testCacheTimes(t *testing.T, m CacheMap) {
	var found bool

	tc := New(50*time.Millisecond, 1*time.Millisecond, m)
	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", 2, NoExpiration)
	tc.Set("c", 3, 20*time.Millisecond)
	tc.Set("d", 4, 70*time.Millisecond)

	<-time.After(25 * time.Millisecond)
	_, found = tc.Get("c")
	if found {
		t.Error("Found c when it should have been automatically deleted")
	}

	<-time.After(30 * time.Millisecond)
	_, found = tc.Get("a")
	if found {
		t.Error("Found a when it should have been automatically deleted")
	}

	_, found = tc.Get("b")
	if !found {
		t.Error("Did not find b even though it was set to never expire")
	}

	_, found = tc.Get("d")
	if !found {
		t.Error("Did not find d even though it was set to expire later than the default")
	}

	<-time.After(20 * time.Millisecond)
	_, found = tc.Get("d")
	if found {
		t.Error("Found d when it should have been automatically deleted (later than the default)")
	}
}

func TestStorePointerToStruct(t *testing.T) {
	testStorePointerToStruct(t, NewRwmMap())
	testStorePointerToStruct(t, NewSyncMap())
	testStorePointerToStruct(t, NewConcurrentMap())
}

func testStorePointerToStruct(t *testing.T, m CacheMap) {
	tc := New(DefaultExpiration, 0, m)
	tc.Set("foo", &TestStruct{Num: 1}, DefaultExpiration)
	x, found := tc.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo")
	}
	foo := x.(*TestStruct)
	foo.Num++

	y, found := tc.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo (second time)")
	}
	bar := y.(*TestStruct)
	if bar.Num != 2 {
		t.Fatal("TestStruct.Num is not 2")
	}
}

func TestDelete(t *testing.T) {
	testDelete(t, NewRwmMap())
	testDelete(t, NewSyncMap())
	testDelete(t, NewConcurrentMap())
}

func testDelete(t *testing.T, m CacheMap) {
	tc := New(DefaultExpiration, 0, m)
	tc.Set("foo", "bar", DefaultExpiration)
	tc.Delete("foo")
	x, found := tc.Get("foo")
	if found {
		t.Error("foo was found, but it should have been deleted")
	}
	if x != nil {
		t.Error("x is not nil:", x)
	}
}

func TestItemCount(t *testing.T) {
	testItemCount(t, NewRwmMap())
	testItemCount(t, NewSyncMap())
	testItemCount(t, NewConcurrentMap())
}

func testItemCount(t *testing.T, m CacheMap) {
	tc := New(DefaultExpiration, 0, m)
	tc.Set("foo", "1", DefaultExpiration)
	tc.Set("bar", "2", DefaultExpiration)
	tc.Set("baz", "3", DefaultExpiration)
	if n := tc.ItemCount(); n != 3 {
		t.Errorf("Item count is not 3: %d", n)
	}
}

func TestFlush(t *testing.T) {
	testFlush(t, NewRwmMap())
	testFlush(t, NewSyncMap())
	testFlush(t, NewConcurrentMap())
}

func testFlush(t *testing.T, m CacheMap) {
	tc := New(DefaultExpiration, 0, m)
	tc.Set("foo", "bar", DefaultExpiration)
	tc.Set("baz", "yes", DefaultExpiration)
	tc.Flush()
	x, found := tc.Get("foo")
	if found {
		t.Error("foo was found, but it should have been deleted")
	}
	if x != nil {
		t.Error("x is not nil:", x)
	}
	x, found = tc.Get("baz")
	if found {
		t.Error("baz was found, but it should have been deleted")
	}
	if x != nil {
		t.Error("x is not nil:", x)
	}
}

func TestGetWithExpiration(t *testing.T) {
	testGetWithExpiration(t, NewRwmMap())
	testGetWithExpiration(t, NewSyncMap())
	testGetWithExpiration(t, NewConcurrentMap())
}

func testGetWithExpiration(t *testing.T, m CacheMap) {
	tc := New(DefaultExpiration, 0, m)

	a, expiration, found := tc.GetWithExpiration("a")
	if found || a != nil || !expiration.IsZero() {
		t.Error("Getting A found value that shouldn't exist:", a)
	}

	b, expiration, found := tc.GetWithExpiration("b")
	if found || b != nil || !expiration.IsZero() {
		t.Error("Getting B found value that shouldn't exist:", b)
	}

	c, expiration, found := tc.GetWithExpiration("c")
	if found || c != nil || !expiration.IsZero() {
		t.Error("Getting C found value that shouldn't exist:", c)
	}

	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", "b", DefaultExpiration)
	tc.Set("c", 3.5, DefaultExpiration)
	tc.Set("d", 1, NoExpiration)
	tc.Set("e", 1, 50*time.Millisecond)

	x, expiration, found := tc.GetWithExpiration("a")
	if !found {
		t.Error("a was not found while getting a2")
	}
	if x == nil {
		t.Error("x for a is nil")
	} else if a2 := x.(int); a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for a is not a zeroed time")
	}

	x, expiration, found = tc.GetWithExpiration("b")
	if !found {
		t.Error("b was not found while getting b2")
	}
	if x == nil {
		t.Error("x for b is nil")
	} else if b2 := x.(string); b2+"B" != "bB" {
		t.Error("b2 (which should be b) plus B does not equal bB; value:", b2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for b is not a zeroed time")
	}

	x, expiration, found = tc.GetWithExpiration("c")
	if !found {
		t.Error("c was not found while getting c2")
	}
	if x == nil {
		t.Error("x for c is nil")
	} else if c2 := x.(float64); c2+1.2 != 4.7 {
		t.Error("c2 (which should be 3.5) plus 1.2 does not equal 4.7; value:", c2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for c is not a zeroed time")
	}

	x, expiration, found = tc.GetWithExpiration("d")
	if !found {
		t.Error("d was not found while getting d2")
	}
	if x == nil {
		t.Error("x for d is nil")
	} else if d2 := x.(int); d2+2 != 3 {
		t.Error("d (which should be 1) plus 2 does not equal 3; value:", d2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for d is not a zeroed time")
	}

	x, expiration, found = tc.GetWithExpiration("e")
	if !found {
		t.Error("e was not found while getting e2")
	}
	if x == nil {
		t.Error("x for e is nil")
	} else if e2 := x.(int); e2+2 != 3 {
		t.Error("e (which should be 1) plus 2 does not equal 3; value:", e2)
	}
	// if expiration.UnixNano() != tc.Get("e").Expiration {
	// 	t.Error("expiration for e is not the correct time")
	// }
	if expiration.UnixNano() < time.Now().UnixNano() {
		t.Error("expiration for e is in the past")
	}
}

func BenchmarkGetExpiring_RwmMap(b *testing.B) {
	benchmarkGet(b, 5*time.Minute, NewRwmMap())
}

func BenchmarkGetExpiring_SyncMap(b *testing.B) {
	benchmarkGet(b, 5*time.Minute, NewSyncMap())
}

func BenchmarkGetExpiring_ConcurrentMap(b *testing.B) {
	benchmarkGet(b, 5*time.Minute, NewConcurrentMap())
}

func BenchmarkGetNotExpiring_RwmMap(b *testing.B) {
	benchmarkGet(b, NoExpiration, NewRwmMap())
}

func BenchmarkGetNotExpiring_SyncMap(b *testing.B) {
	benchmarkGet(b, NoExpiration, NewSyncMap())
}

func BenchmarkGetNotExpiring_ConcurrentMap(b *testing.B) {
	benchmarkGet(b, NoExpiration, NewConcurrentMap())
}

func benchmarkGet(b *testing.B, exp time.Duration, m CacheMap) {
	b.StopTimer()
	tc := New(exp, 0, m)
	tc.Set("foo", "bar", DefaultExpiration)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Get("foo")
	}
}

func BenchmarkGet_RWMutex(b *testing.B) {
	b.StopTimer()
	m := map[string]string{
		"foo": "bar",
	}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_, _ = m["foo"]
		mu.RUnlock()
	}
}

func BenchmarkGetConcurrentExpiring_RwmMap(b *testing.B) {
	benchmarkGetConcurrent(b, 5*time.Minute, NewRwmMap())
}

func BenchmarkGetConcurrentExpiring_SyncMap(b *testing.B) {
	benchmarkGetConcurrent(b, 5*time.Minute, NewSyncMap())
}

func BenchmarkGetConcurrentExpiring_ConcurrentMap(b *testing.B) {
	benchmarkGetConcurrent(b, 5*time.Minute, NewConcurrentMap())
}

func BenchmarkGetConcurrentNotExpiring_RwmMap(b *testing.B) {
	benchmarkGetConcurrent(b, NoExpiration, NewRwmMap())
}

func BenchmarkGetConcurrentNotExpiring_SyncMap(b *testing.B) {
	benchmarkGetConcurrent(b, NoExpiration, NewSyncMap())
}

func BenchmarkGetConcurrentNotExpiring_ConcurrentMap(b *testing.B) {
	benchmarkGetConcurrent(b, NoExpiration, NewConcurrentMap())
}

func benchmarkGetConcurrent(b *testing.B, exp time.Duration, m CacheMap) {
	b.StopTimer()
	tc := New(exp, 0, m)
	tc.Set("foo", "bar", DefaultExpiration)
	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < each; j++ {
				tc.Get("foo")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkGetConcurrent_RWMutex(b *testing.B) {
	b.StopTimer()
	m := map[string]string{
		"foo": "bar",
	}
	mu := sync.RWMutex{}
	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < each; j++ {
				mu.RLock()
				_, _ = m["foo"]
				mu.RUnlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkGetManyConcurrentExpiring_RwmMap(b *testing.B) {
	benchmarkGetManyConcurrent(b, 5*time.Minute, NewRwmMap())
}

func BenchmarkGetManyConcurrentExpiring_SyncMap(b *testing.B) {
	benchmarkGetManyConcurrent(b, 5*time.Minute, NewSyncMap())
}

func BenchmarkGetManyConcurrentExpiring_ConcurrentMap(b *testing.B) {
	benchmarkGetManyConcurrent(b, 5*time.Minute, NewConcurrentMap())
}

func BenchmarkGetManyConcurrentNotExpiring_RwmMap(b *testing.B) {
	benchmarkGetManyConcurrent(b, NoExpiration, NewRwmMap())
}

func BenchmarkGetManyConcurrentNotExpiring_SyncMap(b *testing.B) {
	benchmarkGetManyConcurrent(b, NoExpiration, NewSyncMap())
}

func BenchmarkGetManyConcurrentNotExpiring_ConcurrentMap(b *testing.B) {
	benchmarkGetManyConcurrent(b, NoExpiration, NewConcurrentMap())
}

func benchmarkGetManyConcurrent(b *testing.B, exp time.Duration, m CacheMap) {
	b.StopTimer()
	n := 10000
	tc := New(exp, 0, m)
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		k := "foo" + strconv.Itoa(i)
		keys[i] = k
		tc.Set(k, "bar", DefaultExpiration)
	}
	each := b.N / n
	wg := new(sync.WaitGroup)
	wg.Add(n)
	for _, v := range keys {
		go func(k string) {
			for j := 0; j < each; j++ {
				tc.Get(k)
			}
			wg.Done()
		}(v)
	}
	b.StartTimer()
	wg.Wait()
}

func BenchmarkSetExpiring_RwmMap(b *testing.B) {
	benchmarkSet(b, 5*time.Minute, NewRwmMap())
}

func BenchmarkSetExpiring_SyncMap(b *testing.B) {
	benchmarkSet(b, 5*time.Minute, NewSyncMap())
}

func BenchmarkSetExpiring_ConcurrentMap(b *testing.B) {
	benchmarkSet(b, 5*time.Minute, NewConcurrentMap())
}

func BenchmarkSetNotExpiring_RwmMap(b *testing.B) {
	benchmarkSet(b, NoExpiration, NewRwmMap())
}

func BenchmarkSetNotExpiring_SyncMap(b *testing.B) {
	benchmarkSet(b, NoExpiration, NewSyncMap())
}

func BenchmarkSetNotExpiring_ConcurrentMap(b *testing.B) {
	benchmarkSet(b, NoExpiration, NewConcurrentMap())
}

func benchmarkSet(b *testing.B, exp time.Duration, m CacheMap) {
	b.StopTimer()
	tc := New(exp, 0, m)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Set("foo", "bar", DefaultExpiration)
	}
}

func BenchmarkSet_RWMutex(b *testing.B) {
	b.StopTimer()
	m := map[string]string{}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		m["foo"] = "bar"
		mu.Unlock()
	}
}

func BenchmarkSetDelete_RwmMap(b *testing.B) {
	benchmarkSetDelete(b, NewRwmMap())
}

func BenchmarkSetDelete_SyncMap(b *testing.B) {
	benchmarkSetDelete(b, NewSyncMap())
}

func BenchmarkSetDelete_ConcurrentMap(b *testing.B) {
	benchmarkSetDelete(b, NewConcurrentMap())
}

func benchmarkSetDelete(b *testing.B, m CacheMap) {
	b.StopTimer()
	tc := New(DefaultExpiration, 0, m)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Set("foo", "bar", DefaultExpiration)
		tc.Delete("foo")
	}
}

func BenchmarkSetDelete_RWMutex(b *testing.B) {
	b.StopTimer()
	m := map[string]string{}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		m["foo"] = "bar"
		mu.Unlock()
		mu.Lock()
		delete(m, "foo")
		mu.Unlock()
	}
}

func BenchmarkDeleteExpiredLoop_RwmMap(b *testing.B) {
	benchmarkDeleteExpiredLoop(b, NewRwmMap())
}

func BenchmarkDeleteExpiredLoop_SyncMap(b *testing.B) {
	benchmarkDeleteExpiredLoop(b, NewSyncMap())
}

func BenchmarkDeleteExpiredLoop_ConcurrentMap(b *testing.B) {
	benchmarkDeleteExpiredLoop(b, NewConcurrentMap())
}

func benchmarkDeleteExpiredLoop(b *testing.B, m CacheMap) {
	b.StopTimer()
	tc := New(5*time.Minute, 0, m)
	for i := 0; i < 100000; i++ {
		tc.Set(strconv.Itoa(i), "bar", DefaultExpiration)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.DeleteExpired()
	}
}
