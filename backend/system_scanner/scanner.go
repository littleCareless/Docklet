package systemscanner

import (
	"bytes"
	"fmt"
	"log" // Added for logging errors during lsof
	"os/exec"
	"regexp" // Added for parsing lsof output
	"runtime"
	"strconv" // Added for port conversion
	"strings"
)

var commonWebPorts = map[int]bool{
	80:   true,
	443:  true,
	3000: true, // Common for Node.js dev servers
	3001: true, // Common for React dev servers (sometimes)
	5000: true, // Common for Flask dev servers
	5173: true, // Common for Vite dev servers
	8000: true, // Common for Python dev servers, Django
	8080: true, // Common for Java app servers, other dev servers
	8888: true, // Common for Jupyter, other dev servers
}

func isCommonWebPort(portStr string) bool {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false
	}
	return commonWebPorts[port]
}

// SystemScanner provides methods to scan for native system services.
type SystemScanner struct {
	// Potentially add configuration options here if needed in the future
}

// NewSystemScanner creates a new SystemScanner.
func NewSystemScanner() (*SystemScanner, error) {
	return &SystemScanner{}, nil
}

// ListServices lists all detectable native system services.
// It routes to the appropriate OS-specific implementation.
func (s *SystemScanner) ListServices() ([]SystemServiceInfo, error) {
	switch runtime.GOOS {
	case "darwin":
		return s.listMacServices()
	case "linux":
		return s.listLinuxServices()
	case "windows":
		return s.listWindowsServices()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// listMacServices lists services on macOS using launchctl.
func (s *SystemScanner) listMacServices() ([]SystemServiceInfo, error) {
	var services []SystemServiceInfo

	// List all services known to launchd
	cmd := exec.Command("launchctl", "list")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// launchctl list can return non-zero exit code if some services are in a bad state,
		// but still output useful information. We'll log the error but try to parse.
		// However, if there's no output, it's a more serious issue.
		if out.Len() == 0 {
			return nil, fmt.Errorf("failed to execute launchctl list: %v, output: %s", err, out.String())
		}
		// Log the error but continue: log.Printf("launchctl list returned error (continuing parsing): %v", err)
	}

	lines := strings.Split(out.String(), "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "PID") { // Skip header line
		lines = lines[1:]
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			pidField := fields[0]
			statusField := fields[1] // This is an exit code for jobs, or "-" for daemons/agents not running a process
			label := fields[2]

			// Determine a more user-friendly status
			currentStatus := "unknown"
			isRunning := pidField != "-" && pidField != "0"

			if isRunning {
				currentStatus = "running"
			} else if statusField == "0" {
				currentStatus = "loaded" // Or "stopped" if it's not meant to be persistent
			} else if statusField != "-" {
				currentStatus = fmt.Sprintf("exited(status: %s)", statusField)
			} else {
				currentStatus = "stopped/unloaded"
			}

			var listeningPorts []string
			isLikelyWebService := false

			if isRunning {
				ports, err := getListeningTCPPorts(pidField)
				if err != nil {
					log.Printf("Notice: Failed to get listening ports for PID %s (%s): %v. This might be due to permissions or the process terminating.", pidField, label, err)
				} else {
					listeningPorts = ports
					if len(ports) > 0 {
						for _, p := range ports {
							if isCommonWebPort(p) {
								isLikelyWebService = true
								break
							}
						}
					}
				}
			}

			services = append(services, SystemServiceInfo{
				Name:               label,
				DisplayName:        label,
				Status:             currentStatus,
				PID:                pidField,
				ListeningPorts:     listeningPorts,
				IsLikelyWebService: isLikelyWebService,
			})
		}
	}

	return services, nil
}
// getListeningTCPPorts uses lsof to find TCP ports a given PID is listening on.
// Returns a list of port numbers as strings.
func getListeningTCPPorts(pidStr string) ([]string, error) {
	if pidStr == "-" || pidStr == "0" {
		return nil, nil // Not a running process
	}

	cmd := exec.Command("lsof", "-p", pidStr, "-iTCP", "-sTCP:LISTEN", "-P", "-n")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// lsof returns 1 if no files are found (i.e., no listening ports for the PID). This is not an error for us.
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			if strings.Contains(stderr.String(), "Can't be stat(2)ed") || strings.Contains(stderr.String(), "no such process") {
				// Process might have terminated between launchctl list and lsof
				return []string{}, nil
			}
			if out.Len() == 0 { // If output is empty and exit code is 1, it means no listening ports.
				return []string{}, nil
			}
		}
		// For other errors, or if lsof found something but still exited with an error.
		return nil, fmt.Errorf("lsof command failed for PID %s: %v, stderr: %s, stdout: %s", pidStr, err, stderr.String(), out.String())
	}

	var ports []string
	// Regex to find port numbers like *:80 or 127.0.0.1:8080 or [::1]:80
	// It looks for content like `*:port`, `host:port`, or `[ipv6]:port`
	re := regexp.MustCompile(`\S+:(\d+)\s+\(LISTEN\)`)
	lines := strings.Split(out.String(), "\n")

	for _, line := range lines {
		// Skip header or empty lines
		if strings.HasPrefix(line, "COMMAND") || strings.TrimSpace(line) == "" {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			// Ensure port is not already added (lsof can list IPv4 and IPv6 separately for the same port)
			found := false
			for _, p := range ports {
				if p == matches[1] {
					found = true
					break
				}
			}
			if !found {
				ports = append(ports, matches[1])
			}
		}
	}
	return ports, nil
}

