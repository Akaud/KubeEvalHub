package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/app"
	"backend/internal/config"
	"backend/internal/db"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}
	defer pool.Close()

	srv := app.NewServer(cfg, pool)

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("server started on %s", srv.Addr)
		serverErr <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	case sig := <-stop:
		log.Printf("shutdown signal received: %s", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
			_ = srv.Close()
		}
	}
}
