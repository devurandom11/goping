package target

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// ReadFromFile reads targets from a specified file, one per line
func ReadFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return readLines(file)
}

// ReadFromStdin reads targets from standard input, one per line
func ReadFromStdin() ([]string, error) {
	// Check if there's data available on stdin
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	// If there's no data from pipe or redirect, return empty slice
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return []string{}, nil
	}

	return readLines(os.Stdin)
}

// readLines reads lines from any io.Reader and returns non-empty lines
func readLines(r io.Reader) ([]string, error) {
	var targets []string
	scanner := bufio.NewScanner(r)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			targets = append(targets, line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return targets, nil
}

// GenerateFromRange generates a list of IP addresses from a start and end IP
func GenerateFromRange(startIP, endIP string) ([]string, error) {
	start := net.ParseIP(startIP)
	if start == nil {
		return nil, fmt.Errorf("invalid start IP: %s", startIP)
	}
	
	end := net.ParseIP(endIP)
	if end == nil {
		return nil, fmt.Errorf("invalid end IP: %s", endIP)
	}
	
	// We only support IPv4 for now
	start = start.To4()
	end = end.To4()
	
	if start == nil || end == nil {
		return nil, fmt.Errorf("only IPv4 addresses are supported")
	}
	
	// Compare start and end IPs
	if !lessThanOrEqual(start, end) {
		return nil, fmt.Errorf("start IP must be less than or equal to end IP")
	}
	
	var ips []string
	for ip := cloneIP(start); lessThanOrEqual(ip, end); incrementIP(ip) {
		ips = append(ips, ip.String())
	}
	
	return ips, nil
}

// GenerateFromCIDR generates a list of IP addresses from a CIDR notation
func GenerateFromCIDR(cidr string) ([]string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	
	// Get the first IP in the range
	ip := ipNet.IP.To4()
	if ip == nil {
		return nil, fmt.Errorf("only IPv4 CIDR notation is supported")
	}
	
	// Make a copy of the IP
	start := cloneIP(ip)
	
	// Calculate the last IP in the range
	mask := ipNet.Mask
	end := cloneIP(ip)
	for i := 0; i < len(end); i++ {
		end[i] |= ^mask[i]
	}
	
	// Generate IPs
	var ips []string
	for ip := start; lessThanOrEqual(ip, end); incrementIP(ip) {
		// Skip network and broadcast addresses for /31 and larger
		if mask[3] < 255 { // Not /32
			// Check if it's the network address (first address)
			if ip.Equal(start) {
				incrementIP(ip)
				continue
			}
			
			// Check if it's the broadcast address (last address)
			nextIP := cloneIP(ip)
			incrementIP(nextIP)
			if nextIP.Equal(net.IPv4(end[0], end[1], end[2], end[3])) {
				break
			}
		}
		
		ips = append(ips, ip.String())
	}
	
	return ips, nil
}

// Helper functions for IP manipulation

// cloneIP creates a copy of an IP
func cloneIP(ip net.IP) net.IP {
	clone := make(net.IP, len(ip))
	copy(clone, ip)
	return clone
}

// lessThanOrEqual checks if a is less than or equal to b
func lessThanOrEqual(a, b net.IP) bool {
	for i := 0; i < len(a); i++ {
		if a[i] < b[i] {
			return true
		} else if a[i] > b[i] {
			return false
		}
	}
	return true
}

// incrementIP increments an IP address by 1
func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
} 