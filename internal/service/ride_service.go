package service

import (
	"context"
	"errors"

	"github.com/AbhinavJain1234/matchit/internal/models"
	"github.com/AbhinavJain1234/matchit/internal/repository"
	"github.com/google/uuid"
)

var ErrInvalidRiderID = errors.New("rider_id is required")

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
