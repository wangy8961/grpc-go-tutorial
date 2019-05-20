// Package main implements a client for Math service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	pb "github.com/wangy8961/grpc-go-tutorial/math/mathpb"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func unaryCall(c pb.MathClient) {
	fmt.Printf("--- gRPC Unary RPC Call ---\n")
	// Make unary RPC
	req := &pb.SumRequest{
		FirstNum:  10,
		SecondNum: 20,
	}
	resp, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("failed to call Sum: %v", err)
	}

	fmt.Printf("response:\n")
	fmt.Printf(" - %v\n", resp.Result)
}

func main() {
	addr := flag.String("addr", "localhost:50051", "the address to connect to")
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithInsecure()) // To call service methods, we first need to create a gRPC channel to communicate with the server. We create this by passing the server address and port number to grpc.Dial()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewMathClient(conn) // Once the gRPC channel is setup, we need a client stub to perform RPCs. We get this using the NewMathClient method provided in the pb package we generated from our .proto.

	// Contact the server and print out its response.
	unaryCall(c)
}
