package cache

import (
	"runtime"
	"time"
)

// This is an experimental and unexported (for now) attempt at making a cache
// with better algorithmic complexity than the standard one, namely by
// preventing write locks of the entire cache when an item is added. As of the
// time of writing, the overhead of selecting buckets results in cache
// operations being about twice as slow as for the standard cache with small
// total cache sizes, and faster for larger ones.
//
// See cache_test.go for a few benchmarks.

type ShardedCache struct {
	*shardedCache
}

type Hash32 interface {
	Sum32(k string) uint32
}

type shardedCache struct {
	m       uint32
	cs      []*mutexCache
	janitor *shardedJanitor
	hasher  Hash32
}

func (sc *shardedCache) bucket(k string) *mutexCache {
	return sc.cs[sc.hasher.Sum32(k)%sc.m]
}

func (sc *shardedCache) Set(k string, x interface{}, d time.Duration) {
	sc.bucket(k).Set(k, x, d)
}

func (sc *shardedCache) Add(k string, x interface{}, d time.Duration) error {
	return sc.bucket(k).Add(k, x, d)
}

func (sc *shardedCache) Replace(k string, x interface{}, d time.Duration) error {
	return sc.bucket(k).Replace(k, x, d)
}

func (sc *shardedCache) Get(k string) (interface{}, bool) {
	return sc.bucket(k).Get(k)
}

func (sc *shardedCache) IncrementInt(k string, n int64) (interface{}, error) {
	return sc.bucket(k).IncrementInt(k, n)
}

func (sc *shardedCache) IncrementFloat(k string, n float64) (interface{}, error) {
	return sc.bucket(k).IncrementFloat(k, n)
}

func (sc *shardedCache) DecrementInt(k string, n int64) (interface{}, error) {
	return sc.bucket(k).DecrementInt(k, n)
}

func (sc *shardedCache) DecrementFloat(k string, n float64) (interface{}, error) {
	return sc.bucket(k).DecrementFloat(k, n)
}

func (sc *shardedCache) Delete(k string) {
	sc.bucket(k).Delete(k)
}

func (sc *shardedCache) DeleteExpired() {
	for _, v := range sc.cs {
		v.DeleteExpired()
	}
}

func (sc *shardedCache) ItemCount() int {
	var length int
	for _, v := range sc.cs {
		length += v.ItemCount()
	}
	return length
}

func (sc *shardedCache) Items() map[string]Item {
	res := make(map[string]Item, sc.ItemCount())
	for _, v := range sc.cs {
		items := v.Items()
		for k, item := range items {
			res[k] = item
		}
	}
	return res
}

func (sc *shardedCache) Flush() {
	for _, v := range sc.cs {
		v.Flush()
	}
}

type shardedJanitor struct {
	cleanupInterval time.Duration
	stop            chan bool
}

func (j *shardedJanitor) Run(sc *shardedCache) {
	j.stop = make(chan bool)
	tick := time.Tick(j.cleanupInterval)
	for {
		select {
		case <-tick:
			sc.DeleteExpired()
		case <-j.stop:
			return
		}
	}
}

func stopShardedJanitor(sc *ShardedCache) {
	sc.janitor.stop <- true
}

func runShardedJanitor(sc *shardedCache, cleanupInterval time.Duration) {
	j := &shardedJanitor{
		cleanupInterval: cleanupInterval,
	}
	sc.janitor = j
	go j.Run(sc)
}

func newNestedShardedCache(n int, expirationTime time.Duration, hasher Hash32, onEvicted ...func(string, interface{})) *shardedCache {
	sc := &shardedCache{
		m:      uint32(n),
		cs:     make([]*mutexCache, n),
		hasher: hasher,
	}
	var f func(string, interface{})
	if len(onEvicted) > 0 {
		f = onEvicted[0]
	}
	for i := 0; i < n; i++ {
		c := &mutexCache{
			expirationTime: expirationTime,
			items:          map[string]Item{},
			onEvicted:      f,
		}
		sc.cs[i] = c
	}
	return sc
}

func newShardedCache(expirationTime, cleanupInterval time.Duration,
	shards int, hasher Hash32, onEvicted ...func(string, interface{})) *ShardedCache {
	if expirationTime == 0 {
		expirationTime = -1
	}
	sc := newNestedShardedCache(shards, expirationTime, hasher, onEvicted...)
	SC := &ShardedCache{sc}
	if cleanupInterval > 0 {
		runShardedJanitor(sc, cleanupInterval)
		runtime.SetFinalizer(SC, stopShardedJanitor)
	}
	return SC
}
