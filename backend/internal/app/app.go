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
	userService := service.NewUserService(userRepo, cfg.JWTSecret)
	userHandler := handler.New(userService)

	agentRepo := repository.NewAgentRepository(pool)
	agentService := service.NewAgentService(agentRepo)
	agentHandler := handler.NewAgentHandler(agentService)

	jwtAuthMiddleware := handler.JWTAuthMiddleware(userService)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// public routes
	r.Post("/auth/login", userHandler.Login)

	r.Route("/users", func(r chi.Router) {
		// registration must stay public
		r.Post("/", userHandler.CreateUser)

		// authenticated user-only routes
		r.Group(func(r chi.Router) {
			r.Use(jwtAuthMiddleware)

			r.Put("/{id}", userHandler.UpdateUser)
			r.Patch("/{id}", userHandler.PatchUser)
			r.Delete("/{id}", userHandler.DeleteUser)
		})
	})

	// authenticated owner routes
	r.Route("/agents", func(r chi.Router) {
		r.Use(jwtAuthMiddleware)

		r.Post("/", agentHandler.CreateAgent)
		r.Get("/", agentHandler.ListAgents)
		r.Get("/{id}", agentHandler.GetAgent)
		r.Patch("/{id}/enabled", agentHandler.UpdateAgentEnabled)
	})

	// agent-authenticated routes
	r.Route("/agent", func(r chi.Router) {
		// replace with your agent token auth middleware
		// r.Use(handler.AgentAuthMiddleware(agentService))

		r.Post("/heartbeat", agentHandler.Heartbeat)
	})

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
