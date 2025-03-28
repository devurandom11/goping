package ping

import (
	"os/exec"
	"strings"
)

// IsAdmin checks if the application is running with administrator privileges
func IsAdmin() bool {
	cmd := exec.Command("net", "session")
	output, err := cmd.CombinedOutput()
	
	// Check if the command succeeded
	if err == nil {
		return true
	}
	
	// Check the output for access denied message
	outputStr := strings.ToLower(string(output))
	if strings.Contains(outputStr, "access is denied") {
		return false
	}
	
	// Default to assuming we don't have admin rights if we're not sure
	return false
} 