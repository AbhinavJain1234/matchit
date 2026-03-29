package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/AbhinavJain1234/matchit/internal/api"
	"github.com/AbhinavJain1234/matchit/internal/repository"
	"github.com/AbhinavJain1234/matchit/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load()

	// --- infrastructure ---
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// --- repository layer ---
	locationRepo := repository.NewLocationRepository(rdb)

	var rideRepo repository.RideRepository
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		pool, err := pgxpool.New(context.Background(), dbURL)
		if err != nil {
			log.Fatalf("cannot connect to postgres: %v", err)
		}
		if err := ensureSchema(context.Background(), pool, "configs/schema.sql"); err != nil {
			log.Fatalf("failed to apply schema: %v", err)
		}
		rideRepo = repository.NewPostgresRideRepository(pool)
		log.Println("using PostgreSQL ride repository")
	} else {
		rideRepo = repository.NewInMemoryRideRepository()
		log.Println("DATABASE_URL not set — using in-memory ride repository")
	}

	// --- service layer ---
	driverService := service.NewDriverService(locationRepo)
	rideService := service.NewRideService(rideRepo, driverService)

	// --- handler layer ---
	driverHandler := api.NewDriverHandler(driverService)
	rideHandler := api.NewRideHandler(rideService)

	// --- router ---
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "matchit",
			"status":  "ok",
		})
	})

	r.POST("/driver/location", driverHandler.UpdateLocation)
	r.GET("/drivers/nearby", driverHandler.GetNearbyDrivers)
	r.POST("/ride/request", rideHandler.CreateRideRequest)
	r.GET("/ride/:id/status", rideHandler.GetRideStatus)
	r.POST("/ride/accept", rideHandler.AcceptRide)
	r.POST("/ride/cancelRequest", rideHandler.CancelRideRequest)

	r.Run(":8080")
}

func ensureSchema(ctx context.Context, pool *pgxpool.Pool, schemaPath string) error {
	sqlBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, string(sqlBytes))
	return err
}
