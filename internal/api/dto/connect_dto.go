package dto

type ConnectRequest struct {
	ServerID        string `json:"server_id" binding:"required,uuid"`
	Protocol        string `json:"protocol" binding:"required,oneof=amnezia_wg vless_reality auto"`
	ClientPublicKey string `json:"client_public_key" binding:"required"`
}

type ConnectResponse struct {
	Protocol        string              `json:"protocol"`
	Endpoint        string              `json:"endpoint"`
	ServerPublicKey string              `json:"server_public_key"`
	AssignedIP      string              `json:"assigned_ip"`
	AllowedIPs      string              `json:"allowed_ips"`
	DNS             string              `json:"dns"`
	MTU             int                 `json:"mtu"`
	ConnectionID    string              `json:"connection_id"`
	Obfuscation     *ObfuscationParams  `json:"obfuscation,omitempty"`
	VLESS           *VLESSParams        `json:"vless,omitempty"`
	Fallback        *FallbackConfig     `json:"fallback,omitempty"`
}

type ObfuscationParams struct {
	Jc   int `json:"jc"`
	Jmin int `json:"jmin"`
	Jmax int `json:"jmax"`
	S1   int `json:"s1"`
	S2   int `json:"s2"`
	H1   int `json:"h1"`
	H2   int `json:"h2"`
	H3   int `json:"h3"`
	H4   int `json:"h4"`
}

type VLESSParams struct {
	UUID      string `json:"uuid"`
	PublicKey string `json:"public_key"`
	ShortID   string `json:"short_id"`
	SNI       string `json:"sni"`
	Port      int    `json:"port"`
}

type FallbackConfig struct {
	Protocol        string             `json:"protocol"`
	ConnectionID    string             `json:"connection_id"`
	Endpoint        string             `json:"endpoint,omitempty"`
	ServerPublicKey string             `json:"server_public_key,omitempty"`
	AssignedIP      string             `json:"assigned_ip,omitempty"`
	Obfuscation     *ObfuscationParams `json:"obfuscation,omitempty"`
	VLESS           *VLESSParams       `json:"vless,omitempty"`
}

type DisconnectRequest struct {
	ConnectionID string `json:"connection_id" binding:"required,uuid"`
}
