// Package main implements a reverse proxy gateway for User service.
package main

import (
	"google.golang.org/grpc/credentials"
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/golang/glog"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	gw "github.com/wangy8961/grpc-go-tutorial/restful-api/userpb"
)

var (
	// gRPC server 监听的地址
	endpoint = flag.String("endpoint", "localhost:50051", "endpoint of User service")
	certFile = flag.String("cacert", "cacert.pem", "CA root certificate")
)

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()

	creds, err := credentials.NewClientTLSFromFile(*certFile, "")
	if err != nil {
		log.Fatalf("failed to load CA root certificate: %v", err)
	}
	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}

	err = gw.RegisterUserServiceHandlerFromEndpoint(ctx, mux, *endpoint, opts)
	if err != nil {
		log.Fatalf("failed to automatically dials to endpoint: %v", err)
	}

	// RESTful API 反向代理所监听的地址，并设置 HTTP 服务器的路由使用 mux
	log.Println("start to listen tcp on *:5000")
	return http.ListenAndServe(":5000", mux)
}

func main() {
	flag.Parse()
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
