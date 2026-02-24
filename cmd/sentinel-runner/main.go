package main

import (
	"log"
	"net"
	"os"

	"github.com/g0ulartleo/mirante/internal/sentinel"
	"github.com/g0ulartleo/mirante/internal/sentinel/builtins"
	"github.com/g0ulartleo/mirante/internal/sentinel/runtime/server"
	runtimev1 "github.com/g0ulartleo/mirante/proto/sentinelruntime/v1"
	"google.golang.org/grpc"
)

func main() {
	addr := os.Getenv("SENTINEL_RUNNER_ADDR")
	if addr == "" {
		addr = "0.0.0.0:50051"
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}

	factory := sentinel.NewFactory()
	builtins.Register(factory)

	grpcServer := grpc.NewServer()
	runtimev1.RegisterSentinelRuntimeServer(grpcServer, server.New(factory))

	log.Printf("sentinel-runner listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}
