package cache

import (
	"time"
)

type options struct {
	ExpirationTime  time.Duration
	CleanupInterval time.Duration
	Shards          int
	OnEvicted       func(string, interface{})
	Hasher          Hash32
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func (o *options) init() {
	if o.Shards > 1 && o.Hasher == nil {
		o.Hasher = &DJB33Hasher{Seed: Seed()}
	}
}

type Option func(o *options)

func ExpirationTime(expiration time.Duration) Option {
	return func(o *options) {
		o.ExpirationTime = expiration
	}
}

func CleanupInterval(interval time.Duration) Option {
	return func(o *options) {
		o.CleanupInterval = interval
	}
}

func Shards(shards int) Option {
	return func(o *options) {
		o.Shards = shards
	}
}

func Hasher(hasher Hash32) Option {
	return func(o *options) {
		o.Hasher = hasher
	}
}

func OnEvicted(onEvicted func(string, interface{})) Option {
	return func(o *options) {
		o.OnEvicted = onEvicted
	}
}

func New(opts ...Option) Cache {
	o := new(options)
	o.apply(opts...)
	o.init()
	if o.Shards <= 1 {
		return newMutexCache(o.ExpirationTime, o.CleanupInterval, o.OnEvicted)
	}
	return newShardedCache(o.ExpirationTime, o.CleanupInterval, o.Shards, o.Hasher, o.OnEvicted)
}
