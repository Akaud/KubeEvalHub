package handler

import "backend/internal/service"

// Handler groups all HTTP handlers and their dependencies.
type Handler struct {
	userService *service.UserService
}

func New(userService *service.UserService) *Handler {
	return &Handler{
		userService: userService,
	}
}
