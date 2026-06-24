package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Yertore/sub-aggregator/internal/model"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`

	row := r.db.QueryRow(ctx, query,
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
	)

	var created model.Subscription
	err := row.Scan(
		&created.ID, &created.ServiceName, &created.Price, &created.UserID,
		&created.StartDate, &created.EndDate, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}
	slog.Info("subscription created", "id", created.ID)
	return &created, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1`

	row := r.db.QueryRow(ctx, query, id)

	var sub model.Subscription
	err := row.Scan(
		&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
		&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	slog.Info("subscription got", "id", sub.ID)
	return &sub, nil
}

func (r *Repository) List(ctx context.Context, userID, serviceName string) ([]*model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE ($1 = '' OR user_id::text = $1)
		  AND ($2 = '' OR service_name ILIKE $2)
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID, serviceName)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []*model.Subscription
	for rows.Next() {
		var sub model.Subscription
		if err := rows.Scan(
			&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
			&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, &sub)
	}
	slog.Info("subscriptions listed", "count", len(subs))
	return subs, nil
}

func (r *Repository) Update(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	query := `
		UPDATE subscriptions
		SET service_name = $1,
		    price        = $2,
		    start_date   = $3,
		    end_date     = $4,
		    updated_at   = NOW()
		WHERE id = $5
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`

	row := r.db.QueryRow(ctx, query,
		sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, sub.ID,
	)

	var updated model.Subscription
	err := row.Scan(
		&updated.ID, &updated.ServiceName, &updated.Price, &updated.UserID,
		&updated.StartDate, &updated.EndDate, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update subscription: %w", err)
	}
	slog.Info("subscription updated", "id", updated.ID)
	return &updated, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}
	slog.Info("subscription deleted", "id", id)
	return nil
}

func (r *Repository) TotalCost(ctx context.Context, userID, serviceName, from, to string) (int, error) {
	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE ($1 = '' OR user_id::text = $1)
		  AND ($2 = '' OR service_name ILIKE $2)
		  AND ($3 = '' OR start_date >= $3::date)
		  AND ($4 = '' OR start_date <= $4::date)`

	var total int
	err := r.db.QueryRow(ctx, query, userID, serviceName, from, to).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("total cost: %w", err)
	}
	slog.Info("total cost calculated", "total", total)
	return total, nil
}
