package service

import (
	"context"

	"github.com/AbhinavJain1234/matchit/internal/repository"
)

// DriverService contains all business logic related to drivers.
// It delegates data access to the repository — it never touches Redis directly.
type DriverService struct {
	locationRepo *repository.LocationRepository
}

func NewDriverService(locationRepo *repository.LocationRepository) *DriverService {
	return &DriverService{locationRepo: locationRepo}
}

// UpdateLocation validates and stores a driver's current position.
// Business rules (e.g. rate limiting, status checks) will live here as the system grows.
func (s *DriverService) UpdateLocation(ctx context.Context, driverID string, lat, lon float64) error {
	return s.locationRepo.SaveDriverLocation(ctx, driverID, lat, lon)
}
