// Package main implements a client for Echo service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/wangy8961/grpc-go-tutorial/features/echopb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func unaryCallWithDeadline(c pb.EchoClient, timeout time.Duration, msg string) {
	fmt.Printf("--- gRPC Unary RPC Call ---\n")
	// Make unary RPC
	// ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := &pb.EchoRequest{Message: msg}
	resp, err := c.UnaryEcho(ctx, req)
	if err != nil {
		// Error Handling
		errStatus, ok := status.FromError(err)
		if ok {
			if codes.DeadlineExceeded == errStatus.Code() { // take specific action based on specific error
				fmt.Println("Error: Deadline was exceeded")
			}
		}
		// Otherwise, ok is false and a Status is returned with codes.Unknown and the original error message
		fmt.Printf("Error Code: %v\n", errStatus.Code())
		fmt.Printf("Error Description: %v\n\n", errStatus.Message())
	}

	fmt.Printf("response:\n")
	fmt.Printf(" - %q\n", resp.GetMessage())
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

	c := pb.NewEchoClient(conn) // Once the gRPC channel is setup, we need a client stub to perform RPCs. We get this using the NewEchoClient method provided in the pb package we generated from our .proto.

	// Contact the server and print out its response.
	msg := "Madman"
	if len(os.Args) > 1 {
		msg = os.Args[1]
	}

	// 1. succeed
	unaryCallWithDeadline(c, 5*time.Second, msg)
	fmt.Println()

	// 2. failed
	unaryCallWithDeadline(c, 1*time.Second, msg)
}
