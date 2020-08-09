package Cache

type Cache interface {
	Put(key string, value string)
	Get(key string) (value string, ok bool)
}
