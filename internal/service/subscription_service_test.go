package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wwvpn/wwvpn/internal/model"
	"github.com/wwvpn/wwvpn/internal/repository"
	mockRepo "github.com/wwvpn/wwvpn/internal/repository/mock"
)

func TestSubscriptionService_CanConnect_Active(t *testing.T) {
	subRepo := new(mockRepo.SubscriptionRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := NewSubscriptionService(subRepo, connRepo, nil, nil)
	ctx := context.Background()
	deviceID := uuid.New()

	sub := &model.Subscription{
		ID:        uuid.New(),
		DeviceID:  deviceID,
		Plan:      model.PlanMonthly,
		Status:    model.StatusActive,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	subRepo.On("FindActiveByDeviceID", ctx, deviceID).Return(sub, nil)
	connRepo.On("CountActiveByDevice", ctx, deviceID).Return(0, nil)

	can, err := svc.CanConnect(ctx, deviceID)
	require.NoError(t, err)
	assert.True(t, can)
}

func TestSubscriptionService_CanConnect_LimitReached(t *testing.T) {
	subRepo := new(mockRepo.SubscriptionRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := NewSubscriptionService(subRepo, connRepo, nil, nil)
	ctx := context.Background()
	deviceID := uuid.New()

	sub := &model.Subscription{
		ID:        uuid.New(),
		DeviceID:  deviceID,
		Plan:      model.PlanMonthly,
		Status:    model.StatusActive,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	subRepo.On("FindActiveByDeviceID", ctx, deviceID).Return(sub, nil)
	connRepo.On("CountActiveByDevice", ctx, deviceID).Return(1, nil)

	can, err := svc.CanConnect(ctx, deviceID)
	require.NoError(t, err)
	assert.False(t, can)
}

func TestSubscriptionService_CanConnect_NoSubscription(t *testing.T) {
	subRepo := new(mockRepo.SubscriptionRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := NewSubscriptionService(subRepo, connRepo, nil, nil)
	ctx := context.Background()
	deviceID := uuid.New()

	subRepo.On("FindActiveByDeviceID", ctx, deviceID).Return(nil, repository.ErrNotFound)

	_, err := svc.CanConnect(ctx, deviceID)
	assert.Error(t, err)
}

func TestSubscriptionService_CanConnect_AnnualPlan3Conns(t *testing.T) {
	subRepo := new(mockRepo.SubscriptionRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := NewSubscriptionService(subRepo, connRepo, nil, nil)
	ctx := context.Background()
	deviceID := uuid.New()

	sub := &model.Subscription{
		ID:        uuid.New(),
		DeviceID:  deviceID,
		Plan:      model.PlanAnnual,
		Status:    model.StatusActive,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	subRepo.On("FindActiveByDeviceID", ctx, deviceID).Return(sub, nil)
	connRepo.On("CountActiveByDevice", ctx, deviceID).Return(2, nil)

	can, err := svc.CanConnect(ctx, deviceID)
	require.NoError(t, err)
	assert.True(t, can)
}

func TestSubscriptionService_GetActive(t *testing.T) {
	subRepo := new(mockRepo.SubscriptionRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := NewSubscriptionService(subRepo, connRepo, nil, nil)
	ctx := context.Background()
	deviceID := uuid.New()

	sub := &model.Subscription{
		ID:       uuid.New(),
		DeviceID: deviceID,
		Plan:     model.PlanMonthly,
		Status:   model.StatusActive,
	}
	subRepo.On("FindActiveByDeviceID", ctx, deviceID).Return(sub, nil)

	result, err := svc.GetActive(ctx, deviceID)
	require.NoError(t, err)
	assert.Equal(t, sub.ID, result.ID)
}
