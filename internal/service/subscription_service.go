package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/wwvpn/wwvpn/internal/model"
	"github.com/wwvpn/wwvpn/internal/repository"
)

type SubscriptionService struct {
	subRepo  repository.SubscriptionRepository
	connRepo repository.ConnectionRepository
}

func NewSubscriptionService(subRepo repository.SubscriptionRepository, connRepo repository.ConnectionRepository) *SubscriptionService {
	return &SubscriptionService{subRepo: subRepo, connRepo: connRepo}
}

func (s *SubscriptionService) GetActive(ctx context.Context, deviceUUID uuid.UUID) (*model.Subscription, error) {
	return s.subRepo.FindActiveByDeviceID(ctx, deviceUUID)
}

func (s *SubscriptionService) CanConnect(ctx context.Context, deviceUUID uuid.UUID) (bool, error) {
	sub, err := s.subRepo.FindActiveByDeviceID(ctx, deviceUUID)
	if err != nil {
		return false, err
	}
	if !sub.IsActive() {
		return false, nil
	}

	activeCount, err := s.connRepo.CountActiveByDevice(ctx, deviceUUID)
	if err != nil {
		return false, err
	}

	return activeCount < sub.MaxConnections(), nil
}

func (s *SubscriptionService) Upsert(ctx context.Context, sub *model.Subscription) error {
	return s.subRepo.Upsert(ctx, sub)
}
