package main

import (
	"log"
	"net"
	"net/http"
	"os"

	grpcapi "tip2_pr2/services/auth/internal/grpc"
	httpapi "tip2_pr2/services/auth/internal/http"
	"tip2_pr2/services/auth/internal/service"
	"tip2_pr2/shared/middleware"

	"tip2_pr2/proto"

	"google.golang.org/grpc"
)

func main() {
	httpPort := getEnv("AUTH_PORT", "8081")
	grpcPort := getEnv("AUTH_GRPC_PORT", "50051")

	authService := service.New()

	httpHandler := httpapi.New(authService)
	httpMux := http.NewServeMux()
	httpHandler.Register(httpMux)
	httpApp := middleware.RequestID(middleware.Logging(httpMux))

	go func() {
		addr := ":" + httpPort
		log.Printf("auth http service starting on %s", addr)

		if err := http.ListenAndServe(addr, httpApp); err != nil {
			log.Fatalf("auth http service failed: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("grpc listen failed: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcHandler := grpcapi.New(authService)
	proto.RegisterAuthServiceServer(grpcServer, grpcHandler)

	log.Printf("auth grpc service starting on :%s", grpcPort)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("auth grpc service failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
