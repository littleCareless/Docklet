package scanner

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const (
	DefaultLabelPrefix = "docklet."
	DefaultHost        = "localhost"
)

// GetEnvOrDefault gets an environment variable or returns a default value.
func GetEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// NewScanner creates a new Docker scanner instance (client).
func NewScanner() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return cli, nil
}

// ListServices scans for running Docker containers and extracts service information.
func ListServices(cli *client.Client) ([]ServiceInfo, error) {
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var services []ServiceInfo
	hostIP := GetEnvOrDefault("DOCKLET_HOST_IP", DefaultHost)

	for _, cont := range containers {
		// Default name is the first container name, cleaned up
		serviceName := strings.TrimPrefix(cont.Names[0], "/")

		// Extract info from labels
		title := cont.Labels[DefaultLabelPrefix+"title"]
		if title == "" {
			title = serviceName // Fallback to service name if no specific title
		}
		icon := cont.Labels[DefaultLabelPrefix+"icon"]
		description := cont.Labels[DefaultLabelPrefix+"description"]
		category := cont.Labels[DefaultLabelPrefix+"category"]
		order := cont.Labels[DefaultLabelPrefix+"order"] // Keep as string for now
		customURL := cont.Labels[DefaultLabelPrefix+"url"]

		var serviceURL string
		var portsInfo []string

		if customURL != "" {
			serviceURL = customURL
		} else if len(cont.Ports) > 0 {
			// Try to find the primary port or the first exposed one
			// This logic can be quite complex depending on how "primary" is defined.
			// For now, we take the first public port.
			var lowestPublicPort uint16 = 0
			var chosenPort uint16 = 0 // container port

			for _, p := range cont.Ports {
				if p.PublicPort > 0 {
					portsInfo = append(portsInfo, fmt.Sprintf("%s:%d->%d/%s", p.IP, p.PublicPort, p.PrivatePort, p.Type))
					if lowestPublicPort == 0 || p.PublicPort < lowestPublicPort {
						lowestPublicPort = p.PublicPort
						chosenPort = p.PrivatePort
					}
				} else {
					// For ports without host mapping, just list them
					portsInfo = append(portsInfo, fmt.Sprintf("%d/%s (no host port)", p.PrivatePort, p.Type))
				}
			}

			if lowestPublicPort > 0 {
				// Check for a label specifying which internal port to use for the URL
				// e.g., docklet.port=8080
				urlPortLabel := cont.Labels[DefaultLabelPrefix+"port"]
				if urlPortLabel != "" {
					if labelInternalPort, err := strconv.ParseUint(urlPortLabel, 10, 16); err == nil {
						// Find the corresponding public port for this labeled internal port
						found := false
						for _, p := range cont.Ports {
							if p.PrivatePort == uint16(labelInternalPort) && p.PublicPort > 0 {
								serviceURL = fmt.Sprintf("http://%s:%d", hostIP, p.PublicPort)
								found = true
								break
							}
						}
						if !found {
							log.Printf("Warning: Container %s specified docklet.port %s, but no corresponding host port mapping found. Using first available.", serviceName, urlPortLabel)
						}
					} else {
						log.Printf("Warning: Container %s has invalid docklet.port label '%s'. Ignoring.", serviceName, urlPortLabel)
					}
				}
				// If serviceURL is still not set (no valid docklet.port label or it wasn't found)
				if serviceURL == "" {
					serviceURL = fmt.Sprintf("http://%s:%d", hostIP, lowestPublicPort)
				}

			} else if len(cont.Ports) > 0 && chosenPort > 0 {
				// Fallback if no public port, but we have a private port (less useful for direct access)
				// This case might indicate a service not directly exposed or using host networking.
				// For host networking, the port is directly on the hostIP.
				// We'd need more sophisticated network mode detection for perfect accuracy.
				// For now, if docklet.port is specified, use that with hostIP.
				urlPortLabel := cont.Labels[DefaultLabelPrefix+"port"]
				if urlPortLabel != "" {
					if labelInternalPort, err := strconv.ParseUint(urlPortLabel, 10, 16); err == nil {
						serviceURL = fmt.Sprintf("http://%s:%d", hostIP, labelInternalPort)
					}
				} else {
					// If no docklet.port and no public mapping, the URL is uncertain.
					// We could try to infer if it's host networking, but that's more complex.
					// For now, leave it blank or provide a hint.
					log.Printf("Container %s (%s) has exposed ports but no direct host mapping and no docklet.port label. URL cannot be automatically determined.", serviceName, cont.ID)
				}
			}
		}

		// If still no URL, check for docklet.url_override
		urlOverride := cont.Labels[DefaultLabelPrefix+"url_override"]
		if urlOverride != "" {
			serviceURL = urlOverride
		}


		// Only include services with a valid HTTP/HTTPS URL
		if serviceURL == "" || (!strings.HasPrefix(serviceURL, "http://") && !strings.HasPrefix(serviceURL, "https://")) {
			log.Printf("Skipping container %s (%s) as it does not have a valid HTTP/HTTPS URL: '%s'", serviceName, cont.ID, serviceURL)
			continue // Skip to the next container
		}

		var networkNames []string
		if cont.NetworkSettings != nil && cont.NetworkSettings.Networks != nil {
			for name := range cont.NetworkSettings.Networks {
				networkNames = append(networkNames, name)
			}
		}

		services = append(services, ServiceInfo{
			ID:            cont.ID,
			Name:          serviceName, // User-friendly name, might be same as title initially
			Title:         title,       // Explicit title from label
			Icon:          icon,
			URL:           serviceURL,
			Description:   description,
			Category:      category,
			Order:         order,
			RawLabels:     cont.Labels,
			ContainerName: strings.TrimPrefix(cont.Names[0], "/"), // Keep original for reference
			Ports:         portsInfo,
			Networks:      networkNames,
			ImageName:     cont.Image,
			Status:        cont.State, // e.g. "running", "exited"
		})
	}

	return services, nil
}