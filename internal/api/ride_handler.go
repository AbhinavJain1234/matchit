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
		switch err {
			case service.ErrInvalidRiderID:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			case service.ErrActiveRideExists:
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			case service.ErrNoDriversAvailable:
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
				return
			case service.ErrRideCancelled:
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create ride"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "ride created",
		"ride":    result.Ride,
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

type acceptRideRequest struct {
	RideID   string `json:"ride_id" binding:"required"`
	DriverID string `json:"driver_id" binding:"required"`
}

type cancelRideRequest struct {
	RideID  string `json:"ride_id" binding:"required"`
	RiderID string `json:"rider_id" binding:"required"`
}

// AcceptRide handles POST /ride/accept.
// A driver calls this to claim a ride. Only the first driver to call succeeds.
// If two drivers call simultaneously, only one gets the ride — the other receives 409 Conflict.
func (h *RideHandler) AcceptRide(c *gin.Context) {
	var req acceptRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	ride, err := h.rideService.AcceptRide(c.Request.Context(), req.RideID, req.DriverID)
	if err != nil {
		switch err {
		case service.ErrRideNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "ride not found"})
		case service.ErrAlreadyAssigned:
			// 409 Conflict — another driver already accepted this ride
			c.JSON(http.StatusConflict, gin.H{"error": "ride already accepted by another driver"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "	failed to accept ride"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ride accepted",
		"ride":    ride,
	})
}

func (h *RideHandler) CancelRideRequest(c *gin.Context) {
	var req cancelRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.rideService.CancelRideRequest(c.Request.Context(), req.RideID, req.RiderID)
	if err != nil {
		switch err {
		case service.ErrRideNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "ride not found"})
		case service.ErrRideNotCancelable:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel ride"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ride cancelled",
	})
}

