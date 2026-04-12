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
	query := `INSERT INTO subscriptions
	           (device_id, revenuecat_id, store_transaction_id, platform, product_id, plan, status, expires_at, environment, raw_receipt)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	           ON CONFLICT (store_transaction_id) DO UPDATE SET
	             device_id = EXCLUDED.device_id,
	             revenuecat_id = EXCLUDED.revenuecat_id,
	             platform = EXCLUDED.platform,
	             product_id = EXCLUDED.product_id,
	             plan = EXCLUDED.plan,
	             status = EXCLUDED.status,
	             expires_at = EXCLUDED.expires_at,
	             environment = EXCLUDED.environment,
	             raw_receipt = EXCLUDED.raw_receipt,
	             updated_at = now()
	           RETURNING id, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		sub.DeviceID,
		sub.RevenueCatID,
		sub.StoreTxID,
		sub.Platform,
		sub.ProductID,
		sub.Plan,
		sub.Status,
		sub.ExpiresAt,
		sub.Environment,
		sub.RawReceipt,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO device_subscriptions (device_id, subscription_id)
		 VALUES ($1, $2)
		 ON CONFLICT (device_id, subscription_id) DO NOTHING`,
		sub.DeviceID, sub.ID,
	)
	return err
}

func (r *SubscriptionRepo) FindActiveByDeviceID(ctx context.Context, deviceID uuid.UUID) (*model.Subscription, error) {
	var sub model.Subscription
	query := `SELECT s.*
	           FROM subscriptions s
	           JOIN device_subscriptions ds ON ds.subscription_id = s.id
	           WHERE ds.device_id = $1
	             AND s.status = 'active'
	             AND s.expires_at > now()
	           ORDER BY s.expires_at DESC
	           LIMIT 1`
	if err := r.db.GetContext(ctx, &sub, query, deviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriptionRepo) DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM device_subscriptions WHERE device_id = $1`, deviceID)
	return err
}
