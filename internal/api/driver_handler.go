package api

import (
	"net/http"

	"github.com/AbhinavJain1234/matchit/internal/service"
	"github.com/gin-gonic/gin"
)

// DriverHandler handles HTTP requests for driver-related routes.
// It knows nothing about Redis or business rules — it only speaks HTTP.
type DriverHandler struct {
	driverService *service.DriverService
}

func NewDriverHandler(driverService *service.DriverService) *DriverHandler {
	return &DriverHandler{driverService: driverService}
}

// updateLocationRequest is the expected JSON body for POST /driver/location.
type updateLocationRequest struct {
	DriverID  string  `json:"driver_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// UpdateLocation handles POST /driver/location.
func (h *DriverHandler) UpdateLocation(c *gin.Context) {
	var req updateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.driverService.UpdateLocation(c.Request.Context(), req.DriverID, req.Latitude, req.Longitude); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "location updated"})
}