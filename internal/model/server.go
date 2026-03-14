package model

import (
	"time"

	"github.com/google/uuid"
)

type Server struct {
	ID               uuid.UUID `db:"id" json:"id"`
	Country          string    `db:"country" json:"country"`
	City             string    `db:"city" json:"city"`
	Hostname         string    `db:"hostname" json:"hostname"`
	IPAddress        string    `db:"ip_address" json:"ip_address"`
	GRPCPort         int       `db:"grpc_port" json:"-"`
	AWGPort          int       `db:"awg_port" json:"awg_port"`
	VLESSPort        int       `db:"vless_port" json:"vless_port"`
	AWGPublicKey     string    `db:"awg_public_key" json:"-"`
	VLESSPublicKey   string    `db:"vless_public_key" json:"-"`
	VLESSShortID     string    `db:"vless_short_id" json:"-"`
	CamouflageDomain string   `db:"camouflage_domain" json:"-"`
	Capacity         int       `db:"capacity" json:"-"`
	IsActive         bool      `db:"is_active" json:"is_active"`
	IsPremium        bool      `db:"is_premium" json:"is_premium"`
	CreatedAt        time.Time `db:"created_at" json:"-"`
}
