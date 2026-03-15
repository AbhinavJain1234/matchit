package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/AbhinavJain1234/matchit/internal/models"
	"github.com/AbhinavJain1234/matchit/internal/service"
	"github.com/gin-gonic/gin"
)

// RideHandler owns ride-related HTTP handlers.
type RideHandler struct{
	driverService *service.DriverService
}

func NewRideHandler(driverService *service.DriverService) *RideHandler {
	return &RideHandler{
		driverService: driverService,
	}
}

// CreateRideRequest currently returns a static/mock ride response.
// This endpoint is intentionally simple and will be replaced with real persistence and matching logic.
func (h *RideHandler) CreateRideRequest(c *gin.Context) {
	var req models.Ride
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.RiderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rider_id is required"})
		return
	}

	req.ID = uuid.New().String()
	req.Status = models.RideStatusRequested

	drivers, err := h.driverService.FindNearbyDrivers(c.Request.Context(), req.PickupLat, req.PickupLon, 2, 20)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query nearby drivers"})
			return
	}


	c.JSON(http.StatusOK, gin.H{
		"message": "static ride response",
		"ride": req,
		"count":   len(drivers),
		"drivers": drivers,
	})
}
