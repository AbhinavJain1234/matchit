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

// FindNearbyDrivers returns drivers near a point sorted by ascending distance.
func (r *LocationRepository) FindNearbyDrivers(ctx context.Context, lat, lon, radiusKM float64, count int) ([]redis.GeoLocation, error) {
	return r.rdb.GeoSearchLocation(ctx, "drivers", &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude: lon,
			Latitude: lat,
			Radius: radiusKM,
			RadiusUnit: "km",
			Sort: "ASC",
			Count: count,
		},
		WithCoord: true,
		WithDist:  true,
	}).Result()
}
