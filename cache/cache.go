package cache

// UnImplementedCache cache interface
type UnImplementedCache interface {
	Put(key string, value string) (created bool)
	Get(key string) (value string, ok bool)
}
