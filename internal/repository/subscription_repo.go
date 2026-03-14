package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/wwvpn/wwvpn/internal/model"
)

type SubscriptionRepo struct {
	db *sqlx.DB
}

func NewSubscriptionRepo(db *sqlx.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) Upsert(ctx context.Context, sub *model.Subscription) error {
	query := `INSERT INTO subscriptions (device_id, revenuecat_id, plan, status, expires_at)
	           VALUES ($1, $2, $3, $4, $5)
	           ON CONFLICT (device_id) DO UPDATE SET
	             revenuecat_id = EXCLUDED.revenuecat_id,
	             plan = EXCLUDED.plan,
	             status = EXCLUDED.status,
	             expires_at = EXCLUDED.expires_at,
	             updated_at = now()
	           RETURNING id, created_at, updated_at`
	return r.db.QueryRowxContext(ctx, query,
		sub.DeviceID, sub.RevenueCatID, sub.Plan, sub.Status, sub.ExpiresAt,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
}

func (r *SubscriptionRepo) FindActiveByDeviceID(ctx context.Context, deviceID uuid.UUID) (*model.Subscription, error) {
	var sub model.Subscription
	query := `SELECT * FROM subscriptions
	           WHERE device_id = $1 AND status = 'active' AND expires_at > now()
	           ORDER BY expires_at DESC LIMIT 1`
	if err := r.db.GetContext(ctx, &sub, query, deviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriptionRepo) DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM subscriptions WHERE device_id = $1`, deviceID)
	return err
}
