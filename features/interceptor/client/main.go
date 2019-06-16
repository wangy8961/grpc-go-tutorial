// Package main implements a client for Echo service.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/grpc/credentials/oauth"

	"google.golang.org/grpc/credentials"

	pb "github.com/wangy8961/grpc-go-tutorial/features/echopb"
	"google.golang.org/grpc"
)

func unaryCall(client pb.EchoClient) {
	fmt.Printf("--- gRPC Unary RPC Call ---\n")

	// 设置 10 秒超时时长，可参考 https://madmalls.com/blog/post/grpc-deadline/#21-contextwithtimeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 调用 Unary RPC
	req := &pb.EchoRequest{Message: "madmalls.com"}
	resp, err := client.UnaryEcho(ctx, req)
	if err != nil {
		log.Fatalf("failed to call UnaryEcho: %v", err)
	}

	fmt.Printf("response:\n")
	fmt.Printf(" - %q\n", resp.GetMessage())
}

func bidirectionalStreamingCall(c pb.EchoClient) {
	fmt.Printf("--- gRPC Bidirectional Streaming RPC Call ---\n")

	// 设置 10 秒超时时长，可参考 https://madmalls.com/blog/post/grpc-deadline/#21-contextwithtimeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Make bidirectional streaming RPC
	stream, err := c.BidirectionalStreamingEcho(ctx)
	if err != nil {
		log.Fatalf("failed to call BidirectionalStreamingEcho: %v", err)
	}

	// Send all requests to the server
	for i := 0; i < 5; i++ {
		if err := stream.Send(&pb.EchoRequest{Message: fmt.Sprintf("Request %d", i+1)}); err != nil {
			log.Fatalf("failed to send request due to error: %v", err)
		}
	}

	// closes the send direction of the stream
	stream.CloseSend()

	// Read all the responses
	var rpcStatus error
	fmt.Printf("response:\n")
	for {
		resp, err := stream.Recv()
		if err != nil {
			rpcStatus = err
			break
		}
		fmt.Printf(" - %q\n", resp.Message)
	}
	if rpcStatus != io.EOF {
		log.Fatalf("failed to finish server streaming: %v", rpcStatus)
	}
}

// client-side unary interceptor (For Authentication)
func unaryAuthInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	opts = append(opts, grpc.PerRPCCredentials(oauth.NewOauthAccess(&oauth2.Token{
		AccessToken: "some-oauth2-secret-token",
	})))
	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}

// client-side unary interceptor (For Logging)
func unaryLogInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// Logic before invoking the invoker
	start := time.Now()
	// Calls the invoker to execute RPC
	err := invoker(ctx, method, req, reply, cc, opts...)
	// Logic after invoking the invoker
	log.Printf("Invoked RPC method: %s, Duration time: %s, Error: %v", method, time.Since(start), err)

	return err
}

// client-side streaming interceptor (For Authentication)
func streamAuthInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	opts = append(opts, grpc.PerRPCCredentials(oauth.NewOauthAccess(&oauth2.Token{
		AccessToken: "some-oauth2-secret-token",
	})))
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func main() {
	addr := flag.String("addr", "localhost:50051", "the address to connect to")
	certFile := flag.String("cacert", "cacert.pem", "CA root certificate")
	flag.Parse()

	creds, err := credentials.NewClientTLSFromFile(*certFile, "")
	if err != nil {
		log.Fatalf("failed to load CA root certificate: %v", err)
	}

	opts := []grpc.DialOption{
		// 1. TLS Credential
		grpc.WithTransportCredentials(creds),
		// 2. Client Unary Interceptors
		grpc.WithChainUnaryInterceptor(
			unaryAuthInterceptor,
			unaryLogInterceptor,
		),
		// 3. Client Streaming Interceptor
		grpc.WithStreamInterceptor(streamAuthInterceptor),
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, opts...) // To call service methods, we first need to create a gRPC channel to communicate with the server. We create this by passing the server address and port number to grpc.Dial()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewEchoClient(conn) // Once the gRPC channel is setup, we need a client stub to perform RPCs. We get this using the NewEchoClient method provided in the pb package we generated from our .proto.

	// Contact the server and print out its response.
	// 1. Unary RPC Call
	unaryCall(c)

	// 2. Bidirectional Streaming RPC Call
	// bidirectionalStreamingCall(c)
}
