package main

import (
	"log"
	"net/http"
	"os"

	"tip2_pr2/services/tasks/internal/client/authclient"
	httpapi "tip2_pr2/services/tasks/internal/http"
	"tip2_pr2/services/tasks/internal/service"
	"tip2_pr2/shared/middleware"
)

func main() {
	port := getEnv("TASKS_PORT", "8082")
	authGRPCAddr := getEnv("AUTH_GRPC_ADDR", "localhost:50051")

	taskService := service.New()

	authClient, err := authclient.New(authGRPCAddr)
	if err != nil {
		log.Fatalf("create auth client failed: %v", err)
	}
	defer authClient.Close()

	handler := httpapi.New(taskService, authClient)

	mux := http.NewServeMux()
	handler.Register(mux)

	app := middleware.RequestID(middleware.Logging(mux))

	addr := ":" + port
	log.Printf("tasks service starting on %s, auth_grpc_addr=%s", addr, authGRPCAddr)

	if err := http.ListenAndServe(addr, app); err != nil {
		log.Fatalf("tasks service failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
