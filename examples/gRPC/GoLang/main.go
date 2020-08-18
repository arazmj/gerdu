package main

import (
	"GerduGoClient/proto"
	"context"
	"google.golang.org/grpc"
	"log"
	"time"
)

const (
	address = "localhost:8081"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewGerduClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.Put(ctx, &proto.PutRequest{
		Key:   "Hello",
		Value: []byte("World"),
	})

	response, err := c.Get(ctx, &proto.GetRequest{Key: "Hello"})

	if err != nil {
		log.Fatalf("did not get response: %v", err)
	}

	log.Println("Hello = " + string(response.Value))
}
