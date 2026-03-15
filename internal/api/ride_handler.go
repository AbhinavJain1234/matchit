package api

import (
	"net/http"

	"github.com/AbhinavJain1234/matchit/internal/service"
	"github.com/gin-gonic/gin"
)

// RideHandler owns ride-related HTTP handlers.
type RideHandler struct {
	rideService *service.RideService
}

func NewRideHandler(rideService *service.RideService) *RideHandler {
	return &RideHandler{
		rideService: rideService,
	}
}

type createRideRequest struct {
	RiderID   string  `json:"rider_id" binding:"required"`
	PickupLat float64 `json:"pickup_lat" binding:"required"`
	PickupLon float64 `json:"pickup_lon" binding:"required"`
	DestLat   float64 `json:"dest_lat" binding:"required"`
	DestLon   float64 `json:"dest_lon" binding:"required"`
}

// CreateRideRequest handles POST /ride/request.
func (h *RideHandler) CreateRideRequest(c *gin.Context) {
	var req createRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.rideService.CreateRideRequest(c.Request.Context(), service.CreateRideRequest{
		RiderID:   req.RiderID,
		PickupLat: req.PickupLat,
		PickupLon: req.PickupLon,
		DestLat:   req.DestLat,
		DestLon:   req.DestLon,
	})
	if err != nil {
		if err == service.ErrInvalidRiderID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create ride"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "ride created",
		"ride":    result.Ride,
		"drivers": result.Drivers,
	})
}

// GetRideStatus handles GET /ride/:id/status
// Returns the current state of a ride so the rider can poll it.
func (h *RideHandler) GetRideStatus(c *gin.Context) {
	rideID := c.Param("id")
	if rideID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ride id is required"})
		return
	}

	ride, err := h.rideService.GetRideByID(c.Request.Context(), rideID)
	if err != nil {
		if err == service.ErrRideNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "ride not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ride"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ride": ride})
}
