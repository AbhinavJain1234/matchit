package api

import ( 
	"net/http"
	"strconv"

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

type DriverStatusChangeRequest struct {
	DriverID string `json:"driver_id" binding:"required"`
	Status   string `json:"status" binding:"required,oneof=available unavailable"`
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

// GetNearbyDrivers handles GET /drivers/nearby?lat=..&lon=..&radius_km=..&limit=..
func (h *DriverHandler) GetNearbyDrivers(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat"})
		return
	}

	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lon"})
		return
	}

	radiusKM := 2.0
	if c.Query("radius_km") != "" {
		radiusKM, err = strconv.ParseFloat(c.Query("radius_km"), 64)
		if err != nil || radiusKM <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid radius_km"})
			return
		}
	}

	limit := 20
	if c.Query("limit") != "" {
		parsed, parseErr := strconv.Atoi(c.Query("limit"))
		if parseErr != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		limit = parsed
	}

	drivers, err := h.driverService.FindNearbyDrivers(c.Request.Context(), lat, lon, radiusKM, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query nearby drivers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":   len(drivers),
		"drivers": drivers,
	})
}

// func (h *DriverHandler) ChangeStatus(c *gin.Context) {
// 	var req DriverStatusChangeRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
// 		return
// 	}

// 	if err := h.driverService.ChangeStatus(c.Request.Context(), req.DriverID, req.Status); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to change status"})
// 		return
// 	}
	
// 	c.JSON(http.StatusOK, gin.H{"status": "driver status updated"})
// }