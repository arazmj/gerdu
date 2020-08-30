package grpcserver

import (
	"context"
	"errors"
	"github.com/arazmj/gerdu/cache"
	"github.com/arazmj/gerdu/proto"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
)

type server struct {
	proto.UnimplementedGerduServer
	gerdu cache.UnImplementedCache
}

func grpcServe(s *grpc.Server, host string, gerdu cache.UnImplementedCache) {
	lis, err := net.Listen("tcp", host)
	log.Printf("Gerdu started listening gRPC at %s\n", host)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	proto.RegisterGerduServer(s, &server{
		gerdu: gerdu,
	})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

//GrpcServe start gRPC server in non-secure mode
func GrpcServe(host string, gerdu cache.UnImplementedCache) {
	s := grpc.NewServer()
	grpcServe(s, host, gerdu)
}

//GrpcServeTLS start gRPC server secure
func GrpcServeTLS(host string, tlsCert, tlsKey string, gerdu cache.UnImplementedCache) {
	credentials, err := credentials.NewServerTLSFromFile(tlsCert, tlsKey)
	if err != nil {
		log.Fatalf("Failed to setup TLS for gRPC service: %v", err)
	}

	s := grpc.NewServer(grpc.Creds(credentials))
	grpcServe(s, host, gerdu)
}

func (s *server) Put(ctx context.Context, request *proto.PutRequest) (*proto.PutResponse, error) {
	value := string(request.Value)
	key := request.Key
	created := s.gerdu.Put(key, value)
	if !created {
		log.Printf("gRPC UPDATE Key: %s Value: %s\n", key, value)
	} else {
		log.Printf("gRPC INSERT Key: %s Value: %s\n", key, value)
	}
	return &proto.PutResponse{
		Created: created,
	}, nil
}

func (s *server) Get(ctx context.Context, request *proto.GetRequest) (*proto.GetResponse, error) {
	value, ok := s.gerdu.Get(request.Key)
	if ok {
		log.Printf("gRPC RETREIVED Key: %s Value: %s\n", request.Key, value)
		return &proto.GetResponse{
			Value: []byte(value),
		}, nil
	}
	log.Printf("gRPC MISSED Key: %s \n", request.Key)
	return nil, errors.New("key not found")
}

func (s *server) Delete(ctx context.Context, request *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	ok := s.gerdu.Delete(request.Key)
	if ok {
		log.Printf("gRPC DELETE Key: %s\n", request.Key)
		return &proto.DeleteResponse{
			Deleted: true,
		}, nil
	}
	log.Printf("gRPC MISSED Key: %s \n", request.Key)
	return nil, errors.New("key not found")
}
