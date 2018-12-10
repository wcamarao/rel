package models

import (
	"time"
)

// Product .
type Product struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
