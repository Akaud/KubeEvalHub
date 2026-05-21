package handler

import "backend/internal/service"

type Handler struct {
	userService   *service.UserService
	resendService *service.ResendService
}

func New(userService *service.UserService, resendService *service.ResendService) *Handler {
	return &Handler{
		userService:   userService,
		resendService: resendService,
	}
}
