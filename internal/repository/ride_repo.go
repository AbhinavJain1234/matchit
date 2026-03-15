package repository

import (
	"errors"
	"sync"

	"github.com/AbhinavJain1234/matchit/internal/models"
)

var ErrRideNotFound = errors.New("ride not found")
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
