package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"docklet/api"
	dockerscanner "docklet/docker_scanner" // Renamed import for clarity
	systemscanner "docklet/system_scanner" // Added for system services

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

const (
	DefaultPort = "8888"
)

func main() {
	// Create a new Docker client
	dockerCli, err := dockerscanner.NewScanner()
	if err != nil {
		log.Fatalf("Failed to initialize Docker scanner: %v", err)
	}
	// defer dockerCli.Close() // Consider closing when app exits

	// Create a new System scanner
	sysScanner, err := systemscanner.NewSystemScanner()
	if err != nil {
		log.Fatalf("Failed to initialize System scanner: %v", err)
	}
	// defer sysScanner.Close() // Consider closing when app exits

	// Get port from environment or use default
	port := dockerscanner.GetEnvOrDefault("DOCKLET_PORT", DefaultPort)
	listenAddr := ":" + port

	// Initialize Gin router
	router := gin.Default()

	// API routes
	apiRoutes := router.Group("/api")
	{
		apiRoutes.GET("/services", api.ServicesHandlerGin(dockerCli)) // Docker services
		apiRoutes.GET("/system-services", api.SystemServicesHandlerGin(sysScanner)) // Native system services
		apiRoutes.GET("/health", api.HealthCheckHandlerGin())
	}

	// Serve static files for the frontend
	// Vue/React apps usually build to a 'dist' folder.
	// We'll serve from './frontend/dist' assuming the Vue app is in './frontend'
	// and its build output is in 'dist'.
	frontendDistPath := "./frontend/dist"

	// Check if the frontend build directory exists
	if _, err := os.Stat(frontendDistPath); !os.IsNotExist(err) {
		// Serve static files from frontend/dist
		router.Use(static.Serve("/", static.LocalFile(frontendDistPath, false)))
		// If you want to ensure that non-API routes also serve index.html for SPA routing:
		router.NoRoute(func(c *gin.Context) {
			// Check if it's not an API call and not a file that exists in static
			if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
				// Check if the file exists in the static directory
                // If not, serve index.html
                filePath := filepath.Join(frontendDistPath, c.Request.URL.Path)
                if _, err := os.Stat(filePath); os.IsNotExist(err) {
                    c.File(filepath.Join(frontendDistPath, "index.html"))
                    return
                }
			}
			// Default 404 if it's an API route not found or an existing file not found by static.Serve
			// c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
		})
		log.Printf("Serving Vue/React frontend from %s on /", frontendDistPath)
	} else {
		// Fallback if no frontend build is found (e.g., during backend-only development)
		// This also means the old './static' directory is no longer primary.
		// We can remove the old static serving logic or keep it as a deeper fallback.
		// For now, let's assume the new frontend will be the primary.
		log.Printf("No frontend build directory found at %s. API will be available, but no UI.", frontendDistPath)
		router.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "Docklet Gin API is running. No frontend UI found at %s.", frontendDistPath)
		})
	}

	log.Printf("Docklet Gin server starting on %s", listenAddr)
	log.Printf("Docker Services API: http://%s%s/api/services", dockerscanner.GetEnvOrDefault("DOCKLET_HOST_IP", dockerscanner.DefaultHost), listenAddr)
	log.Printf("System Services API: http://%s%s/api/system-services", dockerscanner.GetEnvOrDefault("DOCKLET_HOST_IP", dockerscanner.DefaultHost), listenAddr)
	log.Printf("Health check: http://%s%s/api/health", dockerscanner.GetEnvOrDefault("DOCKLET_HOST_IP", dockerscanner.DefaultHost), listenAddr)

	if err := router.Run(listenAddr); err != nil {
		log.Fatalf("Failed to start Gin server: %v", err)
	}
}