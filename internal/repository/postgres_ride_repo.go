package repository

import (
	"context"
	"errors"
	"time"

	"github.com/AbhinavJain1234/matchit/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresRideRepository implements RideRepository using PostgreSQL.
type PostgresRideRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRideRepository(pool *pgxpool.Pool) *PostgresRideRepository {
	return &PostgresRideRepository{pool: pool}
}

func (r *PostgresRideRepository) Save(ctx context.Context, ride models.Ride) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO rides (id, rider_id, pickup_lat, pickup_lon, dest_lat, dest_lon, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		ride.ID, ride.RiderID,
		ride.PickupLat, ride.PickupLon,
		ride.DestLat, ride.DestLon,
		ride.Status, time.Now(),
	)
	 if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) &&
            pgErr.Code == "23505" &&
            pgErr.ConstraintName == "uq_rides_active_per_rider" {
            return ErrActiveRideExists
        }
        return err
    }
    return nil
}

func (r *PostgresRideRepository) GetByID(ctx context.Context, id string) (models.Ride, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, rider_id, COALESCE(driver_id, ''), pickup_lat, pickup_lon, dest_lat, dest_lon, status, created_at
		 FROM rides WHERE id = $1`,
		id,
	)
	var ride models.Ride
	err := row.Scan(
		&ride.ID, &ride.RiderID, &ride.DriverID,
		&ride.PickupLat, &ride.PickupLon,
		&ride.DestLat, &ride.DestLon,
		&ride.Status, &ride.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Ride{}, ErrRideNotFound
	}
	return ride, err
}

// AssignDriver uses a single atomic UPDATE with WHERE driver_id IS NULL.
// If the UPDATE matches zero rows, either the ride doesn't exist or another driver already claimed it.
func (r *PostgresRideRepository) AssignDriver(ctx context.Context, rideID, driverID string) (models.Ride, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE rides
		 SET driver_id = $1, status = $2
		 WHERE id = $3 AND driver_id IS NULL
		 RETURNING id, rider_id, driver_id, pickup_lat, pickup_lon, dest_lat, dest_lon, status, created_at`,
		driverID, models.RideStatusDriverAssigned, rideID,
	)
	var ride models.Ride
	err := row.Scan(
		&ride.ID, &ride.RiderID, &ride.DriverID,
		&ride.PickupLat, &ride.PickupLon,
		&ride.DestLat, &ride.DestLon,
		&ride.Status, &ride.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		// Row wasn't updated — check whether ride exists to distinguish the two error cases.
		exists, checkErr := r.rideExists(ctx, rideID)
		if checkErr != nil {
			return models.Ride{}, checkErr
		}
		if !exists {
			return models.Ride{}, ErrRideNotFound
		}
		return models.Ride{}, ErrAlreadyAssigned
	}
	return ride, err
}

func (r *PostgresRideRepository) rideExists(ctx context.Context, rideID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM rides WHERE id = $1)`, rideID).Scan(&exists)
	return exists, err
}

func (r *PostgresRideRepository) HasActiveRide(ctx context.Context, riderID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM rides
			WHERE rider_id = $1 AND status IN ($2, $3, $4)
		)`,
		riderID,
		models.RideStatusRequested,
		models.RideStatusDriverAssigned,
		models.RideStatusInProgress,
	).Scan(&exists)
	return exists, err
}