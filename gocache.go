package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/arazmj/gerdu/cache"
	"github.com/arazmj/gerdu/httpserver"
	"github.com/arazmj/gerdu/lfucache"
	"github.com/arazmj/gerdu/lrucache"
	"github.com/arazmj/gerdu/proto"
	"github.com/arazmj/gerdu/weakcache"
	"github.com/inhies/go-bytesize"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"os"
	"strconv"
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
	capacity, _ := bytesize.Parse(*capacityStr)

	if strings.ToLower(*kind) == "lru" {
		gerdu = lrucache.NewCache(capacity)
	} else if strings.ToLower(*kind) == "lfu" {
		gerdu = lfucache.NewCache(capacity)
	} else if strings.ToLower(*kind) == "weak" {
		gerdu = weakcache.NewWeakCache()
	} else {
		fmt.Println("Invalid value for type")
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
			grpcServer()
		}()
	} else {
		fmt.Println("Invalid value for protocol")
		os.Exit(1)
	}
	wg.Wait()
}

func grpcServer() {
	defer wg.Done()
	host := ":" + strconv.Itoa(*grpcPort)
	lis, err := net.Listen("tcp", host)
	log.Printf("Gerdu started listening gRPC on %d port\n", *grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var s *grpc.Server
	if len(*tlsCert) > 0 && len(*tlsKey) > 0 {
		credentials, err := credentials.NewServerTLSFromFile(*tlsCert, *tlsKey)
		if err != nil {
			log.Fatalf("Failed to setup TLS for gRPC service: %v", err)
		}

		s = grpc.NewServer(grpc.Creds(credentials))
	} else {
		s = grpc.NewServer()
	}
	proto.RegisterGerduServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type server struct {
	proto.UnimplementedGerduServer
}

func (s *server) Put(ctx context.Context, request *proto.PutRequest) (*proto.PutResponse, error) {
	value := string(request.Value)
	key := request.Key
	created := gerdu.Put(key, value)
	if *verbose {
		if !created {
			log.Printf("gRPC UPDATE Key: %s Value: %s\n", key, value)
		} else {
			log.Printf("gRPC INSERT Key: %s Value: %s\n", key, value)
		}
	}
	return &proto.PutResponse{
		Created: created,
	}, nil
}

func (s *server) Get(ctx context.Context, request *proto.GetRequest) (*proto.GetResponse, error) {
	value, ok := gerdu.Get(request.Key)
	if ok {
		if *verbose {
			log.Printf("gRPC RETREIVED Key: %s Value: %s\n", request.Key, value)
		}
		return &proto.GetResponse{
			Value: []byte(value),
		}, nil
	}
	if *verbose {
		log.Printf("gRPC MISSED Key: %s \n", value)
	}
	return nil, errors.New("key not found")
}
