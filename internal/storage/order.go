package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Order struct {
	ID         int64
	Number     string
	UserID     int64
	Status     string
	Accrual    sql.NullFloat64
	UploadedAt time.Time
	UpdatedAt  time.Time
}

var ErrOrderNotFound = errors.New("order not found")

func (s *Storage) CreateOrder(ctx context.Context, userID int64, number string) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO orders (number, user_id, status) VALUES ($1, $2, $3)`,
		number, userID, "NEW",
	)
	return err
}

func (s *Storage) GetOrderByNumber(ctx context.Context, number string) (*Order, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, number, user_id, status, accrual, uploaded_at, updated_at
         FROM orders WHERE number = $1`,
		number,
	)

	var o Order
	if err := row.Scan(&o.ID, &o.Number, &o.UserID, &o.Status, &o.Accrual, &o.UploadedAt, &o.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	return &o, nil
}

func (s *Storage) ListOrdersByUser(ctx context.Context, userID int64) ([]Order, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, number, user_id, status, accrual, uploaded_at, updated_at
         FROM orders WHERE user_id = $1
         ORDER BY uploaded_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.Number, &o.UserID, &o.Status, &o.Accrual, &o.UploadedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, o)
	}
	return res, rows.Err()
}
func (s *Storage) ListOrdersForAccrual(ctx context.Context, limit int) ([]Order, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, number, user_id, status, accrual, uploaded_at, updated_at
         FROM orders
         WHERE status IN ('NEW', 'PROCESSING')
         ORDER BY uploaded_at
         LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.Number, &o.UserID, &o.Status, &o.Accrual, &o.UploadedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, o)
	}
	return res, rows.Err()
}
func (s *Storage) UpdateOrderAccrual(ctx context.Context, number, status string, accrual *float64) error {
	var acc sql.NullFloat64
	if accrual != nil {
		acc.Valid = true
		acc.Float64 = *accrual
	} else {
		acc.Valid = false
	}

	_, err := s.db.ExecContext(
		ctx,
		`UPDATE orders
         SET status = $2,
             accrual = $3,
             updated_at = now()
         WHERE number = $1`,
		number, status, acc,
	)
	return err
}
