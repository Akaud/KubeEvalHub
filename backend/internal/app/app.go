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
	refreshTokenRepo := repository.NewRefreshTokenRepository(pool)
	userService := service.NewUserService(userRepo, refreshTokenRepo, cfg.JWTSecret)
	userHandler := handler.New(userService)

	jwtAuthMiddleware := handler.JWTAuthMiddleware(userService)

	agentRepo := repository.NewAgentRepository(pool)
	agentService := service.NewAgentService(agentRepo)
	agentHandler := handler.NewAgentHandler(agentService)

	clusterRepo := repository.NewClusterRepository(pool)
	clusterService := service.NewClusterService(clusterRepo)
	clusterHandler := handler.NewClusterHandler(clusterService)

	metricRepo := repository.NewMetricRepository(pool)
	metricService := service.NewMetricService(clusterRepo, metricRepo)
	metricHandler := handler.NewMetricHandler(metricService)

	inventoryRepo := repository.NewInventoryRepository(pool)
	inventoryService := service.NewInventoryService(clusterRepo, inventoryRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	analysisService := service.NewAnalysisService(metricRepo, inventoryRepo)
	analysisHandler := handler.NewAnalysisHandler(analysisService)

	overProvisionHandler := handler.NewOverProvisionHandler(analysisService)
	underProvisionHandler := handler.NewUnderProvisionHandler(analysisService)
	recommendationHandler := handler.NewRecommendationHandler(analysisService)
	capacityHandler := handler.NewCapacityHandler(analysisService)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Post("/auth/login", userHandler.Login)
	r.Post("/auth/logout", userHandler.Logout)
	r.Post("/auth/refresh", userHandler.RefreshToken)

	r.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.CreateUser)
		r.Group(func(r chi.Router) {
			r.Use(jwtAuthMiddleware)
			r.Get("/me", userHandler.GetCurrentUser)
			r.Put("/{id}", userHandler.UpdateUser)
			r.Patch("/{id}", userHandler.PatchUser)
			r.Delete("/{id}", userHandler.DeleteUser)
		})
	})

	r.Route("/agents", func(r chi.Router) {
		r.Use(jwtAuthMiddleware)

		r.Post("/", agentHandler.CreateAgent)
		r.Get("/", agentHandler.ListAgents)
		r.Get("/{id}", agentHandler.GetAgent)
		r.Patch("/{id}/enabled", agentHandler.UpdateAgentEnabled)
		r.Delete("/{id}", agentHandler.DeleteAgent)
	})

	r.Route("/agent", func(r chi.Router) {
		r.Use(handler.AgentAuthMiddleware(agentService))
		r.Post("/heartbeat", agentHandler.Heartbeat)
		r.Post("/inventory", inventoryHandler.IngestInventory)
		r.Post("/metrics", metricHandler.IngestMetrics)
	})

	r.Route("/clusters", func(r chi.Router) {
		r.Use(jwtAuthMiddleware)
		r.Get("/", clusterHandler.ListClusters)
		r.Get("/{id}/metrics", metricHandler.GetClusterMetrics)
		r.Get("/{id}/inventory/latest", inventoryHandler.GetLatestInventory)
		r.Post("/{id}/forecast", metricHandler.ForecastMetric)
		r.Get("/{id}/analysis/utilization", analysisHandler.GetWorkloadUtilization)
		r.Get("/{id}/analysis/overprovisioned", overProvisionHandler.GetOverProvisionedWorkloads)
		r.Get("/{id}/analysis/underprovisioned", underProvisionHandler.GetUnderProvisionedWorkloads)
		r.Get("/{id}/recommendations", recommendationHandler.GetRightSizingRecommendations)
		r.Get("/{id}/capacity", capacityHandler.GetClusterCapacity)
	})

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
