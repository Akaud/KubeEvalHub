package handler

import (
	"errors"
	"fmt"
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

type verifyEmailRequest struct {
	Token string `json:"token"`
}

type resendVerificationRequest struct {
	Email string `json:"email"`
}

type createUserResponse struct {
	Message string       `json:"message"`
	User    userResponse `json:"user,omitempty"`
}

type userResponse struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func toUserResponse(user *model.User) userResponse {
	return userResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
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

	user, verifyToken, err := h.userService.CreateUserWithVerification(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	publicURL := getPublicURL(r)
	verifyLink := fmt.Sprintf("%s/verify-email?token=%s", publicURL, verifyToken)

	if h.resendService != nil {
		go func() {
			if err := h.resendService.SendVerificationEmail(user.Email, user.Name, verifyLink); err != nil {
				println("Failed to send verification email:", err.Error())
			}
		}()
	}

	writeJSON(w, http.StatusCreated, createUserResponse{
		Message: "Registration successful. Please check your email to verify your account.",
		User:    toUserResponse(user),
	})
}

// In your handler package, add this helper function
func getPublicURL(r *http.Request) string {
	scheme := "http"

	// Check for HTTPS (important for production)
	if r.TLS != nil {
		scheme = "https"
	}

	// Check for forwarded protocol (Cloudflare, reverse proxies)
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}

	// Get host from request
	host := r.Host

	// Check for forwarded host (important for Cloudflare tunnel)
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

// VerifyEmail verifies a user's email address
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.userService.VerifyEmail(r.Context(), req.Token); err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Email verified successfully. You can now log in.",
	})
}

func (h *Handler) ResendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	var req resendVerificationRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	verifyToken, err := h.userService.ResendVerificationEmail(r.Context(), req.Email)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	user, err := h.userService.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	publicURL := getPublicURL(r)
	verifyLink := fmt.Sprintf("%s/verify-email?token=%s", publicURL, verifyToken)

	if h.resendService != nil {
		go func() {
			if err := h.resendService.SendVerificationEmail(user.Email, user.Name, verifyLink); err != nil {
				println("Failed to send verification email:", err.Error())
			}
		}()
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Verification email sent. Please check your inbox.",
	})
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
	case errors.Is(err, service.ErrEmailNotVerified):
		writeError(w, http.StatusForbidden, "email not verified. Please verify your email before logging in.")
	case errors.Is(err, service.ErrUserNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrEmailAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrInvalidVerificationToken):
		writeError(w, http.StatusBadRequest, "invalid or expired verification token")
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
