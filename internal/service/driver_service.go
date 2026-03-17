package service

import (
	"context"

	"github.com/AbhinavJain1234/matchit/internal/models"
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

// FindNearbyDrivers returns nearby drivers with coordinates and distance.
func (s *DriverService) FindNearbyDrivers(ctx context.Context, lat, lon, radiusKM float64, limit int) ([]models.NearbyDriver, error) {
	locations, err := s.locationRepo.FindNearbyDrivers(ctx, lat, lon, radiusKM, limit)
	if err != nil {
		return nil, err
	}

	drivers := make([]models.NearbyDriver, 0, len(locations))
	for _, loc := range locations {
		drivers = append(drivers, models.NearbyDriver{
			DriverID:   loc.Name,
			Latitude:   loc.Latitude,
			Longitude:  loc.Longitude,
			DistanceKM: loc.Dist,
		})
	}

	return drivers, nil
}

//fix this later
func (s *DriverService) ChangeDriverStatus(ctx context.Context, driverID string, status string) error {
	return ErrRideNotFound
}