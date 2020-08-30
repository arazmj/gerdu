package main

import (
	"flag"
	cache "github.com/arazmj/gerdu/cache"
	"github.com/arazmj/gerdu/grpcserver"
	"github.com/arazmj/gerdu/httpserver"
	"github.com/arazmj/gerdu/lfucache"
	"github.com/arazmj/gerdu/lrucache"
	"github.com/arazmj/gerdu/memcached"
	"github.com/arazmj/gerdu/raftproxy"
	"github.com/arazmj/gerdu/weakcache"
	"github.com/inhies/go-bytesize"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
)

var gerdu raftproxy.RaftCache
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
	protocols = flag.String("protocols", "",
		"protocol 'grpc' or 'mcd' (memcached), multiple comma-separated values, http is not optional")
	tlsKey   = flag.String("key", "", "SSL certificate private key")
	tlsCert  = flag.String("cert", "", "SSL certificate public key")
	host     = flag.String("host", "127.0.0.1", "The host that server listens")
	raftAddr = flag.String("raft", "127.0.0.1:12000", "Set Raft bind address")
	joinAddr = flag.String("join", "", "Set join address, if any")
	nodeID   = flag.String("id", "master", "Node ID")
	storage  = flag.String("storage", "", "Path to store log files and snapshot, will store in memory if not set")

	secure = len(*tlsCert) > 0 && len(*tlsKey) > 0
)

func main() {
	flag.Parse()
	setLogLevel()
	setCache()
	serve()
}

func serve() {
	*protocols = strings.ToLower(*protocols)

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

	if strings.Contains(*protocols, "grpc") {
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

	wg.Wait()

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("Gerdu exiting")
}

func setCache() {
	capacity, err := bytesize.Parse(*capacityStr)
	if err != nil {
		log.Fatal("Invalid value for capacity", err.Error())
	}

	var c cache.UnImplementedCache
	if strings.ToLower(*kind) == "lru" {
		c = lrucache.NewCache(capacity)
	} else if strings.ToLower(*kind) == "lfu" {
		c = lfucache.NewCache(capacity)
	} else if strings.ToLower(*kind) == "weak" {
		c = weakcache.NewWeakCache()
	} else {
		log.Fatalf("Invalid value for type")
		os.Exit(1)
	}
	gerdu = raftproxy.NewRaftProxy(c, *raftAddr, *joinAddr, *nodeID)
	err = gerdu.OpenRaft(*storage)
	if err != nil {
		log.Fatalf("Cannot open raft peer connection: %s", err)
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
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
	default:
		log.Fatalf("Invalid log level value %s\n", *loglevel)
		os.Exit(1)
	}
}
