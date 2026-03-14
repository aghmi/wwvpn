package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wwvpn/wwvpn/internal/node"
	"github.com/wwvpn/wwvpn/internal/repository"
)

type HealthMonitor struct {
	serverRepo  repository.ServerRepository
	newClient   node.NodeClientFactory
	interval    time.Duration
	maxFailures int
	mu          sync.Mutex
	failCounts  map[uuid.UUID]int
	wasInactive map[uuid.UUID]bool
}

func NewHealthMonitor(serverRepo repository.ServerRepository, factory node.NodeClientFactory) *HealthMonitor {
	return &HealthMonitor{
		serverRepo:  serverRepo,
		newClient:   factory,
		interval:    30 * time.Second,
		maxFailures: 3,
		failCounts:  make(map[uuid.UUID]int),
		wasInactive: make(map[uuid.UUID]bool),
	}
}

func (m *HealthMonitor) Start(ctx context.Context) {
	log.Println("[health-monitor] started")
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[health-monitor] stopped")
			return
		case <-ticker.C:
			m.checkAll(ctx)
		}
	}
}

func (m *HealthMonitor) checkAll(ctx context.Context) {
	servers, err := m.serverRepo.FindAll(ctx)
	if err != nil {
		log.Printf("[health-monitor] failed to fetch servers: %v", err)
		return
	}

	for _, srv := range servers {
		go m.checkOne(ctx, srv.ID, fmt.Sprintf("%s:%d", srv.IPAddress, srv.GRPCPort))
	}
}

func (m *HealthMonitor) checkOne(ctx context.Context, serverID uuid.UUID, addr string) {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	healthy := false
	client, err := m.newClient(addr)
	if err == nil {
		defer client.Close()
		h, _, hErr := client.HealthCheck(checkCtx)
		healthy = hErr == nil && h
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if healthy {
		if m.failCounts[serverID] > 0 {
			log.Printf("[health-monitor] server %s recovered", serverID)
		}
		m.failCounts[serverID] = 0

		if m.wasInactive[serverID] {
			if err := m.serverRepo.SetActive(context.Background(), serverID, true); err != nil {
				log.Printf("[health-monitor] failed to reactivate %s: %v", serverID, err)
			} else {
				log.Printf("[health-monitor] server %s reactivated", serverID)
				m.wasInactive[serverID] = false
			}
		}
		return
	}

	m.failCounts[serverID]++
	log.Printf("[health-monitor] server %s failed check (%d/%d)", serverID, m.failCounts[serverID], m.maxFailures)

	if m.failCounts[serverID] >= m.maxFailures && !m.wasInactive[serverID] {
		if err := m.serverRepo.SetActive(context.Background(), serverID, false); err != nil {
			log.Printf("[health-monitor] failed to deactivate %s: %v", serverID, err)
		} else {
			log.Printf("[health-monitor] server %s deactivated (unreachable)", serverID)
			m.wasInactive[serverID] = true
		}
	}
}
