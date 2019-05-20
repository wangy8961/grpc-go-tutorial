// Package main implements a server for Math service.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/wangy8961/grpc-go-tutorial/math/mathpb"
	"google.golang.org/grpc"
)

// server is used to implement mathpb.MathServer.
type server struct{}

// Sum implements mathpb.MathServer
func (s *server) Sum(ctx context.Context, in *pb.SumRequest) (*pb.SumResponse, error) {
	fmt.Printf("--- gRPC Unary RPC ---\n")
	fmt.Printf("request received: %v\n", in)
	return &pb.SumResponse{Result: in.FirstNum + in.SecondNum}, nil
}

// PrimeFactors implements mathpb.MathServer
func (s *server) PrimeFactors(in *pb.PrimeFactorsRequest, stream pb.Math_PrimeFactorsServer) error {
	fmt.Printf("--- gRPC Server-side Streaming RPC ---\n")
	fmt.Printf("request received: %v\n", in)

	num := in.Num
	factor := int64(2)

	for num > 1 {
		if num%factor == 0 {
			stream.Send(&pb.PrimeFactorsResponse{Result: factor})
			num = num / factor
			continue
		}
		factor++
	}

	return nil
}

// Average implements mathpb.MathServer
func (s *server) Average(stream pb.Math_AverageServer) error {
	fmt.Printf("--- gRPC Client-side Streaming RPC ---\n")

	// Read requests and send responses
	var sum int32
	count := 0

	for {
		in, err := stream.Recv()

		if err == io.EOF {
			fmt.Printf("Receiving client streaming data completed\n")
			average := float64(sum) / float64(count)
			return stream.SendAndClose(&pb.AverageResponse{Result: average})
		}

		fmt.Printf("request received: %v\n", in)

		if err != nil {
			log.Fatalf("Error while receiving client streaming data: %v", err)
		}

		sum += in.Num
		count++
	}
}

func main() {
	port := flag.Int("port", 50051, "the port to serve on")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port)) // Specify the port we want to use to listen for client requests
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("server listening at %v\n", lis.Addr())

	s := grpc.NewServer()                // Create an instance of the gRPC server
	pb.RegisterMathServer(s, &server{})  // Register our service implementation with the gRPC server
	if err := s.Serve(lis); err != nil { // Call Serve() on the server with our port details to do a blocking wait until the process is killed or Stop() is called.
		log.Fatalf("failed to serve: %v", err)
	}
}
