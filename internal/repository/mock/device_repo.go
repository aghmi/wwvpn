package mock

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/wwvpn/wwvpn/internal/model"
)

type DeviceRepo struct {
	mock.Mock
}

func (m *DeviceRepo) FindOrCreate(ctx context.Context, deviceID string) (*model.Device, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *DeviceRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *DeviceRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
