package inventory

import "time"

// Reservation statuses.
const (
	ReservationStatusActive    = "active"
	ReservationStatusConfirmed = "confirmed"
	ReservationStatusExpired   = "expired"
)

// Item is a stock keeping unit.
type Item struct {
	ItemID        string    `gorm:"column:item_id;primaryKey;type:text" json:"item_id"`
	Name          string    `gorm:"column:name;type:text;not null" json:"name"`
	TotalStock    int       `gorm:"column:total_stock;not null" json:"total_stock"`
	ReservedStock int       `gorm:"column:reserved_stock;not null" json:"reserved_stock"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"-"`
}

// TableName maps Item to the items table.
func (Item) TableName() string { return "items" }

// Reservation records a stock hold that expires unless confirmed.
type Reservation struct {
	ReservationID string     `gorm:"column:reservation_id;primaryKey;type:text" json:"reservation_id"`
	UserID        string     `gorm:"column:user_id;type:text;not null;index" json:"user_id"`
	ItemID        string     `gorm:"column:item_id;type:text;not null;index" json:"item_id"`
	Quantity      int        `gorm:"column:quantity;not null" json:"quantity"`
	Status        string     `gorm:"column:status;type:text;not null" json:"status"`
	ExpiresAt     time.Time  `gorm:"column:expires_at;not null" json:"expires_at"`
	ConfirmedAt   *time.Time `gorm:"column:confirmed_at" json:"confirmed_at,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"-"`
	UpdatedAt     time.Time  `gorm:"column:updated_at" json:"-"`
}

// TableName maps Reservation to the reservations table.
func (Reservation) TableName() string { return "reservations" }
