package api

import (
	"log"
	"net/http"

	dockerscanner "docklet/docker_scanner" // Renamed to avoid conflict
	systemscanner "docklet/system_scanner"

	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

// ServicesHandlerGin handles requests to list Docker services using Gin.
func ServicesHandlerGin(dockerCli *client.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		services, err := dockerscanner.ListServices(dockerCli)
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

// SystemServicesHandlerGin handles requests to list native system services using Gin.
func SystemServicesHandlerGin(sysScanner *systemscanner.SystemScanner) gin.HandlerFunc {
	return func(c *gin.Context) {
		allServices, err := sysScanner.ListServices()
		if err != nil {
			log.Printf("Error listing system services: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list system services"})
			return
		}

		// Filter for likely web services
		webServices := []systemscanner.SystemServiceInfo{}
		for _, service := range allServices {
			if service.IsLikelyWebService {
				webServices = append(webServices, service)
			}
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, webServices)
	}
}

// HealthCheckHandlerGin provides a simple health check endpoint using Gin.
func HealthCheckHandlerGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}