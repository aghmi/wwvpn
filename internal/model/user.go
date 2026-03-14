package model

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	DeviceID  string     `db:"device_id" json:"device_id"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"-"`
}
