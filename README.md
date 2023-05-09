# go-cache

go-cache is an in-memory key:value store/cache similar to memcached that is
suitable for applications running on a single machine. Its major advantage is
that, being essentially a thread-safe `map[string]interface{}` with expiration
times, it doesn't need to serialize or transmit its contents over the network.

Any object can be stored, for a given duration or forever, and the cache can be
safely used by multiple goroutines.

### Installation

`go get github.com/wyyadd/go-cache`

### Usage

```go
import (
	"fmt"
	"github.com/wyyadd/go-cache"
	"time"
)

func main() {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	// This project offer three different maps to store key:value; (See BenchMark)
	c := cache.New(5*time.Minute, 10*time.Minute, NewRwmMap())
	c := cache.New(5*time.Minute, 10*time.Minute, NewSyncMap())
	c := cache.New(5*time.Minute, 10*time.Minute, NewConcurrentMap())

	// Set the value of the key "foo" to "bar", with the default expiration time
	c.Set("foo", "bar", cache.DefaultExpiration)

	// Set the value of the key "baz" to 42, with no expiration time
	// (the item won't be removed until it is re-set, or removed using
	// c.Delete("baz")
	c.Set("baz", 42, cache.NoExpiration)

	// Get the string associated with the key "foo" from the cache
	foo, found := c.Get("foo")
	if found {
		fmt.Println(foo)
	}

	// Since Go is statically typed, and cache values can be anything, type
	// assertion is needed when values are being passed to functions that don't
	// take arbitrary types, (i.e. interface{}). The simplest way to do this for
	// values which will only be used once--e.g. for passing to another
	// function--is:
	foo, found := c.Get("foo")
	if found {
		MyFunction(foo.(string))
	}

	// This gets tedious if the value is used several times in the same function.
	// You might do either of the following instead:
	if x, found := c.Get("foo"); found {
		foo := x.(string)
		// ...
	}
	// or
	var foo string
	if x, found := c.Get("foo"); found {
		foo = x.(string)
	}
	// ...
	// foo can then be passed around freely as a string

	// Want performance? Store pointers!
	c.Set("foo", &MyStruct, cache.DefaultExpiration)
	if x, found := c.Get("foo"); found {
		foo := x.(*MyStruct)
			// ...
	}
}
```
### BenchMark
```
goos: windows
goarch: amd64
pkg: github.com/wyyadd/go-cache
cpu: AMD Ryzen 5 5600 6-Core Processor
BenchmarkGetExpiring_RwmMap-12                                  97032516                11.90 ns/op            0 B/op         0 allocs/op
BenchmarkGetExpiring_SyncMap-12                                 47984644                23.69 ns/op            0 B/op          0 allocs/op
BenchmarkGetExpiring_ConcurrentMap-12                           82739790                14.66 ns/op            0 B/op          0 allocs/op
BenchmarkGetNotExpiring_RwmMap-12                               155887176               11.24 ns/op            0 B/op          0 allocs/op
BenchmarkGetNotExpiring_SyncMap-12                              59958028                19.88 ns/op            0 B/op          0 allocs/op
BenchmarkGetNotExpiring_ConcurrentMap-12                        100000000               10.28 ns/op            0 B/op          0 allocs/op
BenchmarkGet_RWMutex-12                                         292061973                4.112 ns/op           0 B/op          0 allocs/op
BenchmarkGetConcurrentExpiring_RwmMap-12                        45670094                26.23 ns/op            0 B/op          0 allocs/op
BenchmarkGetConcurrentExpiring_SyncMap-12                       298293536                4.008 ns/op           0 B/op          0 allocs/op
BenchmarkGetConcurrentExpiring_ConcurrentMap-12                 46130602                25.54 ns/op            0 B/op          0 allocs/op
BenchmarkGetConcurrentNotExpiring_RwmMap-12                     35811273                35.54 ns/op            0 B/op          0 allocs/op
BenchmarkGetConcurrentNotExpiring_SyncMap-12                    374984414                3.167 ns/op           0 B/op          0 allocs/op
BenchmarkGetConcurrentNotExpiring_ConcurrentMap-12              47858529                24.87 ns/op            0 B/op          0 allocs/op
BenchmarkGetConcurrent_RWMutex-12                               47170923                24.97 ns/op            0 B/op          0 allocs/op
BenchmarkGetManyConcurrentExpiring_RwmMap-12                    100000000               21.88 ns/op            0 B/op          0 allocs/op
BenchmarkGetManyConcurrentExpiring_SyncMap-12                   290337565                4.174 ns/op           0 B/op          0 allocs/op
BenchmarkGetManyConcurrentExpiring_ConcurrentMap-12             246717678                4.958 ns/op           0 B/op          0 allocs/op
BenchmarkGetManyConcurrentNotExpiring_RwmMap-12                 36202042                36.76 ns/op            0 B/op          0 allocs/op
BenchmarkGetManyConcurrentNotExpiring_SyncMap-12                356510050                3.398 ns/op           0 B/op          0 allocs/op
BenchmarkGetManyConcurrentNotExpiring_ConcurrentMap-12          270097808                4.345 ns/op           0 B/op          0 allocs/op
BenchmarkSetExpiring_RwmMap-12                                  23313872                51.17 ns/op           24 B/op          1 allocs/op
BenchmarkSetExpiring_SyncMap-12                                 10110098               115.9 ns/op            56 B/op          3 allocs/op
BenchmarkSetExpiring_ConcurrentMap-12                           21408118                55.15 ns/op           24 B/op          1 allocs/op
BenchmarkSetNotExpiring_RwmMap-12                               29246250                45.86 ns/op           24 B/op          1 allocs/op
BenchmarkSetNotExpiring_SyncMap-12                              11419250               108.4 ns/op            56 B/op          3 allocs/op
BenchmarkSetNotExpiring_ConcurrentMap-12                        26648071                46.37 ns/op           24 B/op          1 allocs/op
BenchmarkSet_RWMutex-12                                         88821778                14.10 ns/op            0 B/op          0 allocs/op
BenchmarkSetDelete_RwmMap-12                                    18248229                82.52 ns/op           24 B/op          1 allocs/op
BenchmarkSetDelete_SyncMap-12                                    2575394               442.7 ns/op           368 B/op          9 allocs/op
BenchmarkSetDelete_ConcurrentMap-12                             16538492                69.65 ns/op           24 B/op          1 allocs/op
BenchmarkSetDelete_RWMutex-12                                   42067917                28.70 ns/op            0 B/op          0 allocs/op
BenchmarkDeleteExpiredLoop_RwmMap-12                                1201           1130260 ns/op              24 B/op          1 allocs/op
BenchmarkDeleteExpiredLoop_SyncMap-12                                576           1933035 ns/op              40 B/op          2 allocs/op
BenchmarkDeleteExpiredLoop_ConcurrentMap-12                           91          14652297 ns/op         6617957 B/op        201 allocs/op
PASS
ok      github.com/wyyadd/go-cache      50.179s
```