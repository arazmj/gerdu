package grpcserver

import (
	"context"
	"errors"
	"github.com/arazmj/gerdu/cache"
	"github.com/arazmj/gerdu/proto"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"strconv"
)

type server struct {
	proto.UnimplementedGerduServer
	verbose bool
	gerdu   cache.UnImplementedCache
}

func grpcServe(s *grpc.Server, port int, gerdu cache.UnImplementedCache, verbose bool) {
	host := ":" + strconv.Itoa(port)
	lis, err := net.Listen("tcp", host)
	log.Printf("Gerdu started listening gRPC on %d port\n", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	proto.RegisterGerduServer(s, &server{
		gerdu:   gerdu,
		verbose: verbose,
	})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func GrpcServe(port int, gerdu cache.UnImplementedCache, verbose bool) {
	s := grpc.NewServer()
	grpcServe(s, port, gerdu, verbose)
}

func GrpcServeTLS(port int, tlsCert, tlsKey string, gerdu cache.UnImplementedCache, verbose bool) {
	credentials, err := credentials.NewServerTLSFromFile(tlsCert, tlsKey)
	if err != nil {
		log.Fatalf("Failed to setup TLS for gRPC service: %v", err)
	}

	s := grpc.NewServer(grpc.Creds(credentials))
	grpcServe(s, port, gerdu, verbose)
}

func (s *server) Put(ctx context.Context, request *proto.PutRequest) (*proto.PutResponse, error) {
	value := string(request.Value)
	key := request.Key
	created := s.gerdu.Put(key, value)
	if s.verbose {
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
	value, ok := s.gerdu.Get(request.Key)
	if ok {
		if s.verbose {
			log.Printf("gRPC RETREIVED Key: %s Value: %s\n", request.Key, value)
		}
		return &proto.GetResponse{
			Value: []byte(value),
		}, nil
	}
	if s.verbose {
		log.Printf("gRPC MISSED Key: %s \n", value)
	}
	return nil, errors.New("key not found")
}
