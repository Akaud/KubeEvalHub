package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Akaud/KubeEvalHub/config"
	"github.com/Akaud/KubeEvalHub/db"
	"github.com/Akaud/KubeEvalHub/helpers"
	"github.com/Akaud/KubeEvalHub/middleware"
	usersvc "github.com/Akaud/KubeEvalHub/services/users"
)

func Run() int {
	helpers.InitLogger()

	dbCfg := config.DB()

	pg, err := db.Open(dbCfg)
	if err != nil {
		helpers.Log.Error("db open failed", "error", err)
		return 1
	}
	defer pg.Close()

	if err := db.Ping(context.Background(), pg, dbCfg.PingTimeoutSec); err != nil {
		helpers.Log.Error("db ping failed", "error", err)
		return 1
	}
	helpers.Log.Info("db connected", "host", dbCfg.Host, "db", dbCfg.Name)

	if err := db.Migrate(pg); err != nil {
		helpers.Log.Error("migrations failed", "error", err)
		return 1
	}
	helpers.Log.Info("migrations applied")

	r := chi.NewRouter()
	r.Use(middleware.ChiLogger)

	r.Get("/api/health", HealthHandler(pg))

	uh := usersvc.NewHandler(pg)
	uh.RegisterRoutes(r)
	uh.RegisterAuthRoutes(r)

	srv := &http.Server{
		Addr:         ":" + config.String("PORT", "5000"),
		Handler:      r,
		ReadTimeout:  config.DurationSeconds("READ_TIMEOUT", 5),
		WriteTimeout: config.DurationSeconds("WRITE_TIMEOUT", 10),
		IdleTimeout:  config.DurationSeconds("IDLE_TIMEOUT", 120),
	}

	go func() {
		helpers.Log.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			helpers.Log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

	helpers.Log.Info("server exited properly")
	return 0
}
