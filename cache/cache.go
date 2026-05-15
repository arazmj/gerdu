// Package cache general interface for cache
package cache

import "time"

// UnImplementedCache cache interface
type UnImplementedCache interface {
	Put(key string, value string) (created bool)
	PutWithTTL(key string, value string, ttl time.Duration) (created bool)
	Get(key string) (value string, ok bool)
	Delete(key string) (ok bool)
}
