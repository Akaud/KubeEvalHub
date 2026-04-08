package main

import (
	"context"
	"errors"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	kubeClients, err := kube.NewInClusterClients()
	if err != nil {
		log.Fatalf("failed to create kube clients: %v", err)
	}

	backendClient := transport.New(
		cfg.BackendURL,
		cfg.AgentToken,
		cfg.RequestTimeout,
	)

	r := runner.New(cfg, kubeClients, backendClient)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := r.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("agent stopped with error: %v", err)
	}
}
