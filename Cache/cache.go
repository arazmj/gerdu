package cache

// ICache cache interface
type ICache interface {
	Put(key string, value string) (created bool)
	Get(key string) (value string, ok bool)
}
