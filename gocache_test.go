package main

import (
	"context"
	"github.com/arazmj/gerdu/lrucache"
	"github.com/arazmj/gerdu/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"testing"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	proto.RegisterGerduServer(s, &server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestServerGrpc(t *testing.T) {
	gerdu = lrucache.NewCache(1000)
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := proto.NewGerduClient(conn)
	resp, err := client.Put(ctx, &proto.PutRequest{
		Key:   "Key1",
		Value: []byte("Value1"),
	})
	if err != nil {
		t.Fatalf("gRPC Put failed: %v", err)
	}

	if resp.Created != true {
		t.Fatalf("gRPC Put could not create the key")
	}

	response, err := client.Get(ctx, &proto.GetRequest{
		Key: "Key1",
	})

	if err != nil {
		t.Fatalf("gRPC Get failed: %v", err)
	}

	value := string(response.Value)
	if value != "Value1" {
		t.Fatalf("gRPC Get the value does not match expecting Value1, but got %v", value)
	}
}
