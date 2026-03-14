package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// LocationRepository handles all driver-location reads and writes against Redis.
// Nothing in this file knows about HTTP or business rules.
type LocationRepository struct {
	rdb *redis.Client
}

func NewLocationRepository(rdb *redis.Client) *LocationRepository {
	return &LocationRepository{rdb: rdb}
}

// SaveDriverLocation stores (or updates) a driver's position in the Redis GEO index.
func (r *LocationRepository) SaveDriverLocation(ctx context.Context, driverID string, lat, lon float64) error {
	return r.rdb.GeoAdd(ctx, "drivers", &redis.GeoLocation{
		Name:      driverID,
		Longitude: lon,
		Latitude:  lat,
	}).Err()
}
