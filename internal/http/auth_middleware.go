package http

import (
	"context"
	"net/http"

	"github.com/Bekw/go-practicum-diploma/internal/auth"
)

type contextKey string

const userIDCtxKey contextKey = "userID"

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.CookieName())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := auth.ParseToken(cookie.Value)
		if err != nil || userID == 0 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDCtxKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserID(ctx context.Context) int64 {
	v := ctx.Value(userIDCtxKey)
	if v == nil {
		return 0
	}
	id, _ := v.(int64)
	return id
}
