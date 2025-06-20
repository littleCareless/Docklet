package scanner

// ServiceInfo represents a discovered Docker service.
// It will be serialized to JSON for the API.
type ServiceInfo struct {
	ID            string            `json:"id"`              // Container ID
	Name          string            `json:"name"`            // User-friendly name (from labels.title or container name)
	Title         string            `json:"title"`           // Explicit title from docklet.title, if different from Name
	Icon          string            `json:"icon"`            // Icon URL or class (from docklet.icon)
	URL           string            `json:"url"`             // Access URL (e.g., http://<host_ip_or_domain>:<port>)
	Description   string            `json:"description"`     // Service description (from docklet.description)
	Category      string            `json:"category"`        // Service category (from docklet.category)
	Order         string            `json:"order"`           // Service order hint (from docklet.order), string for now
	RawLabels     map[string]string `json:"raw_labels"`      // All labels from the container
	ContainerName string            `json:"container_name"`  // Original container name
	Ports         []string          `json:"ports"`           // Exposed ports info: "host_ip:host_port->container_port/protocol"
	Networks      []string          `json:"networks"`        // Networks the container is attached to
	ImageName     string            `json:"image_name"`      // Name of the image used by the container
	Status        string            `json:"status"`          // Container status
}

// ScannerConfig for the scanner, might include label prefixes, default host IP, etc.
// Not used actively in the current simplified version but good for future expansion.
type ScannerConfig struct {
	DockerHost      string // e.g., "unix:///var/run/docker.sock"
	DefaultHostIP   string // Default IP to use if not found in labels
	LabelPrefix     string // e.g., "docklet."
}