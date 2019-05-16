// Package main implements a server for Greeter service.
package main

import (
	"context"
	"log"
	"net"

	pb "github.com/wangy8961/learn-gRPC/greet/greetpb"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement greetpb.GreeterServer.
type server struct{}

// SayHello implements greetpb.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {
	lis, err := net.Listen("tcp", port) // Specify the port we want to use to listen for client requests
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()                  // Create an instance of the gRPC server
	pb.RegisterGreeterServer(s, &server{}) // Register our service implementation with the gRPC server
	if err := s.Serve(lis); err != nil {   // Call Serve() on the server with our port details to do a blocking wait until the process is killed or Stop() is called.
		log.Fatalf("failed to serve: %v", err)
	}
}
