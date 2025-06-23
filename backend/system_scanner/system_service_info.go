package systemscanner

// SystemServiceInfo holds information about a native system service.
type SystemServiceInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"` // Often more user-friendly
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`       // e.g., running, stopped, paused
	StartType   string `json:"start_type,omitempty"`   // e.g., auto, manual, disabled
	PathName    string `json:"path_name,omitempty"`    // Path to the service executable
	PID         string `json:"pid,omitempty"` // Process ID, if running (string for flexibility with "-")
	ListeningPorts []string `json:"listening_ports,omitempty"` // Ports the service is listening on
	IsLikelyWebService bool `json:"is_likely_web_service,omitempty"` // True if it's likely a web service
}