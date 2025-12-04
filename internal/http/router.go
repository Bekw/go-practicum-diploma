package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Bekw/go-practicum-diploma/internal/auth"
	"github.com/Bekw/go-practicum-diploma/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	store *storage.Storage
}

func NewRouter(store *storage.Storage) http.Handler {
	h := &Handler{store: store}

	r := chi.NewRouter()

	r.Post("/api/user/register", h.handleRegister)
	r.Post("/api/user/login", h.handleLogin)

	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)

		r.Post("/api/user/orders", h.handlePostOrder)
		r.Get("/api/user/orders", h.handleGetOrders)

		r.Get("/api/user/balance", h.handleGetBalance)
		r.Post("/api/user/balance/withdraw", h.handleWithdraw)
		r.Get("/api/user/withdrawals", h.handleGetWithdrawals)
	})

	return r
}

type credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if creds.Login == "" || creds.Password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	taken, err := h.store.IsLoginTaken(ctx, creds.Login)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if taken {
		http.Error(w, "login already in use", http.StatusConflict) // 409
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	userID, err := h.store.CreateUser(ctx, creds.Login, string(hash))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if creds.Login == "" || creds.Password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	user, err := h.store.GetUserByLogin(ctx, creds.Login)
	if err != nil {
		if err == storage.ErrUserNotFound {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}
