# GoCache

GoCache is a thread-safe key-value in-memory database server written in GoLang.
Currently it supports two eviction policy LFU (Least Frequently Used) and LRU (Least Recently Used). 
It also supports for weak reference type of cache where the cache consumes as much memory as the garbage collector allows it to use.
```
 -capacity int
        how big the cache will be, the old values will be evicted (default 100)
 -port int
        the server port number. (default 8080)
 -type string
        type of cache, lru or lfu (default "lru")
```

Example of usage:
To insert or update a key 
```
curl --request POST http://localhost:8080/cache/1/1
curl --request POST http://localhost:8080/cache/2/2
curl --request POST http://localhost:8080/cache/3/3
curl --request POST http://localhost:8080/cache/3/alskdjfhaslkdjfhasklfhdlkasjdhflaksfhdakljdshflkasjhfalskdjfhasldkhfasdklfdhlksajdfhas
```

To retrieve the key
```
curl --request GET localhost:8080/cache/3
```

Prometheus metrics
```
curl --request GET localhost:8080/metrics
```