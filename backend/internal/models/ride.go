package models
import "time"

// Ride status constants represent the lifecycle of a ride.
const (
	RideStatusRequested      = "REQUESTED"
	RideStatusDriverAssigned = "DRIVER_ASSIGNED"
	RideStatusInProgress     = "IN_PROGRESS"
	RideStatusCompleted      = "COMPLETED"
	RideStatusCancelled      = "CANCELLED"
)

// Ride represents a ride in the system.
type Ride struct {
	ID        string  `json:"id"`
	RiderID   string  `json:"rider_id"`
	DriverID  string  `json:"driver_id"`
	PickupLat float64 `json:"pickup_lat"`
	PickupLon float64 `json:"pickup_lon"`
	DestLat   float64 `json:"dest_lat"`
	DestLon   float64 `json:"dest_lon"`
	Status    string  `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
