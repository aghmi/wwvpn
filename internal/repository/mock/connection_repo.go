package mock

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/wwvpn/wwvpn/internal/model"
)

type ConnectionRepo struct {
	mock.Mock
}

func (m *ConnectionRepo) Create(ctx context.Context, conn *model.Connection) error {
	return m.Called(ctx, conn).Error(0)
}

func (m *ConnectionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Connection, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Connection), args.Error(1)
}

func (m *ConnectionRepo) Disconnect(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *ConnectionRepo) DisconnectAllByDevice(ctx context.Context, deviceID uuid.UUID) error {
	return m.Called(ctx, deviceID).Error(0)
}

func (m *ConnectionRepo) CountActiveByDevice(ctx context.Context, deviceID uuid.UUID) (int, error) {
	args := m.Called(ctx, deviceID)
	return args.Int(0), args.Error(1)
}

func (m *ConnectionRepo) FindActiveByDevice(ctx context.Context, deviceID uuid.UUID) ([]model.Connection, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Connection), args.Error(1)
}
