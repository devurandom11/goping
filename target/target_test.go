package target

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadLinesFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Single line",
			input:    "192.168.1.1",
			expected: []string{"192.168.1.1"},
		},
		{
			name:     "Multiple lines",
			input:    "192.168.1.1\n192.168.1.2\n192.168.1.3",
			expected: []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
		},
		{
			name:     "Lines with whitespace",
			input:    "  192.168.1.1  \n\t192.168.1.2\n  192.168.1.3  ",
			expected: []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
		},
		{
			name:     "Lines with comments",
			input:    "192.168.1.1\n# This is a comment\n192.168.1.2",
			expected: []string{"192.168.1.1", "192.168.1.2"},
		},
		{
			name:     "Lines with empty lines",
			input:    "192.168.1.1\n\n\n192.168.1.2",
			expected: []string{"192.168.1.1", "192.168.1.2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.input)
			got, err := readLines(reader)
			if err != nil {
				t.Fatalf("readLines() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("readLines() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestGenerateFromRange(t *testing.T) {
	tests := []struct {
		name     string
		startIP  string
		endIP    string
		expected []string
		wantErr  bool
	}{
		{
			name:     "Valid range",
			startIP:  "192.168.1.1",
			endIP:    "192.168.1.3",
			expected: []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
			wantErr:  false,
		},
		{
			name:     "Invalid start IP",
			startIP:  "invalid",
			endIP:    "192.168.1.3",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Invalid end IP",
			startIP:  "192.168.1.1",
			endIP:    "invalid",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "End IP before start IP",
			startIP:  "192.168.1.10",
			endIP:    "192.168.1.1",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GenerateFromRange(test.startIP, test.endIP)
			if (err != nil) != test.wantErr {
				t.Errorf("GenerateFromRange() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !test.wantErr && !reflect.DeepEqual(got, test.expected) {
				t.Errorf("GenerateFromRange() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestGenerateFromCIDR(t *testing.T) {
	tests := []struct {
		name     string
		cidr     string
		minLen   int
		maxLen   int
		wantErr  bool
	}{
		{
			name:     "Valid CIDR /24",
			cidr:     "192.168.1.0/24",
			minLen:   250, // We skip network and broadcast
			maxLen:   300,
			wantErr:  false,
		},
		{
			name:     "Valid CIDR /30",
			cidr:     "192.168.1.0/30",
			minLen:   2, // Only 2 usable IPs in a /30
			maxLen:   2,
			wantErr:  false,
		},
		{
			name:     "Invalid CIDR",
			cidr:     "invalid",
			minLen:   0,
			maxLen:   0,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GenerateFromCIDR(test.cidr)
			if (err != nil) != test.wantErr {
				t.Errorf("GenerateFromCIDR() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !test.wantErr {
				if len(got) < test.minLen || len(got) > test.maxLen {
					t.Errorf("GenerateFromCIDR() returned %d IPs, expected between %d and %d",
						len(got), test.minLen, test.maxLen)
				}
			}
		})
	}
} 