package models

// Driver represents a driver in the system.
type Driver struct {
	ID        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    string  `json:"status"`
}

// NearbyDriver is the response shape for proximity searches.
type NearbyDriver struct {
	DriverID   string  `json:"driver_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	DistanceKM float64 `json:"distance_km"`
}
