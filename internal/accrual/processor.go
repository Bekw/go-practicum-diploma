package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Bekw/go-practicum-diploma/internal/storage"
)

type Processor struct {
	baseURL string
	store   *storage.Storage
	client  *http.Client
}

func NewProcessor(baseURL string, store *storage.Storage) *Processor {
	return &Processor{
		baseURL: strings.TrimRight(baseURL, "/"),
		store:   store,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (p *Processor) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var nextAllowed time.Time

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !nextAllowed.IsZero() && time.Now().Before(nextAllowed) {
				continue
			}
			_ = p.processBatch(ctx, &nextAllowed)
		}
	}
}

func (p *Processor) processBatch(ctx context.Context, nextAllowed *time.Time) error {
	const batchSize = 10

	orders, err := p.store.ListOrdersForAccrual(ctx, batchSize)
	if err != nil {
		return err
	}
	if len(orders) == 0 {
		return nil
	}

	for _, o := range orders {
		if err := p.processOrder(ctx, &o, nextAllowed); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}
	}

	return nil
}

type accrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

func (p *Processor) processOrder(ctx context.Context, o *storage.Order, nextAllowed *time.Time) error {
	url := fmt.Sprintf("%s/api/orders/%s", p.baseURL, o.Number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var ar accrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
			return err
		}

		var newStatus string
		switch ar.Status {
		case "REGISTERED", "PROCESSING":
			newStatus = "PROCESSING"
		case "INVALID":
			newStatus = "INVALID"
		case "PROCESSED":
			newStatus = "PROCESSED"
		default:
			newStatus = o.Status
		}

		return p.store.UpdateOrderAccrual(ctx, o.Number, newStatus, ar.Accrual)

	case http.StatusNoContent:
		return nil

	case http.StatusTooManyRequests:
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if sec, err := strconv.Atoi(ra); err == nil {
				*nextAllowed = time.Now().Add(time.Duration(sec) * time.Second)
			}
		}
		return nil

	case http.StatusInternalServerError:
		return nil

	default:
		return nil
	}
}
