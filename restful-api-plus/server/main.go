// Package main implements a server for User service.
package main

import (
	"mime"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"strings"
	"net/http"
	"context"
	"flag"
	"fmt"
	"log"

	"google.golang.org/grpc/credentials"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"

	pb "github.com/wangy8961/grpc-go-tutorial/restful-api-plus/userpb"
	swagger "github.com/wangy8961/grpc-go-tutorial/restful-api-plus/go-bindata-assetfs"
	"github.com/elazarl/go-bindata-assetfs"
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

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func serveSwagger(mux *http.ServeMux) {
	mime.AddExtensionType(".svg", "image/svg+xml")

	// Expose files in third_party/swagger-ui/ on <host>/swagger-ui
	fileServer := http.FileServer(&assetfs.AssetFS{
		Asset:    swagger.Asset,
		AssetDir: swagger.AssetDir,
		Prefix:   "third_party/swagger-ui",
	})
	prefix := "/swagger-ui/"
	mux.Handle(prefix, http.StripPrefix(prefix, fileServer))
}

func main() {
	host := flag.String("host", "localhost", "gRPC Server Name or IP")
	port := flag.Int("port", 50051, "the port to serve on")
	certFile := flag.String("certfile", "server.crt", "Server certificate")
	keyFile := flag.String("keyfile", "server.key", "Server private key")
	caCertFile := flag.String("cacert", "cacert.pem", "CA root certificate")
	swaggerJSON := flag.String("swagger", "../userpb/service.swagger.json", "Swagger JSON file")
	flag.Parse()

	// gRPC 服务和反向代理服务共同监听的地址
	endpoint := fmt.Sprintf("%s:%d", *host, *port)

	// gRPC 服务端
	creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
	if err != nil {
		log.Fatalf("failed to load certificates: %v", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds)) // Create an instance of the gRPC server
	pb.RegisterUserServiceServer(grpcServer, NewServer()) // Register our service implementation with the gRPC server
	
	// grpc-gateway 反向代理
	// 它相当于 gRPC 客户端，负责将 RESTful API 的客户端的请求转发给 gRPC 服务端
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gwCreds, gwErr := credentials.NewClientTLSFromFile(*caCertFile, "")
	if gwErr != nil {
		log.Fatalf("failed to load CA root certificate: %v", gwErr)
	}
	opts := []grpc.DialOption{grpc.WithTransportCredentials(gwCreds)}
	gwMux := runtime.NewServeMux()
	gwErr = pb.RegisterUserServiceHandlerFromEndpoint(ctx, gwMux, endpoint, opts)
	if gwErr != nil {
		log.Fatalf("failed to register grpc-gateway: %v", gwErr)
	}
	
	// 指定 gRPC-gateway 反向代理所有的 HTTP2 服务的路由
	mux := http.NewServeMux()
	mux.Handle("/", gwMux)
	// Swagger
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		// io.Copy(w, strings.NewReader(*swaggerJSON))
		http.ServeFile(w, r, *swaggerJSON)
	})
	serveSwagger(mux)
	// 启动 HTTP2 服务器（需要指定服务器的数字证书和私钥）
	srv := &http.Server{
        Addr:         endpoint,
        Handler:      grpcHandlerFunc(grpcServer, mux),  // HTTP2 服务器接收到任何请求后，再由 grpcHandlerFunc 根据请求的协议判断是直接调用 gRPC 服务端还是由 gRPC-gateway 继续反向代理
	}
	
	log.Printf("gRPC server and gRPC-gateway listening at %v\n", endpoint)
    if httpErr := srv.ListenAndServeTLS(*certFile, *keyFile); httpErr != nil {
		log.Fatalf("failed to listen and serve: %v", httpErr)
    }
}
