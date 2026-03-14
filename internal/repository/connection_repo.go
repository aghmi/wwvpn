package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/wwvpn/wwvpn/internal/model"
)

type ConnectionRepo struct {
	db *sqlx.DB
}

func NewConnectionRepo(db *sqlx.DB) *ConnectionRepo {
	return &ConnectionRepo{db: db}
}

func (r *ConnectionRepo) Create(ctx context.Context, conn *model.Connection) error {
	query := `INSERT INTO connections (device_id, server_id, protocol, client_public_key, vless_uuid)
	           VALUES ($1, $2, $3, $4, $5)
	           RETURNING id, connected_at`
	return r.db.QueryRowxContext(ctx, query,
		conn.DeviceID, conn.ServerID, conn.Protocol, conn.ClientPublicKey, conn.VlessUUID,
	).Scan(&conn.ID, &conn.ConnectedAt)
}

func (r *ConnectionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Connection, error) {
	var conn model.Connection
	if err := r.db.GetContext(ctx, &conn, `SELECT * FROM connections WHERE id = $1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &conn, nil
}

func (r *ConnectionRepo) Disconnect(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE connections SET disconnected_at = now() WHERE id = $1`, id)
	return err
}

func (r *ConnectionRepo) DisconnectAllByDevice(ctx context.Context, deviceID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE connections SET disconnected_at = now() WHERE device_id = $1 AND disconnected_at IS NULL`,
		deviceID)
	return err
}

func (r *ConnectionRepo) CountActiveByDevice(ctx context.Context, deviceID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM connections WHERE device_id = $1 AND disconnected_at IS NULL`,
		deviceID)
	return count, err
}

func (r *ConnectionRepo) FindActiveByDevice(ctx context.Context, deviceID uuid.UUID) ([]model.Connection, error) {
	var conns []model.Connection
	err := r.db.SelectContext(ctx, &conns,
		`SELECT * FROM connections WHERE device_id = $1 AND disconnected_at IS NULL`,
		deviceID)
	return conns, err
}
