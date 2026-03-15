package main

import (
	"net/http"

	"github.com/AbhinavJain1234/matchit/internal/api"
	"github.com/AbhinavJain1234/matchit/internal/repository"
	"github.com/AbhinavJain1234/matchit/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	// --- infrastructure ---
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// --- repository layer ---
	locationRepo := repository.NewLocationRepository(rdb)
	rideRepo := repository.NewRideRepository()

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

	r.Run(":8080")
}
