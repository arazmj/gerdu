package main

import (
	"flag"
	"github.com/arazmj/gerdu/cache"
	"github.com/arazmj/gerdu/grpcserver"
	"github.com/arazmj/gerdu/httpserver"
	"github.com/arazmj/gerdu/lfucache"
	"github.com/arazmj/gerdu/lrucache"
	"github.com/arazmj/gerdu/memcached"
	"github.com/arazmj/gerdu/weakcache"
	"github.com/inhies/go-bytesize"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"sync"
)

var gerdu cache.UnImplementedCache
var wg = sync.WaitGroup{}

var (
	loglevel = flag.String("log", "info",
		"log level can be any of values of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', 'trace'")
	capacityStr = flag.String("capacity", "64MB",
		"The size of cache, once cache reached this capacity old values will evicted.\n"+
			"Specify a numerical value followed by one of the following units (not case sensitive)"+
			"\nK or KB: Kilobytes"+
			"\nM or MB: Megabytes"+
			"\nG or GB: Gigabytes"+
			"\nT or TB: Terabytes")
	httpPort  = flag.Int("httpport", 8080, "the http server port number")
	grpcPort  = flag.Int("grpcport", 8081, "the grpc server port number")
	mcdPort   = flag.Int("mcdport", 11211, "the memcached server port number")
	kind      = flag.String("type", "lru", "type of cache, lru or lfu, weak")
	protocols = flag.String("protocols", "http",
		"protocol 'grpc', 'http' or 'mcd' (memcached), multiple comma-separated values")
	tlsKey  = flag.String("key", "", "SSL certificate private key")
	tlsCert = flag.String("cert", "", "SSL certificate public key")
	host    = flag.String("host", "127.0.0.1", "The host that server listens")
	secure  = len(*tlsCert) > 0 && len(*tlsKey) > 0
)

func main() {
	flag.Parse()
	setLogLevel()
	setCache()
	serve()
}

func serve() {
	*protocols = strings.ToLower(*protocols)
	var validProtocol bool
	if strings.Contains(*protocols, "http") {
		validProtocol = true
		wg.Add(1)
		go func() {
			defer wg.Done()
			httpHost := *host + ":" + strconv.Itoa(*httpPort)
			if secure {
				httpserver.HTTPServeTLS(httpHost, *tlsCert, *tlsKey, gerdu)
			} else {
				httpserver.HTTPServe(httpHost, gerdu)
			}
		}()
	}
	if strings.Contains(*protocols, "grpc") {
		validProtocol = true
		wg.Add(1)
		go func() {
			defer wg.Done()
			grpcHost := *host + ":" + strconv.Itoa(*grpcPort)
			if secure {
				grpcserver.GrpcServeTLS(grpcHost, *tlsCert, *tlsKey, gerdu)
			} else {
				grpcserver.GrpcServe(grpcHost, gerdu)
			}
		}()
	}
	if strings.Contains(*protocols, "mcd") {
		validProtocol = true
		wg.Add(1)
		go func() {
			defer wg.Done()
			mcdHost := *host + ":" + strconv.Itoa(*mcdPort)
			if secure {
				log.Fatalln("Memcached protocol does not support TLS")
				os.Exit(1)
			}
			memcached.Serve(mcdHost, gerdu)
		}()
	}
	if !validProtocol {
		log.Fatalf("Invalid value for protocol")
		os.Exit(1)
	}
	wg.Wait()
}

func setCache() {
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
}

func setLogLevel() {
	switch *loglevel {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.DebugLevel)
	default:
		log.Fatalf("Invalid log level value %s\n", *loglevel)
		os.Exit(1)
	}
}
