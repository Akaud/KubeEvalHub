package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"backend/internal/model"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type upsertUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type patchUserRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type authResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type refreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type logoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type userResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toUserResponse(user *model.User) userResponse {
	return userResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	accessToken, refreshToken, err := h.userService.AuthenticateUser(r.Context(), req.Identifier, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *Handler) DeleteCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := getAuthenticatedUserID(r)
	if !ok || userID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.userService.DeleteUser(r.Context(), userID); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	accessToken, refreshToken, err := h.userService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, refreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.userService.RevokeRefreshToken(r.Context(), req.RefreshToken); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req upsertUserRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userService.CreateUser(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	accessToken, refreshToken, err := h.userService.AuthenticateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "user created but token generation failed")
		return
	}

	w.Header().Set("Location", "/users/"+strconv.FormatInt(user.ID, 10))
	writeJSON(w, http.StatusCreated, struct {
		User         userResponse `json:"user"`
		AccessToken  string       `json:"accessToken"`
		RefreshToken string       `json:"refreshToken"`
	}{
		User:         toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func requireSameUser(r *http.Request, targetUserID int64) error {
	authUserID, ok := getAuthenticatedUserID(r)
	if !ok || authUserID <= 0 {
		return service.ErrInvalidCredentials
	}

	if authUserID != targetUserID {
		return errors.New("forbidden")
	}

	return nil
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := requireSameUser(r, id); err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req upsertUserRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), id, req.Name, req.Email, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *Handler) PatchUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := requireSameUser(r, id); err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req patchUserRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == nil && req.Email == nil && req.Password == nil {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	user, err := h.userService.PatchUser(r.Context(), id, req.Name, req.Email, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := requireSameUser(r, id); err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.userService.DeleteUser(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseIDParam(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, service.ErrInvalidUserID
	}
	return id, nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidUserID),
		errors.Is(err, service.ErrInvalidUserName),
		errors.Is(err, service.ErrInvalidUserEmail),
		errors.Is(err, service.ErrInvalidUserPassword):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials),
		errors.Is(err, service.ErrInvalidToken),
		errors.Is(err, service.ErrRefreshTokenExpired):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrEmailAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := getAuthenticatedUserID(r)
	if !ok || userID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}
