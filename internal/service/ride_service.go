package service

import (
	"context"
	"errors"

	"github.com/AbhinavJain1234/matchit/internal/models"
	"github.com/AbhinavJain1234/matchit/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrInvalidRiderID  = errors.New("rider_id is required")
	ErrRideNotFound    = repository.ErrRideNotFound
	ErrAlreadyAssigned = repository.ErrAlreadyAssigned
)

type CreateRideRequest struct {
	RiderID   string
	PickupLat float64
	PickupLon float64
	DestLat   float64
	DestLon   float64
}

type CreateRideResponse struct {
	Ride    models.Ride
	Drivers []models.NearbyDriver
}

// RideService contains ride-related business logic.
type RideService struct {
	rideRepo      *repository.RideRepository
	driverService *DriverService
}

func NewRideService(rideRepo *repository.RideRepository, driverService *DriverService) *RideService {
	return &RideService{rideRepo: rideRepo, driverService: driverService}
}

func (s *RideService) CreateRideRequest(ctx context.Context, req CreateRideRequest) (CreateRideResponse, error) {
	if req.RiderID == "" {
		return CreateRideResponse{}, ErrInvalidRiderID
	}

	ride := models.Ride{
		ID:        uuid.New().String(),
		RiderID:   req.RiderID,
		PickupLat: req.PickupLat,
		PickupLon: req.PickupLon,
		DestLat:   req.DestLat,
		DestLon:   req.DestLon,
		Status:    models.RideStatusRequested,
	}

	if err := s.rideRepo.Save(ride); err != nil {
		return CreateRideResponse{}, err
	}

	drivers, err := s.driverService.FindNearbyDrivers(ctx, ride.PickupLat, ride.PickupLon, 2, 20)
	if err != nil {
		return CreateRideResponse{}, err
	}

	return CreateRideResponse{
		Ride:    ride,
		Drivers: drivers,
	}, nil
}

// GetRideByID fetches a ride by its ID.
// Returns ErrRideNotFound if the ride does not exist.
func (s *RideService) GetRideByID(ctx context.Context, rideID string) (models.Ride, error) {
	return s.rideRepo.GetByID(rideID)
}

// AcceptRide attempts to assign a driver to a ride.
// Returns ErrAlreadyAssigned if another driver accepted first (race condition handled).
func (s *RideService) AcceptRide(ctx context.Context, rideID, driverID string) (models.Ride, error) {
	if rideID == "" || driverID == "" {
		return models.Ride{}, errors.New("ride_id and driver_id are required")
	}
	return s.rideRepo.AssignDriver(rideID, driverID)
}
