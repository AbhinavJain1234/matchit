package service

import (
	"context"
	"errors"
	"log"
	"sort"
	"time"

	"github.com/AbhinavJain1234/matchit/internal/models"
	"github.com/AbhinavJain1234/matchit/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrInvalidRiderID  = errors.New("rider_id is required")
	ErrRideNotFound    = repository.ErrRideNotFound
	ErrAlreadyAssigned = repository.ErrAlreadyAssigned
	ErrActiveRideExists = repository.ErrActiveRideExists
	ErrNoDriversAvailable = errors.New("no drivers available, try after sometime")
	ErrRideCancelled   = errors.New("ride is cancelled")
)

type CreateRideRequest struct {
	RiderID   string
	PickupLat float64
	PickupLon float64
	DestLat   float64
	DestLon   float64
}

type CreateRideResponse struct {
	Ride models.Ride
}

// RideService contains ride-related business logic.
type RideService struct {
	rideRepo      repository.RideRepository
	driverService *DriverService
}

func NewRideService(rideRepo repository.RideRepository, driverService *DriverService) *RideService {
	return &RideService{rideRepo: rideRepo, driverService: driverService}
}

func (s *RideService) CreateRideRequest(ctx context.Context, req CreateRideRequest) (CreateRideResponse, error) {
	if req.RiderID == "" {
		return CreateRideResponse{}, ErrInvalidRiderID
	}

	if hasActiveRide, err := s.rideRepo.HasActiveRide(ctx, req.RiderID); err != nil {
		return CreateRideResponse{}, err
	} else if hasActiveRide {
		return CreateRideResponse{}, repository.ErrActiveRideExists
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

	if err := s.rideRepo.Save(ctx, ride); err != nil {
		return CreateRideResponse{}, err
	}

	go s.findNearbyDriversAsync(ride)

	return CreateRideResponse{
		Ride: ride,
	}, nil
}

func (s *RideService) findNearbyDriversAsync(ride models.Ride) {
	if err := s.FindNearbyDrivers(context.Background(), ride.RiderID, ride.ID, ride.PickupLat, ride.PickupLon); err != nil {
		log.Printf("background matching failed for ride %s: %v", ride.ID, err)
	}
}

func (s *RideService) FindNearbyDrivers(ctx context.Context, riderID, rideID string, lat, lon float64) error {
	const notifyBatchSize = 10
	const maxFetchFromDB = 1000
	const maxRounds = 5
	const roundTimeout = 60 * time.Second

	triedDriverIDs := make(map[string]struct{})

	for round := 0; round < maxRounds; round++ {
		isAvailable, err := s.rideRepo.IsRideAvailable(ctx, rideID)
		if err != nil {
			return err
		}
		if !isAvailable {
			return ErrRideCancelled
		}

		notifiedThisRound, err := s.notifyFromRadius(ctx, lat, lon, 2, notifyBatchSize, maxFetchFromDB, triedDriverIDs)
		if err != nil {
			return err
		}

		if notifiedThisRound == 0 {
			notifiedThisRound, err = s.notifyFromRadius(ctx, lat, lon, 3, notifyBatchSize, maxFetchFromDB, triedDriverIDs)
			if err != nil {
				return err
			}
		}

		if notifiedThisRound == 0 {
			notifiedThisRound, err = s.notifyFromRadius(ctx, lat, lon, 5, notifyBatchSize, maxFetchFromDB, triedDriverIDs)
			if err != nil {
				return err
			}
		}

		if round < maxRounds-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(roundTimeout):
			}
		}
	}

	return ErrNoDriversAvailable
}

func (s *RideService) notifyFromRadius(
	ctx context.Context,
	lat, lon, radiusKM float64,
	notifyBatchSize int,
	maxFetchFromDB int,
	triedDriverIDs map[string]struct{},
) (int, error) {
	nearbyDrivers, err := s.driverService.FindNearbyDrivers(ctx, lat, lon, radiusKM, maxFetchFromDB)
	if err != nil {
		return 0, errors.New("unable to fetch drivers")
	}

	sort.Slice(nearbyDrivers, func(i, j int) bool {
		return nearbyDrivers[i].DistanceKM < nearbyDrivers[j].DistanceKM
	})

	notifiedCount := 0
	for _, driver := range nearbyDrivers {
		if notifiedCount >= notifyBatchSize {
			break
		}
		if _, exists := triedDriverIDs[driver.DriverID]; exists {
			continue
		}

		// TODO: Replace this with actual push notification dispatch.
		triedDriverIDs[driver.DriverID] = struct{}{}
		notifiedCount++
	}

	return notifiedCount, nil
}

// GetRideByID fetches a ride by its ID.
// Returns ErrRideNotFound if the ride does not exist.
func (s *RideService) GetRideByID(ctx context.Context, rideID string) (models.Ride, error) {
	return s.rideRepo.GetByID(ctx, rideID)
}

// AcceptRide attempts to assign a driver to a ride.
// Returns ErrAlreadyAssigned if another driver accepted first (race condition handled).
func (s *RideService) AcceptRide(ctx context.Context, rideID, driverID string) (models.Ride, error) {
	if rideID == "" || driverID == "" {
		return models.Ride{}, errors.New("ride_id and driver_id are required")
	}
	return s.rideRepo.AssignDriver(ctx, rideID, driverID)
}
