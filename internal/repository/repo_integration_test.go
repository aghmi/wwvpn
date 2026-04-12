//go:build integration

package repository

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/wwvpn/wwvpn/internal/model"
)

func migrationsDir() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "..", "..", "migrations")
}

func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("wwvpn_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts(
			filepath.Join(migrationsDir(), "000001_init.up.sql"),
			filepath.Join(migrationsDir(), "000002_connections_keys.up.sql"),
			filepath.Join(migrationsDir(), "000003_iap_subscriptions.up.sql"),
		),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sqlx.Connect("postgres", connStr)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	return db
}

func TestDeviceRepo_Integration(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeviceRepo(db)
	ctx := context.Background()

	device, err := repo.FindOrCreate(ctx, "integration-device-1")
	require.NoError(t, err)
	assert.Equal(t, "integration-device-1", device.DeviceID)
	assert.NotEqual(t, uuid.Nil, device.ID)

	same, err := repo.FindOrCreate(ctx, "integration-device-1")
	require.NoError(t, err)
	assert.Equal(t, device.ID, same.ID)

	found, err := repo.FindByID(ctx, device.ID)
	require.NoError(t, err)
	assert.Equal(t, device.DeviceID, found.DeviceID)

	err = repo.SoftDelete(ctx, device.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, device.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestServerRepo_Integration(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepo(db)
	ctx := context.Background()

	server := &model.Server{
		Country:          "NL",
		City:             "Amsterdam",
		Hostname:         "nl-ams-01",
		IPAddress:        "1.2.3.4",
		GRPCPort:         50051,
		AWGPort:          51820,
		VLESSPort:        443,
		AWGPublicKey:     "pubkey",
		VLESSPublicKey:   "vlesspub",
		VLESSShortID:     "abcdef",
		CamouflageDomain: "www.microsoft.com",
		Capacity:         100,
		IsActive:         true,
	}
	err := repo.Create(ctx, server)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, server.ID)

	found, err := repo.FindByID(ctx, server.ID)
	require.NoError(t, err)
	assert.Equal(t, "NL", found.Country)

	active, err := repo.FindActive(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 1)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	err = repo.SetActive(ctx, server.ID, false)
	require.NoError(t, err)
	active, err = repo.FindActive(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 0)

	count, err := repo.CountActiveConnections(ctx, server.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	err = repo.Delete(ctx, server.ID)
	require.NoError(t, err)
	_, err = repo.FindByID(ctx, server.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestConnectionRepo_Integration(t *testing.T) {
	db := setupTestDB(t)
	deviceRepo := NewDeviceRepo(db)
	serverRepo := NewServerRepo(db)
	connRepo := NewConnectionRepo(db)
	ctx := context.Background()

	device, err := deviceRepo.FindOrCreate(ctx, "conn-test-device")
	require.NoError(t, err)

	server := &model.Server{
		Country: "DE", City: "Frankfurt", Hostname: "de-01",
		IPAddress: "5.6.7.8", GRPCPort: 50051, AWGPort: 51820, VLESSPort: 443,
		AWGPublicKey: "k1", VLESSPublicKey: "k2", VLESSShortID: "sid",
		CamouflageDomain: "example.com", Capacity: 50, IsActive: true,
	}
	err = serverRepo.Create(ctx, server)
	require.NoError(t, err)

	pubKey := "client-pub-key"
	conn := &model.Connection{
		DeviceID:        device.ID,
		ServerID:        server.ID,
		Protocol:        model.ProtocolAmneziaWG,
		ClientPublicKey: &pubKey,
	}
	err = connRepo.Create(ctx, conn)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, conn.ID)

	found, err := connRepo.FindByID(ctx, conn.ID)
	require.NoError(t, err)
	assert.Equal(t, model.ProtocolAmneziaWG, found.Protocol)

	count, err := connRepo.CountActiveByDevice(ctx, device.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	active, err := connRepo.FindActiveByDevice(ctx, device.ID)
	require.NoError(t, err)
	assert.Len(t, active, 1)

	err = connRepo.Disconnect(ctx, conn.ID)
	require.NoError(t, err)

	count, err = connRepo.CountActiveByDevice(ctx, device.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestSubscriptionRepo_Integration(t *testing.T) {
	db := setupTestDB(t)
	deviceRepo := NewDeviceRepo(db)
	subRepo := NewSubscriptionRepo(db)
	ctx := context.Background()

	device, err := deviceRepo.FindOrCreate(ctx, "sub-test-device")
	require.NoError(t, err)

	sub := &model.Subscription{
		DeviceID:     device.ID,
		RevenueCatID: "rc-123",
		Plan:         model.PlanMonthly,
		StoreTxID:    "legacy-test-store-tx-1",
		Platform:     "legacy",
		ProductID:    model.PlanMonthly,
		Status:       model.StatusActive,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
	}
	err = subRepo.Upsert(ctx, sub)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, sub.ID)

	found, err := subRepo.FindActiveByDeviceID(ctx, device.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PlanMonthly, found.Plan)

	sub.Plan = model.PlanAnnual
	err = subRepo.Upsert(ctx, sub)
	require.NoError(t, err)

	found, err = subRepo.FindActiveByDeviceID(ctx, device.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PlanAnnual, found.Plan)

	err = subRepo.DeleteByDeviceID(ctx, device.ID)
	require.NoError(t, err)
	_, err = subRepo.FindActiveByDeviceID(ctx, device.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}
