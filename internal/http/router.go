package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/api/user/register", stubNotImplemented)
	r.Post("/api/user/login", stubNotImplemented)
	r.Post("/api/user/orders", stubAuthRequired)
	r.Get("/api/user/orders", stubAuthRequired)
	r.Get("/api/user/balance", stubAuthRequired)
	r.Post("/api/user/balance/withdraw", stubAuthRequired)
	r.Get("/api/user/withdrawals", stubAuthRequired)

	return r
}

func stubNotImplemented(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"error":"not implemented yet"}`))
}

func stubAuthRequired(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}
