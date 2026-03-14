package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/wwvpn/wwvpn/internal/model"
)

var ErrNotFound = errors.New("not found")

type DeviceRepo struct {
	db *sqlx.DB
}

func NewDeviceRepo(db *sqlx.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) FindOrCreate(ctx context.Context, deviceID string) (*model.Device, error) {
	var device model.Device
	query := `SELECT * FROM devices WHERE device_id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &device, query, deviceID)
	if err == nil {
		return &device, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	insertQuery := `INSERT INTO devices (device_id) VALUES ($1) RETURNING id, device_id, created_at`
	err = r.db.QueryRowxContext(ctx, insertQuery, deviceID).Scan(&device.ID, &device.DeviceID, &device.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &device, nil
}

func (r *DeviceRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Device, error) {
	var device model.Device
	query := `SELECT * FROM devices WHERE id = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &device, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &device, nil
}

func (r *DeviceRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE devices SET deleted_at = now() WHERE id = $1`, id)
	return err
}
