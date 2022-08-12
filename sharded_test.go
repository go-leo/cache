package cache

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

// func TestDjb33(t *testing.T) {
// }

var shardedKeys = []string{
	"f",
	"fo",
	"foo",
	"barf",
	"barfo",
	"foobar",
	"bazbarf",
	"bazbarfo",
	"bazbarfoo",
	"foobarbazq",
	"foobarbazqu",
	"foobarbazquu",
	"foobarbazquux",
}

func TestShardedCache(t *testing.T) {
	tc := New(Shards(13)).(*ShardedCache)
	for _, v := range shardedKeys {
		tc.Set(v, "value", DefaultExpiration)
	}
}

func BenchmarkShardedCacheGetExpiring(b *testing.B) {
	benchmarkShardedCacheGet(b, 5*time.Minute)
}

func BenchmarkShardedCacheGetNotExpiring(b *testing.B) {
	benchmarkShardedCacheGet(b, NoExpiration)
}

func benchmarkShardedCacheGet(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := New(Shards(13)).(*ShardedCache)
	tc.Set("foobarba", "zquux", DefaultExpiration)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Get("foobarba")
	}
}

func BenchmarkShardedCacheGetManyConcurrentExpiring(b *testing.B) {
	benchmarkShardedCacheGetManyConcurrent(b, 5*time.Minute)
}

func BenchmarkShardedCacheGetManyConcurrentNotExpiring(b *testing.B) {
	benchmarkShardedCacheGetManyConcurrent(b, NoExpiration)
}

func benchmarkShardedCacheGetManyConcurrent(b *testing.B, exp time.Duration) {
	b.StopTimer()
	n := 10000
	tc := New(Shards(13)).(*ShardedCache)

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
