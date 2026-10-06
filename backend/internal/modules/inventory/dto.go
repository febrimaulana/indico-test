package inventory

// ReserveRequest is the POST /inventory/reserve payload.
type ReserveRequest struct {
	UserID   string `json:"user_id" binding:"required,min=3,max=64,identifier" example:"usr_9981"`
	ItemID   string `json:"item_id" binding:"required,min=3,max=64,identifier" example:"item_4021"`
	Quantity int    `json:"quantity" binding:"gt=0,max=1000" example:"2"`
}

// ReserveResponse is the success payload for reserve.
type ReserveResponse struct {
	Status        string `json:"status" example:"success"`
	ReservationID string `json:"reservation_id" example:"res_883291"`
	ItemID        string `json:"item_id" example:"item_4021"`
	Quantity      int    `json:"quantity" example:"2"`
	ExpiresAt     string `json:"expires_at" example:"2026-07-20T16:35:00Z"`
}

// ConfirmRequest is the POST /inventory/confirm payload.
type ConfirmRequest struct {
	ReservationID string `json:"reservation_id" binding:"required,min=3,max=64,identifier" example:"res_883291"`
}

// ConfirmResponse is the success payload for confirm.
type ConfirmResponse struct {
	Status        string `json:"status" example:"success"`
	ReservationID string `json:"reservation_id" example:"res_883291"`
	ConfirmedAt   string `json:"confirmed_at" example:"2026-07-20T16:32:00Z"`
}

// StockQuery is the GET /inventory/stock query string.
type StockQuery struct {
	ItemID string `form:"item_id" binding:"required,min=3,max=64,identifier" example:"item_4021"`
}

// StockResponse is the success payload for stock.
type StockResponse struct {
	ItemID         string `json:"item_id" example:"item_4021"`
	TotalStock     int    `json:"total_stock" example:"100"`
	ReservedStock  int    `json:"reserved_stock" example:"15"`
	AvailableStock int    `json:"available_stock" example:"85"`
}
