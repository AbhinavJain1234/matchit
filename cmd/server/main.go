package main

import (
	"net/http"

	"github.com/AbhinavJain1234/matchit/internal/api"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	// --- dependencies ---
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// --- handlers ---
	driverHandler := api.NewDriverHandler(rdb)

	// --- router ---
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "matchit",
			"status":  "ok",
		})
	})

	r.POST("/driver/location", driverHandler.UpdateLocation)

	r.Run(":8080")
}