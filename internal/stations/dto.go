package stations

type CreateStationRequest struct {
	Name         string   `json:"name" binding:"required,min=2,max=150"`
	AddressLine1 string   `json:"address_line1" binding:"required"`
	AddressLine2 *string  `json:"address_line2,omitempty"`
	City         string   `json:"city" binding:"required"`
	District     *string  `json:"district,omitempty"`
	State        string   `json:"state" binding:"required"`
	Country      *string  `json:"country,omitempty"`
	Pincode      *string  `json:"pincode,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
}

type StationResponse struct {
	ID           string   `json:"id"`
	OwnerID      *string  `json:"owner_id,omitempty"`
	Name         string   `json:"name"`
	AddressLine1 string   `json:"address_line1"`
	AddressLine2 *string  `json:"address_line2,omitempty"`
	City         string   `json:"city"`
	District     *string  `json:"district,omitempty"`
	State        string   `json:"state"`
	Country      string   `json:"country"`
	Pincode      *string  `json:"pincode,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
	IsActive     bool     `json:"is_active"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}