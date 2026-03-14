package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/wwvpn/wwvpn/internal/model"
)

type DeviceRepository interface {
	FindOrCreate(ctx context.Context, deviceID string) (*model.Device, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Device, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type ServerRepository interface {
	FindAll(ctx context.Context) ([]model.Server, error)
	FindActive(ctx context.Context) ([]model.Server, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Server, error)
	Create(ctx context.Context, server *model.Server) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
	CountActiveConnections(ctx context.Context, serverID uuid.UUID) (int, error)
}

type SubscriptionRepository interface {
	Upsert(ctx context.Context, sub *model.Subscription) error
	FindActiveByDeviceID(ctx context.Context, deviceID uuid.UUID) (*model.Subscription, error)
	DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error
}

type ConnectionRepository interface {
	Create(ctx context.Context, conn *model.Connection) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Connection, error)
	Disconnect(ctx context.Context, id uuid.UUID) error
	DisconnectAllByDevice(ctx context.Context, deviceID uuid.UUID) error
	CountActiveByDevice(ctx context.Context, deviceID uuid.UUID) (int, error)
	FindActiveByDevice(ctx context.Context, deviceID uuid.UUID) ([]model.Connection, error)
}
