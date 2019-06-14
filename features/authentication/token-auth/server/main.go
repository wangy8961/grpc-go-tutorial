// Package main implements a server for Echo service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	pb "github.com/wangy8961/grpc-go-tutorial/features/echopb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// server is used to implement echopb.EchoServer.
type server struct{}

/*
func (s *server) UnaryEcho(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnaryEcho not implemented")
}
*/
func (s *server) UnaryEcho(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	fmt.Printf("--- gRPC Unary RPC ---\n")
	fmt.Printf("request received: %v\n", req)

	// md 的值类似于: map[:authority:[192.168.40.123:50051] authorization:[Bearer some-secret-token] content-type:[application/grpc] user-agent:[grpc-go/1.20.1]]
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}
	// fmt.Printf("Type of 'metadata.MD' is %T, and its value is %v \n", md, md)

	// 1. 判断是否存在 authorization 请求头
	authorization, ok := md["authorization"]
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, `missing "Authorization" header`)
	}
	// fmt.Printf("Type of 'authorization' is %T, and its value is %v \n", authorization, authorization)

	const prefix = "Bearer "

	// 2. 如果存在 authorization 请求头的话，则 md["authorization"] 是一个 []string
	if !strings.HasPrefix(authorization[0], prefix) {
		return nil, status.Errorf(codes.Unauthenticated, `missing "Bearer " prefix in "Authorization" header`)
	}

	// 3. 验证 token 是否一致
	if token := strings.TrimPrefix(authorization[0], prefix); token != "some-secret-token" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	return &pb.EchoResponse{Message: req.GetMessage()}, nil
}

func (s *server) ServerStreamingEcho(req *pb.EchoRequest, stream pb.Echo_ServerStreamingEchoServer) error {
	return status.Errorf(codes.Unimplemented, "method ServerStreamingEcho not implemented")
}

func (s *server) ClientStreamingEcho(stream pb.Echo_ClientStreamingEchoServer) error {
	return status.Errorf(codes.Unimplemented, "method ClientStreamingEcho not implemented")
}

func (s *server) BidirectionalStreamingEcho(stream pb.Echo_BidirectionalStreamingEchoServer) error {
	return status.Errorf(codes.Unimplemented, "method BidirectionalStreamingEcho not implemented")
}

func main() {
	port := flag.Int("port", 50051, "the port to serve on")
	certFile := flag.String("certfile", "server.crt", "Server certificate")
	keyFile := flag.String("keyfile", "server.key", "Server private key")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port)) // Specify the port we want to use to listen for client requests
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("server listening at %v\n", lis.Addr())

	creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
	if err != nil {
		log.Fatalf("failed to load certificates: %v", err)
	}

	s := grpc.NewServer(grpc.Creds(creds)) // Create an instance of the gRPC server

	pb.RegisterEchoServer(s, &server{})  // Register our service implementation with the gRPC server
	if err := s.Serve(lis); err != nil { // Call Serve() on the server with our port details to do a blocking wait until the process is killed or Stop() is called.
		log.Fatalf("failed to serve: %v", err)
	}
}
