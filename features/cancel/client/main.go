// Package main implements a client for Echo service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "github.com/wangy8961/grpc-go-tutorial/features/echopb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func sendMessage(stream pb.Echo_BidirectionalStreamingEchoClient, msg string) error {
	fmt.Printf("sending message %q\n", msg)
	return stream.Send(&pb.EchoRequest{Message: msg})
}

func recvMessage(stream pb.Echo_BidirectionalStreamingEchoClient, wantErrCode codes.Code) {
	res, err := stream.Recv()
	if status.Code(err) != wantErrCode {
		log.Fatalf("stream.Recv() = %v, %v; want=%v, status.Code(err)=%v", res, err, wantErrCode, status.Code(err))
	}
	if err != nil {
		fmt.Printf("stream.Recv() returned expected error %v\n", err)
		return
	}
	fmt.Printf("received message %q\n", res.GetMessage())
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

	// Initiate the stream with a context that supports cancellation.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	stream, err := c.BidirectionalStreamingEcho(ctx)
	if err != nil {
		log.Fatalf("error creating stream: %v", err)
	}

	// Send some test messages.
	if err := sendMessage(stream, "hello"); err != nil {
		log.Fatalf("error sending on stream: %v", err)
	}
	if err := sendMessage(stream, "world"); err != nil {
		log.Fatalf("error sending on stream: %v", err)
	}

	// Ensure the RPC is working.
	recvMessage(stream, codes.OK)
	recvMessage(stream, codes.OK)

	fmt.Println("cancelling context")
	cancel()

	// This Send may or may not return an error, depending on whether the
	// monitored context detects cancellation before the call is made.
	sendMessage(stream, "closed")

	// This Recv should never succeed.
	recvMessage(stream, codes.Canceled)
}
