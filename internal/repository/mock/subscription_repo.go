package mock

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/wwvpn/wwvpn/internal/model"
)

type SubscriptionRepo struct {
	mock.Mock
}

func (m *SubscriptionRepo) Upsert(ctx context.Context, sub *model.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *SubscriptionRepo) FindActiveByDeviceID(ctx context.Context, deviceID uuid.UUID) (*model.Subscription, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *SubscriptionRepo) DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	return m.Called(ctx, deviceID).Error(0)
}
