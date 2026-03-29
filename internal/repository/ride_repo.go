package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/AbhinavJain1234/matchit/internal/models"
)

var (
	ErrRideNotFound     = errors.New("ride not found")
	ErrAlreadyAssigned  = errors.New("ride already assigned to another driver")
	ErrActiveRideExists = errors.New("rider already has an active ride")
	ErrRideNotCancelable = errors.New("ride cannot be cancelled")
)

// RideRepository defines the data access contract for rides.
// Both InMemoryRideRepository and PostgresRideRepository satisfy this interface.
type RideRepository interface {
	Save(ctx context.Context, ride models.Ride) error
	GetByID(ctx context.Context, id string) (models.Ride, error)
	AssignDriver(ctx context.Context, rideID, driverID string) (models.Ride, error)
	CancelRideRequest(ctx context.Context, rideID, riderID string) (error)
	HasActiveRide(ctx context.Context, riderID string) (bool, error)
	IsRideAvailable(ctx context.Context, rideID string) (bool, error)
}

func (r *InMemoryRideRepository) IsRideAvailable(ctx context.Context, rideID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ride, ok := r.rides[rideID]
	if !ok {
		return false, ErrRideNotFound
	}
	return ride.DriverID == "" && ride.Status == models.RideStatusRequested, nil
}

// InMemoryRideRepository is a thread-safe in-memory implementation used when no database is configured.
type InMemoryRideRepository struct {
	mu    sync.RWMutex
	rides map[string]models.Ride
}

func NewInMemoryRideRepository() *InMemoryRideRepository {
	return &InMemoryRideRepository{rides: make(map[string]models.Ride)}
}

func (r *InMemoryRideRepository) Save(_ context.Context, ride models.Ride) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rides[ride.ID] = ride
	return nil
}

func (r *InMemoryRideRepository) GetByID(_ context.Context, id string) (models.Ride, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ride, ok := r.rides[id]
	if !ok {
		return models.Ride{}, ErrRideNotFound
	}
	return ride, nil
}

// AssignDriver atomically assigns a driver only if no driver is set yet.
// Mirrors: UPDATE rides SET driver_id=$1, status=$2 WHERE id=$3 AND driver_id IS NULL
func (r *InMemoryRideRepository) AssignDriver(_ context.Context, rideID, driverID string) (models.Ride, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ride, ok := r.rides[rideID]
	if !ok {
		return models.Ride{}, ErrRideNotFound
	}
	// TODO: Intentionally deferred for now.
	// Add status validation so assignment is allowed only when ride.Status == REQUESTED.
	if ride.DriverID != "" {
		return models.Ride{}, ErrAlreadyAssigned
	}

	if ride.Status != models.RideStatusRequested {
		return models.Ride{}, errors.New("ride is not in a state to be accepted")
	}
	
	ride.DriverID = driverID
	ride.Status = models.RideStatusDriverAssigned
	r.rides[rideID] = ride
	return ride, nil
}

func (r *InMemoryRideRepository) HasActiveRide(_ context.Context, riderID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, ride := range r.rides {
		if ride.RiderID == riderID && (ride.Status == models.RideStatusDriverAssigned || ride.Status == models.RideStatusRequested || ride.Status == models.RideStatusInProgress) {
			return true, nil
		}
	}
	return false, nil
}

func (r *InMemoryRideRepository) CancelRideRequest(_ context.Context, rideID, riderID string) (error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ride, ok := r.rides[rideID]
	if !ok {
		return ErrRideNotFound
	}
	if ride.RiderID != riderID {
		return errors.New("rider_id does not match the ride's rider")
	}
	if ride.Status != models.RideStatusRequested {
		return ErrRideNotCancelable
	}
	ride.Status = models.RideStatusCancelled
	r.rides[rideID] = ride
	return nil
}
