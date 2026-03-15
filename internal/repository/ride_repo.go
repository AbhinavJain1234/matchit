package repository

import (
	"errors"
	"sync"

	"github.com/AbhinavJain1234/matchit/internal/models"
)

var (
	ErrRideNotFound    = errors.New("ride not found")
	ErrAlreadyAssigned = errors.New("ride already assigned to another driver")
)

type RideRepository struct {
	mu    sync.RWMutex
	rides map[string]models.Ride
}

func NewRideRepository() *RideRepository {
	return &RideRepository{rides: make(map[string]models.Ride)}
}

func (r *RideRepository) Save(ride models.Ride) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rides[ride.ID] = ride
	return nil
}

func (r *RideRepository) GetByID(id string) (models.Ride, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ride, ok := r.rides[id]
	if !ok {
		return models.Ride{}, ErrRideNotFound
	}

	return ride, nil
}

// AssignDriver atomically assigns a driver to a ride only if no driver is assigned yet.
// This mirrors the SQL pattern:
//
//	UPDATE rides SET driver_id=$1, status='DRIVER_ASSIGNED'
//	WHERE id=$2 AND driver_id IS NULL
//
// The write lock ensures two concurrent accepts cannot both pass the empty-check.
func (r *RideRepository) AssignDriver(rideID, driverID string) (models.Ride, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ride, ok := r.rides[rideID]
	if !ok {
		return models.Ride{}, ErrRideNotFound
	}

	// If driver_id is already set, a different driver got here first.
	if ride.DriverID != "" {
		return models.Ride{}, ErrAlreadyAssigned
	}

	ride.DriverID = driverID
	ride.Status = models.RideStatusDriverAssigned
	r.rides[rideID] = ride

	return ride, nil
}
