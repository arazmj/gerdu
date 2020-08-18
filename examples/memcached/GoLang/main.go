package main

import (
	"github.com/bradfitz/gomemcache/memcache"
	"log"
)

func main() {
	mc := memcache.New("localhost:11211")
	mc.Set(&memcache.Item{Key: "Hello", Value: []byte("World")})
	get, err := mc.Get("Hello")
	if err != nil {
		log.Fatalf("Key not found %v", err)
	}
	log.Println("Hello = " + string(get.Value))
}
