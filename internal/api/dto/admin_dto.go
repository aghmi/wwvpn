package dto

type AddServerRequest struct {
	Country          string `json:"country" binding:"required,len=2"`
	City             string `json:"city" binding:"required"`
	Hostname         string `json:"hostname" binding:"required"`
	IPAddress        string `json:"ip_address" binding:"required"`
	GRPCPort         int    `json:"grpc_port" binding:"required"`
	AWGPort          int    `json:"awg_port" binding:"required"`
	VLESSPort        int    `json:"vless_port" binding:"required"`
	AWGPublicKey     string `json:"awg_public_key" binding:"required"`
	VLESSPublicKey   string `json:"vless_public_key" binding:"required"`
	VLESSShortID     string `json:"vless_short_id" binding:"required"`
	CamouflageDomain string `json:"camouflage_domain"`
	Capacity         int    `json:"capacity"`
	IsPremium        bool   `json:"is_premium"`
}

type SetActiveRequest struct {
	Active bool `json:"active"`
}

type AdminServerResponse struct {
	ID               string `json:"id"`
	Country          string `json:"country"`
	City             string `json:"city"`
	Hostname         string `json:"hostname"`
	IPAddress        string `json:"ip_address"`
	IsActive         bool   `json:"is_active"`
	IsPremium        bool   `json:"is_premium"`
	ActiveConnections int   `json:"active_connections"`
}
