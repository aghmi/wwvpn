package mock

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/wwvpn/wwvpn/internal/model"
)

type ServerRepo struct {
	mock.Mock
}

func (m *ServerRepo) FindAll(ctx context.Context) ([]model.Server, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Server), args.Error(1)
}

func (m *ServerRepo) FindActive(ctx context.Context) ([]model.Server, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Server), args.Error(1)
}

func (m *ServerRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Server, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Server), args.Error(1)
}

func (m *ServerRepo) Create(ctx context.Context, server *model.Server) error {
	return m.Called(ctx, server).Error(0)
}

func (m *ServerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *ServerRepo) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return m.Called(ctx, id, active).Error(0)
}

func (m *ServerRepo) CountActiveConnections(ctx context.Context, serverID uuid.UUID) (int, error) {
	args := m.Called(ctx, serverID)
	return args.Int(0), args.Error(1)
}
