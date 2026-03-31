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
	ErrRideNotCancelable = repository.ErrRideNotCancelable
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
	log.Printf("[Ride %s] Starting driver search at (%.4f, %.4f)", rideID, lat, lon)

	const notifyBatchSize = 10
	const maxFetchFromDB = 1000
	const maxRounds = 5
	const roundTimeout = 60 * time.Second

	triedDriverIDs := make(map[string]struct{})
	previousRoundNotified := 0

	for round := 0; round < maxRounds; round++ {
		log.Printf("[Round %d] Starting search for ride %s", round, rideID)
		
		isAvailable, err := s.rideRepo.IsRideAvailable(ctx, rideID)
		if err != nil {
			return err
		}
		if !isAvailable {
			return ErrRideCancelled
		}

		// If previous round had 0 drivers, cancel immediately
		if round > 0 && previousRoundNotified == 0 {
			log.Printf("[Round %d] Previous round found 0 drivers - cancelling ride", round)
			break
		}

		notifiedThisRound := 0
		var radiusesToTry []float64

		// Start from base radius 2km, then expand if needed
		if round == 0 {
			// First round: start at 2km and expand if needed to reach 10
			radiusesToTry = []float64{2, 3, 5, 10}
		} else if previousRoundNotified >= notifyBatchSize {
			// Previous round had full batch: expand search to new areas
			radiusesToTry = []float64{3, 5, 10}
		} else {
			// Previous round had partial batch: try to fill up within same round
			radiusesToTry = []float64{2, 3, 5, 10}
		}

		// Try each radius sequentially, accumulating drivers until we hit batch size
		for _, radius := range radiusesToTry {
			if notifiedThisRound >= notifyBatchSize {
				log.Printf("[Round %d] Reached batch size of %d - stopping radius expansion", round, notifyBatchSize)
				break // Stop if we've already notified enough drivers
			}

			needed := notifyBatchSize - notifiedThisRound
			notified, err := s.notifyFromRadius(ctx, lat, lon, radius, needed, maxFetchFromDB, triedDriverIDs)
			if err != nil {
				log.Printf("[Round %d] Radius %.0fkm: error - %v", round, radius, err)
				continue
			}

			if notified > 0 {
				log.Printf("[Round %d] Radius %.0fkm: notified %d drivers (needed %d, total this round: %d)", 
					round, radius, notified, needed, notifiedThisRound+notified)
				notifiedThisRound += notified
			} else {
				log.Printf("[Round %d] Radius %.0fkm: no new drivers found", round, radius)
			}
		}

		log.Printf("[Round %d] Final: Notified %d drivers total", round, notifiedThisRound)
		previousRoundNotified = notifiedThisRound

		// If we got a full batch, wait before next round
		if notifiedThisRound >= notifyBatchSize && round < maxRounds-1 {
			log.Printf("[Round %d] Got full batch - waiting %v before next round...", round, roundTimeout)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(roundTimeout):
			}
		}
	}
	
	err := s.CancelRideRequest(ctx, rideID, riderID)
	if err != nil {
		return err
	}

	return ErrNoDriversAvailable
}

func (s *RideService) notifyFromRadius(ctx context.Context, lat, lon, radiusKM float64, notifyBatchSize int, maxFetchFromDB int, triedDriverIDs map[string]struct{}) (int, error) {
	nearbyDrivers, err := s.driverService.FindNearbyDrivers(ctx, lat, lon, radiusKM, maxFetchFromDB)
	if err != nil {
		return 0, errors.New("unable to fetch drivers")
	}

	if len(nearbyDrivers) == 0 {
		return 0, errors.New("no drivers found in radius")
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

		print("round count ",driver.DriverID, " distance ", driver.DistanceKM, "\n")
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

func (s *RideService) CancelRideRequest(ctx context.Context, rideID, riderID string) ( error) {
	if rideID == "" || riderID == "" {
		return errors.New("ride_id and rider_id are required")
	}

	hasActiveRide, err := s.rideRepo.HasActiveRide(ctx, riderID)
	if err != nil {
		return  err
	}
	if !hasActiveRide {
		return ErrRideNotCancelable
	}

	return s.rideRepo.CancelRideRequest(ctx, rideID, riderID)
}
