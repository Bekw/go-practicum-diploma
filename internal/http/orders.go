package http

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Bekw/go-practicum-diploma/internal/storage"
)

type orderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

func (h *Handler) handlePostOrder(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	number := strings.TrimSpace(string(body))
	if number == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if !isLuhnValid(number) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()

	existing, err := h.store.GetOrderByNumber(ctx, number)
	if err != nil && err != storage.ErrOrderNotFound {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if existing != nil {
		if existing.UserID == userID {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "order belongs to another user", http.StatusConflict)
		}
		return
	}

	if err := h.store.CreateOrder(ctx, userID, number); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) handleGetOrders(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	orders, err := h.store.ListOrdersByUser(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		var accrual *float64
		if o.Accrual.Valid {
			v := o.Accrual.Float64
			accrual = &v
		}

		resp = append(resp, orderResponse{
			Number:     o.Number,
			Status:     o.Status,
			Accrual:    accrual,
			UploadedAt: o.UploadedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
