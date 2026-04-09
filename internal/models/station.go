package models

import "time"

type Station struct {
	ID           string     `json:"id" db:"id"`
	OwnerID      *string    `json:"owner_id,omitempty" db:"owner_id"`
	Name         string     `json:"name" db:"name"`
	AddressLine1 string     `json:"address_line1" db:"address_line1"`
	AddressLine2 *string    `json:"address_line2,omitempty" db:"address_line2"`
	City         string     `json:"city" db:"city"`
	District     *string    `json:"district,omitempty" db:"district"`
	State        string     `json:"state" db:"state"`
	Country      string     `json:"country" db:"country"`
	Pincode      *string    `json:"pincode,omitempty" db:"pincode"`
	Latitude     *float64   `json:"latitude,omitempty" db:"latitude"`
	Longitude    *float64   `json:"longitude,omitempty" db:"longitude"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}