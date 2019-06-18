// Package main implements a client for User service.
package main

import (
	"time"
	"context"
	"flag"
	"log"

	"google.golang.org/grpc/credentials"

	pb "github.com/wangy8961/grpc-go-tutorial/restful-api/userpb"
	"google.golang.org/grpc"
)

func createUserCall(client pb.UserServiceClient, username, password string) {
	log.Println("--- gRPC Create RPC Call ---")

	// 设置 10 秒超时时长，可参考 https://madmalls.com/blog/post/grpc-deadline/#21-contextwithtimeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 调用 Create RPC
	req := &pb.CreateRequest{
		User: &pb.User{
			Username: username,
			Password: password,
		},
	}
	resp, err := client.Create(ctx, req)
	if err != nil {
		log.Fatalf("failed to call Create RPC: %v", err)
	}

	log.Println("response:")
	log.Printf(" - %q\n", resp)
}

func getUserCall(client pb.UserServiceClient, username string) {
	log.Println("--- gRPC Get RPC Call ---")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 调用 Get RPC
	req := &pb.GetRequest{
		Username: username,
	}
	resp, err := client.Get(ctx, req)
	if err != nil {
		log.Fatalf("failed to call Get RPC: %v", err)
	}

	log.Println("response:")
	log.Printf(" - %q\n", resp)
}

func main() {
	addr := flag.String("addr", "localhost:50051", "the address to connect to")
	certFile := flag.String("cacert", "cacert.pem", "CA root certificate")
	flag.Parse()

	creds, err := credentials.NewClientTLSFromFile(*certFile, "")
	if err != nil {
		log.Fatalf("failed to load CA root certificate: %v", err)
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(creds)) // To call service methods, we first need to create a gRPC channel to communicate with the server. We create this by passing the server address and port number to grpc.Dial()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewUserServiceClient(conn) // Once the gRPC channel is setup, we need a client stub to perform RPCs. We get this using the NewEchoClient method provided in the pb package we generated from our .proto.

	// Contact the server and print out its response.
	// 1. Create RPC Call
	createUserCall(c, "Alice", "123")
	createUserCall(c, "Bob", "pass")

	// 2. Get RPC Call
	getUserCall(c, "Alice")
}
