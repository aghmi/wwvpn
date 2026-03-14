package dto

type ServerResponse struct {
	ID        string   `json:"id"`
	Country   string   `json:"country"`
	City      string   `json:"city"`
	IsActive  bool     `json:"is_active"`
	IsPremium bool     `json:"is_premium"`
	Load      int      `json:"load"`
	Capacity  int      `json:"capacity"`
	Protocols []string `json:"protocols"`
}

type ServerListResponse struct {
	Servers []ServerResponse `json:"servers"`
}
