// Package main implements a server for User service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc/credentials"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"

	pb "github.com/wangy8961/grpc-go-tutorial/restful-api/userpb"
	"google.golang.org/grpc"
)

// server is used to implement pb.UserServiceServer.
type server struct {
	users map[string]pb.User
}

// NewServer creates User service
func NewServer() pb.UserServiceServer {
	return &server{
		users: make(map[string]pb.User),
	}
}

// Create a new user
func (s *server) Create(ctx context.Context, req *pb.CreateRequest) (*empty.Empty, error) {
	log.Println("--- Creating new user... ---")
	log.Printf("request received: %v\n", req)

	user := req.GetUser()
	if user.Username == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "username cannot be empty")
	}
	if user.Password == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "password cannot be empty")
	}

	s.users[user.Username] = *user

	log.Println("--- User created! ---")
	return &empty.Empty{}, nil
}

// Get a specified user
func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	log.Println("--- Getting user... ---")

	if req.Username == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "username cannot be empty")
	}

	u, exists := s.users[req.Username]
	if !exists {
		return nil, grpc.Errorf(codes.NotFound, "user not found")
	}

	log.Println("--- User found! ---")
	return &pb.GetResponse{User: &u}, nil
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

	pb.RegisterUserServiceServer(s, NewServer()) // Register our service implementation with the gRPC server
	if err := s.Serve(lis); err != nil {         // Call Serve() on the server with our port details to do a blocking wait until the process is killed or Stop() is called.
		log.Fatalf("failed to serve: %v", err)
	}
}
