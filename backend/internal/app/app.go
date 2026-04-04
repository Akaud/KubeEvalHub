package app

import (
	"net/http"

	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewServer(cfg *config.Config, pool *pgxpool.Pool) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	userRepo := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepo)
	h := handler.New(userService)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.CreateUser)
		r.Put("/{id}", h.UpdateUser)
		r.Patch("/{id}", h.PatchUser)
		r.Delete("/{id}", h.DeleteUser)
	})

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
