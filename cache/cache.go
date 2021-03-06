// Package cache general interface for cache
package cache

// UnImplementedCache cache interface
type UnImplementedCache interface {
	Put(key string, value string) (created bool)
	Get(key string) (value string, ok bool)
	Delete(key string) (ok bool)
}
