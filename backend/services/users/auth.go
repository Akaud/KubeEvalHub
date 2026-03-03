package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Akaud/KubeEvalHub/helpers"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type registerReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Login    string `json:"login"` // email OR username
	Password string `json:"password"`
}

type authResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"` // seconds
	User        resp   `json:"user"`
}

func (h *Handler) RegisterAuthRoutes(r chi.Router) {
	r.Post("/api/auth/register", h.register)
	r.Post("/api/auth/login", h.login)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Username == "" || req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username, email, password required"})
		return
	}
	if !strings.Contains(req.Email, "@") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email"})
		return
	}
	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password too short"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		helpers.Log.Error("bcrypt failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}

	u, err := h.repo.Create(r.Context(), req.Username, req.Email, string(hash))
	if err != nil {
		helpers.Log.Error("register failed", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot create user"})
		return
	}

	token, expiresIn, err := mintJWT(u.ID, u.Email)
	if err != nil {
		helpers.Log.Error("jwt failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}

	writeJSON(w, http.StatusCreated, authResp{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User: resp{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		},
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "login and password required"})
		return
	}

	u, err := h.repo.GetByLogin(r.Context(), req.Login)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		helpers.Log.Error("login get user failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token, expiresIn, err := mintJWT(u.ID, u.Email)
	if err != nil {
		helpers.Log.Error("jwt failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}

	writeJSON(w, http.StatusOK, authResp{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User: resp{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		},
	})
}

func mintJWT(userID int64, email string) (token string, expiresInSec int64, err error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", 0, errors.New("JWT_SECRET not set")
	}

	ttlMin := 60
	if s := os.Getenv("JWT_TTL_MIN"); s != "" {
		if v, e := strconv.Atoi(s); e == nil && v > 0 {
			ttlMin = v
		}
	}

	now := time.Now()
	exp := now.Add(time.Duration(ttlMin) * time.Minute)
	expiresInSec = int64(time.Until(exp).Seconds())

	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iat":   now.Unix(),
		"exp":   exp.Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	return signed, expiresInSec, nil
}