// listLinuxServices lists services on Linux.
// This is a placeholder and needs implementation (e.g., using systemctl, service, or /etc/init.d).
func (s *SystemScanner) listLinuxServices() ([]SystemServiceInfo, error) {
	services := []SystemServiceInfo{}
	// Example: using systemctl (requires systemctl to be installed and user to have permissions)
	/*
	   cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager")
	   var out bytes.Buffer
	   cmd.Stdout = &out
	   err := cmd.Run()
	   if err != nil {
	       return nil, fmt.Errorf("failed to execute systemctl: %v", err)
	   }

	   lines := strings.Split(out.String(), "\n")
	   for _, line := range lines {
	       // Parse systemctl output
	       // UNIT          LOAD   ACTIVE SUB     DESCRIPTION
	       // service1.service loaded active running Service 1 Description
	       // This parsing logic can be complex.
	       if strings.Contains(line, ".service") {
	           fields := strings.Fields(line)
	           if len(fields) >= 4 {
	               serviceName := fields[0]
	               status := fields[2] // active, inactive, failed etc.
	               description := strings.Join(fields[4:], " ")
	               services = append(services, SystemServiceInfo{
	                   Name:        serviceName,
	                   DisplayName: description,
	                   Status:      status,
	               })
	           }
	       }
	   }
	*/
	// For now, return a dummy service for Linux
	services = append(services, SystemServiceInfo{
		Name:        "dummy-linux-service",
		DisplayName: "Dummy Linux Service",
		Status:      "running",
		Description: "This is a placeholder for Linux service detection.",
	IsLikelyWebService: true, // Make it show up for testing
	ListeningPorts: []string{"8080"},
})
return services, nil
}

// listWindowsServices lists services on Windows.
// This is a placeholder and needs implementation (e.g., using `sc query` or WMI).
func (s *SystemScanner) listWindowsServices() ([]SystemServiceInfo, error) {
	services := []SystemServiceInfo{}
	// Example: using sc query (requires appropriate permissions)
	/*
	   cmd := exec.Command("sc", "query", "state=", "all", "type=", "service")
	   var out bytes.Buffer
	   cmd.Stdout = &out
	   err := cmd.Run()
	   if err != nil {
	       return nil, fmt.Errorf("failed to execute sc query: %v", err)
	   }
	   // Parsing 'sc query' output is non-trivial.
	   // It typically looks like:
	   // SERVICE_NAME: service1
	   // DISPLAY_NAME: Service 1
	   //         TYPE               : 10  WIN32_OWN_PROCESS
	   //         STATE              : 4  RUNNING
	   //                                (STOPPABLE, NOT_PAUSABLE, IGNORES_SHUTDOWN)
	   //         WIN32_EXIT_CODE    : 0  (0x0)
	   //         SERVICE_EXIT_CODE  : 0  (0x0)
	   //         CHECKPOINT         : 0x0
	   //         WAIT_HINT          : 0x0
	*/
	// For now, return a dummy service for Windows
	services = append(services, SystemServiceInfo{
		Name:        "dummy-windows-service",
		DisplayName: "Dummy Windows Service",
		Status:      "running",
		Description: "This is a placeholder for Windows service detection.",
	IsLikelyWebService: true, // Make it show up for testing
	ListeningPorts: []string{"80"},
})
return services, nil
}

// Close cleans up any resources used by the SystemScanner.
// Currently, no resources are held that need explicit cleanup.
func (s *SystemScanner) Close() error {
	// No-op for now
	return nil
}