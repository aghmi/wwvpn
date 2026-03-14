package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/wwvpn/wwvpn/internal/model"
)

type ServerRepo struct {
	db *sqlx.DB
}

func NewServerRepo(db *sqlx.DB) *ServerRepo {
	return &ServerRepo{db: db}
}

func (r *ServerRepo) FindAll(ctx context.Context) ([]model.Server, error) {
	var servers []model.Server
	if err := r.db.SelectContext(ctx, &servers, `SELECT * FROM servers ORDER BY country, city`); err != nil {
		return nil, err
	}
	return servers, nil
}

func (r *ServerRepo) FindActive(ctx context.Context) ([]model.Server, error) {
	var servers []model.Server
	query := `SELECT * FROM servers WHERE is_active = true ORDER BY country, city`
	if err := r.db.SelectContext(ctx, &servers, query); err != nil {
		return nil, err
	}
	return servers, nil
}

func (r *ServerRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Server, error) {
	var server model.Server
	query := `SELECT * FROM servers WHERE id = $1`
	if err := r.db.GetContext(ctx, &server, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &server, nil
}

func (r *ServerRepo) Create(ctx context.Context, server *model.Server) error {
	query := `INSERT INTO servers (country, city, hostname, ip_address, grpc_port, awg_port, vless_port,
	           awg_public_key, vless_public_key, vless_short_id, camouflage_domain, capacity, is_active, is_premium)
	           VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	           RETURNING id, created_at`
	return r.db.QueryRowxContext(ctx, query,
		server.Country, server.City, server.Hostname, server.IPAddress,
		server.GRPCPort, server.AWGPort, server.VLESSPort,
		server.AWGPublicKey, server.VLESSPublicKey, server.VLESSShortID,
		server.CamouflageDomain, server.Capacity, server.IsActive, server.IsPremium,
	).Scan(&server.ID, &server.CreatedAt)
}

func (r *ServerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM servers WHERE id = $1`, id)
	return err
}

func (r *ServerRepo) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE servers SET is_active = $2 WHERE id = $1`, id, active)
	return err
}

func (r *ServerRepo) CountActiveConnections(ctx context.Context, serverID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM connections WHERE server_id = $1 AND disconnected_at IS NULL`, serverID)
	return count, err
}
