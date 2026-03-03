package users

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Akaud/KubeEvalHub/helpers"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	repo *Repo
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: NewRepo(db)}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/users", func(r chi.Router) {
		r.Post("/", h.create)
		r.Route("/{id}", func(r chi.Router) {
			r.Put("/", h.update)
			r.Patch("/", h.update)
			r.Delete("/", h.delete)
		})
	})
}

type createReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateReq struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type resp struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createReq
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
		helpers.Log.Error("create user failed", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot create user"})
		return
	}

	writeJSON(w, http.StatusCreated, resp{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(chi.URLParam(r, "id"))
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var req updateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if req.Username != nil {
		s := strings.TrimSpace(*req.Username)
		req.Username = &s
		if s == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username cannot be empty"})
			return
		}
	}
	if req.Email != nil {
		s := strings.TrimSpace(strings.ToLower(*req.Email))
		req.Email = &s
		if s == "" || !strings.Contains(s, "@") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email"})
			return
		}
	}

	var passwordHash *string
	if req.Password != nil {
		if len(*req.Password) < 8 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password too short"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			helpers.Log.Error("bcrypt failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
			return
		}
		s := string(hash)
		passwordHash = &s
	}

	u, err := h.repo.Update(r.Context(), id, req.Username, req.Email, passwordHash)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		helpers.Log.Error("update user failed", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot update user"})
		return
	}

	writeJSON(w, http.StatusOK, resp{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	})
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(chi.URLParam(r, "id"))
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		helpers.Log.Error("delete user failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseID(s string) (int64, bool) {
	if s == "" {
		return 0, false
	}
	var n int64
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int64(s[i]-'0')
	}
	return n, n > 0
}
