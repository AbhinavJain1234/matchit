package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// DriverHandler holds the dependencies needed by driver-related routes.
type DriverHandler struct {
	rdb *redis.Client
}

// NewDriverHandler creates a DriverHandler with the given Redis client.
func NewDriverHandler(rdb *redis.Client) *DriverHandler {
	return &DriverHandler{rdb: rdb}
}

// updateLocationRequest is the expected JSON body for POST /driver/location.
type updateLocationRequest struct {
	DriverID  string  `json:"driver_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// UpdateLocation handles POST /driver/location.
// It stores the driver position in the Redis GEO index so nearby-search works later.
func (h *DriverHandler) UpdateLocation(c *gin.Context) {
	var req updateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.rdb.GeoAdd(context.Background(), "drivers", &redis.GeoLocation{
		Name:      req.DriverID,
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	}).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "location updated"})
}













































