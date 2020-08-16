package main

import (
	"flag"
	"github.com/arazmj/gerdu/cache"
	"github.com/arazmj/gerdu/grpcserver"
	"github.com/arazmj/gerdu/httpserver"
	"github.com/arazmj/gerdu/lfucache"
	"github.com/arazmj/gerdu/lrucache"
	"github.com/arazmj/gerdu/weakcache"
	"github.com/inhies/go-bytesize"
	"log"
	"os"
	"strings"
	"sync"
)

var gerdu cache.UnImplementedCache
var wg = sync.WaitGroup{}

var (
	verbose     = flag.Bool("verbose", false, "verbose logging")
	capacityStr = flag.String("capacity", "64MB",
		"The size of cache, once cache reached this capacity old values will evicted.\n"+
			"Specify a numerical value followed by one of the following units (not case sensitive)"+
			"\nK or KB: Kilobytes"+
			"\nM or MB: Megabytes"+
			"\nG or GB: Gigabytes"+
			"\nT or TB: Terabytes")
	httpPort  = flag.Int("httpport", 8080, "the http server port number")
	grpcPort  = flag.Int("grpcport", 8081, "the grpc server port number")
	kind      = flag.String("type", "lru", "type of cache, lru or lfu, weak")
	protocols = flag.String("protocols", "http", "protocol grpc or http, multiple values can be selected seperated by comma")
	tlsKey    = flag.String("key", "", "SSL certificate private key")
	tlsCert   = flag.String("cert", "", "SSL certificate public key")
	secure    = len(*tlsCert) > 0 && len(*tlsKey) > 0
)

func main() {
	flag.Parse()
	capacity, err := bytesize.Parse(*capacityStr)

	if err != nil {
		log.Fatal("Invalid value for capacity", err.Error())
	}

	if strings.ToLower(*kind) == "lru" {
		gerdu = lrucache.NewCache(capacity)
	} else if strings.ToLower(*kind) == "lfu" {
		gerdu = lfucache.NewCache(capacity)
	} else if strings.ToLower(*kind) == "weak" {
		gerdu = weakcache.NewWeakCache()
	} else {
		log.Fatalf("Invalid value for type")
		os.Exit(1)
	}

	*protocols = strings.ToLower(*protocols)
	if strings.Contains(*protocols, "http") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if secure {
				httpserver.HttpServeTLS(*httpPort, *tlsCert, *tlsKey, gerdu, *verbose)
			} else {
				httpserver.HttpServe(*httpPort, gerdu, *verbose)
			}
		}()
	}
	if strings.Contains(*protocols, "grpc") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if secure {
				grpcserver.GrpcServeTLS(*grpcPort, *tlsCert, *tlsKey, gerdu, *verbose)
			} else {
				grpcserver.GrpcServe(*grpcPort, gerdu, *verbose)
			}
		}()
	} else {
		log.Fatalf("Invalid value for protocol")
		os.Exit(1)
	}
	wg.Wait()
}
