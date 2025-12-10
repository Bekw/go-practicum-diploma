package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Bekw/go-practicum-diploma/internal/storage"
)

type balanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type withdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (h *Handler) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	current, withdrawn, err := h.store.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := balanceResponse{
		Current:   current,
		Withdrawn: withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
func (h *Handler) handleWithdraw(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Order == "" || req.Sum <= 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if !isLuhnValid(req.Order) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err := h.store.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if errors.Is(err, storage.ErrInsufficientFunds) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func (h *Handler) handleGetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	items, err := h.store.ListWithdrawalsByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if len(items) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]withdrawalResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, withdrawalResponse{
			Order:       it.OrderNumber,
			Sum:         it.Sum,
			ProcessedAt: it.ProcessedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
