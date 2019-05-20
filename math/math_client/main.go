// Package main implements a client for Math service.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/wangy8961/grpc-go-tutorial/math/mathpb"
	"google.golang.org/grpc"
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

func serverSideStreamingCall(c pb.MathClient) {
	fmt.Printf("--- gRPC Server-side Streaming RPC Call ---\n")
	// Make server-side streaming RPC
	req := &pb.PrimeFactorsRequest{Num: 48}
	stream, err := c.PrimeFactors(context.Background(), req)
	if err != nil {
		log.Fatalf("failed to call PrimeFactors: %v", err)
	}

	// Read all the responses
	var rpcStatus error
	fmt.Printf("response:\n")
	for {
		resp, err := stream.Recv()
		if err != nil {
			rpcStatus = err
			break
		}
		fmt.Printf(" - %v\n", resp.Result)
	}
	if rpcStatus != io.EOF {
		log.Fatalf("failed to finish server-side streaming: %v", rpcStatus)
	}
}

func clientSideStreamingCall(c pb.MathClient) {
	fmt.Printf("--- gRPC Client-side Streaming RPC Call ---\n")
	// Make client-side streaming RPC
	stream, err := c.Average(context.Background())
	if err != nil {
		log.Fatalf("failed to call Average: %v", err)
	}

	// Send all requests to the server
	nums := []int32{10, 8, 2, 36, 24, 16, 98}
	for _, num := range nums {
		if err := stream.Send(&pb.AverageRequest{Num: num}); err != nil {
			log.Fatalf("failed to send streaming: %v", err)
		}
	}

	// Read the response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to CloseAndRecv: %v", err)
	}
	fmt.Printf("response:\n")
	fmt.Printf(" - %v\n", resp.Result)
}

func bidirectionalStreamingCall(c pb.MathClient) {
	fmt.Printf("--- gRPC Bidirectional Streaming RPC Call ---\n")
	// Make bidirectional streaming RPC
	stream, err := c.Maximum(context.Background())
	if err != nil {
		log.Fatalf("failed to call Maximum: %v", err)
	}

	// goroutine: Send all requests to the server
	go func() {
		nums := []int32{2, 10, 8, 24, 16, 32, 98}
		for _, num := range nums {
			if err := stream.Send(&pb.MaximumRequest{Num: num}); err != nil {
				log.Fatalf("failed to send streaming: %v", err)
			}
			// Sleep 1 second
			time.Sleep(1 * time.Second)
		}
		// closes the send direction of the stream
		stream.CloseSend()
	}()

	// Read all the responses
	var rpcStatus error
	fmt.Printf("response:\n")
	for {
		resp, err := stream.Recv()
		if err != nil {
			rpcStatus = err
			break
		}
		fmt.Printf(" - %v\n", resp.Result)
	}
	if rpcStatus != io.EOF {
		log.Fatalf("failed to finish server streaming: %v", rpcStatus)
	}
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
	// 1. Unary RPC Call
	// unaryCall(c)

	// 2. Server-side Streaming RPC Call
	// serverSideStreamingCall(c)

	// 3. Client-side Streaming RPC Call
	// clientSideStreamingCall(c)

	// 4. Bidirectional Streaming RPC Call
	bidirectionalStreamingCall(c)
}
