package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Withdrawal struct {
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}

var ErrInsufficientFunds = errors.New("insufficient funds")

func (s *Storage) GetBalance(ctx context.Context, userID int64) (current, withdrawn float64, err error) {
	var accrual sql.NullFloat64

	if err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(accrual), 0)
         FROM orders
         WHERE user_id = $1 AND status = 'PROCESSED'`,
		userID,
	).Scan(&accrual); err != nil {
		return
	}
	if accrual.Valid {
		current = accrual.Float64
	}

	if err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(sum), 0)
         FROM withdrawals
         WHERE user_id = $1`,
		userID,
	).Scan(&withdrawn); err != nil {
		return
	}

	current = current - withdrawn
	return
}

func (s *Storage) Withdraw(ctx context.Context, userID int64, order string, sum float64) (err error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var accrualTotal float64
	if err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(accrual), 0)
         FROM orders
         WHERE user_id = $1 AND status = 'PROCESSED'`,
		userID,
	).Scan(&accrualTotal); err != nil {
		return err
	}

	var withdrawnTotal float64
	if err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(sum), 0)
         FROM withdrawals
         WHERE user_id = $1
         FOR UPDATE`,
		userID,
	).Scan(&withdrawnTotal); err != nil {
		return err
	}

	current := accrualTotal - withdrawnTotal
	if current+1e-9 < sum {
		err = ErrInsufficientFunds
		return err
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO withdrawals (user_id, order_number, sum)
         VALUES ($1, $2, $3)`,
		userID, order, sum,
	); err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func (s *Storage) CreateWithdrawal(ctx context.Context, userID int64, order string, sum float64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO withdrawals (user_id, order_number, sum)
         VALUES ($1, $2, $3)`,
		userID, order, sum,
	)
	return err
}

func (s *Storage) ListWithdrawalsByUser(ctx context.Context, userID int64) ([]Withdrawal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT order_number, sum, processed_at
         FROM withdrawals
         WHERE user_id = $1
         ORDER BY processed_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Withdrawal
	for rows.Next() {
		var w Withdrawal
		if err := rows.Scan(&w.OrderNumber, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		res = append(res, w)
	}
	return res, rows.Err()
}
