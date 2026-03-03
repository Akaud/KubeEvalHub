package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/go-chi/chi/v5"

	"github.com/Akaud/KubeEvalHub/db"
	"github.com/Akaud/KubeEvalHub/helpers"
	"github.com/Akaud/KubeEvalHub/middleware"
	usersvc "github.com/Akaud/KubeEvalHub/services/users"
)

func main() {
	_ = godotenv.Load()
	helpers.InitLogger()

	dbCfg := db.LoadConfigFromEnv()
	pg, err := db.Open(dbCfg)
	if err != nil {
		helpers.Log.Error("db open failed", "error", err)
		os.Exit(1)
	}
	defer pg.Close()

	if err := db.Ping(context.Background(), pg); err != nil {
		helpers.Log.Error("db ping failed", "error", err)
		os.Exit(1)
	}
	helpers.Log.Info("db connected", "host", dbCfg.Host, "db", dbCfg.Name)

	if err := db.Migrate(pg); err != nil {
		helpers.Log.Error("migrations failed", "error", err)
		os.Exit(1)
	}
	helpers.Log.Info("migrations applied")

	port := getEnv("PORT", "5000")

	r := chi.NewRouter()
	r.Use(middleware.ChiLogger)

	r.Get("/api/health", healthHandler(pg))
	uh := usersvc.NewHandler(pg)
	uh.RegisterRoutes(r)
	uh.RegisterAuthRoutes(r)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  time.Duration(getEnvAsInt("READ_TIMEOUT", 5)) * time.Second,
		WriteTimeout: time.Duration(getEnvAsInt("WRITE_TIMEOUT", 10)) * time.Second,
		IdleTimeout:  time.Duration(getEnvAsInt("IDLE_TIMEOUT", 120)) * time.Second,
	}

	go func() {
		helpers.Log.Info("server starting", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			helpers.Log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = server.Shutdown(ctx)
	helpers.Log.Info("server exited properly")
}

func healthHandler(pg *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pg.PingContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("db down"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if s, ok := os.LookupEnv(key); ok {
		var n int
		for i := 0; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return fallback
			}
			n = n*10 + int(s[i]-'0')
		}
		return n
	}
	return fallback
}
