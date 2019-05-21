// Package main implements a client for Echo service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	pb "github.com/wangy8961/grpc-go-tutorial/features/echopb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

	resp, err := c.UnaryEcho(context.Background(), &pb.EchoRequest{Message: msg}) // Now let’s look at how we call our service methods. Note that in gRPC-Go, RPCs operate in a blocking/synchronous mode, which means that the RPC call waits for the server to respond, and will either return a response or an error.
	if err != nil {
		// log.Fatalf("failed to call UnaryEcho: %v", err)
		errStatus, _ := status.FromError(err)
		fmt.Printf("Error Code: %v\n", errStatus.Code())
		fmt.Printf("Error Description: %v\n\n", errStatus.Message())

		if codes.InvalidArgument == errStatus.Code() {
			fmt.Println("You can take specific action based on specific error!")
		}
	}
	fmt.Printf("response:\n")
	fmt.Printf(" - %q\n", resp.GetMessage())
}
