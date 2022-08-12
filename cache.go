package cache

import "time"

const (
	// For use with functions that take an expiration time.
	NoExpiration time.Duration = -1
	// For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to newMutexCache() or
	// newMutexCacheFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 0
)

type Cache interface {
	Set(k string, x interface{}, d time.Duration)
	Add(k string, x interface{}, d time.Duration) error
	Replace(k string, x interface{}, d time.Duration) error
	Get(k string) (interface{}, bool)
	IncrementInt(k string, n int64) (interface{}, error)
	IncrementFloat(k string, n float64) (interface{}, error)
	DecrementInt(k string, n int64) (interface{}, error)
	DecrementFloat(k string, n float64) (interface{}, error)
	Delete(k string)
	DeleteExpired()
	Items() map[string]Item
	ItemCount() int
	Flush()
}
