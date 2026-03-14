package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type DriverLocationRequest struct {
	DriverID  string  `json:"driver_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

func main() {

	r := gin.Default()

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "server running",
		})
	})

	// Driver location endpoint
	r.POST("/driver/location", func(c *gin.Context) {

		var req DriverLocationRequest

		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		// Store location in Redis GEO index
		err := rdb.GeoAdd(ctx, "drivers", &redis.GeoLocation{
			Name:      req.DriverID,
			Longitude: req.Longitude,
			Latitude:  req.Latitude,
		}).Err()

		if err != nil {
			c.JSON(500, gin.H{"error": "failed to store location"})
			return
		}

		fmt.Println("Driver location stored:", req.DriverID)

		c.JSON(200, gin.H{
			"status": "location stored",
		})
	})

	r.Run(":8080")
}