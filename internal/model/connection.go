package model

import (
	"time"

	"github.com/google/uuid"
)

type Connection struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	DeviceID        uuid.UUID  `db:"device_id" json:"device_id"`
	ServerID        uuid.UUID  `db:"server_id" json:"server_id"`
	Protocol        string     `db:"protocol" json:"protocol"`
	ClientPublicKey *string    `db:"client_public_key" json:"-"`
	VlessUUID       *string    `db:"vless_uuid" json:"-"`
	ConnectedAt     time.Time  `db:"connected_at" json:"connected_at"`
	DisconnectedAt  *time.Time `db:"disconnected_at" json:"disconnected_at,omitempty"`
}

const (
	ProtocolAmneziaWG    = "amnezia_wg"
	ProtocolVLESSReality = "vless_reality"
)
