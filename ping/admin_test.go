package ping

import (
	"testing"
)

func TestIsAdmin(t *testing.T) {
	// This test is minimal since it's system-dependent
	// Just ensure it runs without panic
	result := IsAdmin()
	
	// The result might be true or false depending on the privileges
	// Just ensure it's one of those values
	if result != true && result != false {
		t.Errorf("IsAdmin() returned invalid value: %v", result)
	}
} 