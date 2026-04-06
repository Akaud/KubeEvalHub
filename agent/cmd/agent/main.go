package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"agent/internal/config"
	"agent/internal/kube"
	"agent/internal/runner"
	"agent/internal/transport"
)

func main() {
	cfg := config.Load()

	if cfg.AgentToken == "" {
		log.Fatal("AGENT_TOKEN is required")
	}
	if cfg.BackendURL == "" {
		log.Fatal("BACKEND_URL is required")
	}

	kubeClients, err := kube.NewInClusterClients()
	if err != nil {
		log.Fatalf("failed to create kube clients: %v", err)
	}

	backendClient := transport.New(
		cfg.BackendURL,
		cfg.AgentToken,
		cfg.RequestTimeout,
		cfg.InsecureSkipTLS,
	)

	r := runner.New(cfg, kubeClients, backendClient)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := r.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("agent stopped with error: %v", err)
	}
}
