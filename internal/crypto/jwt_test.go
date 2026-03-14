package crypto

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateAndValidate(t *testing.T) {
	mgr := NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)
	deviceID := uuid.New()

	token, exp, err := mgr.Generate(deviceID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, exp > time.Now().Unix())

	claims, err := mgr.Validate(token)
	require.NoError(t, err)
	assert.Equal(t, deviceID, claims.DeviceUUID)
	assert.Equal(t, "wwvpn", claims.Issuer)
}

func TestJWTManager_InvalidToken(t *testing.T) {
	mgr := NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)

	_, err := mgr.Validate("invalid.token.here")
	assert.Error(t, err)
}

func TestJWTManager_WrongSecret(t *testing.T) {
	mgr1 := NewJWTManager("secret-one-32-chars-long!!!!!!!!", 60)
	mgr2 := NewJWTManager("secret-two-32-chars-long!!!!!!!!", 60)

	token, _, err := mgr1.Generate(uuid.New())
	require.NoError(t, err)

	_, err = mgr2.Validate(token)
	assert.Error(t, err)
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	mgr := &JWTManager{
		secret: []byte("test-secret-32-chars-long!!!!!!!!"),
		ttl:    -1 * time.Minute,
	}

	token, _, err := mgr.Generate(uuid.New())
	require.NoError(t, err)

	_, err = mgr.Validate(token)
	assert.Error(t, err)
}
