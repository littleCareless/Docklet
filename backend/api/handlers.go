package api

import (
	"log"
	"net/http"

	scanner "docklet/docker_scanner"

	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

// ServicesHandlerGin handles requests to list Docker services using Gin.
func ServicesHandlerGin(dockerCli *client.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		services, err := scanner.ListServices(dockerCli)
		if err != nil {
			log.Printf("Error listing services: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list Docker services"})
			return
		}
		// Allow all origins for simplicity in development
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, services)
	}
}

// HealthCheckHandlerGin provides a simple health check endpoint using Gin.
func HealthCheckHandlerGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}