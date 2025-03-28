package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/windows-fping/goping/ping"
	"github.com/windows-fping/goping/target"
)

func main() {
	if runtime.GOOS != "windows" {
		fmt.Println("GoPing is designed specifically for Windows systems")
		os.Exit(1)
	}

	// Check for administrator privileges
	if !ping.IsAdmin() {
		fmt.Println("GoPing requires administrator privileges to send ICMP packets")
		fmt.Println("Please run this program as an administrator")
		os.Exit(1)
	}

	// Define flags/options
	count := flag.Int("c", 1, "Number of pings to send to each target")
	timeout := flag.Int("t", 500, "Timeout in milliseconds")
	interval := flag.Int("i", 1000, "Interval in milliseconds between pings to the same target")
	period := flag.Int("p", 25, "Period in milliseconds between pings to consecutive targets")
	aliveOnly := flag.Bool("a", false, "Show only alive hosts")
	unreachableOnly := flag.Bool("u", false, "Show only unreachable hosts")
	quiet := flag.Bool("q", false, "Quiet mode - only show summary")
	showStats := flag.Bool("s", false, "Show summary statistics")
	inputFile := flag.String("f", "", "Read targets from a file")
	cidrOrRange := flag.String("g", "", "Generate targets from IP range (start-end) or CIDR notation (x.x.x.x/y)")

	flag.Parse()

	if *aliveOnly && *unreachableOnly {
		fmt.Println("Error: Cannot use both -a and -u options simultaneously")
		os.Exit(1)
	}

	var targets []string
	var err error

	// Handle target input
	if *cidrOrRange != "" {
		// CIDR notation
		if strings.Contains(*cidrOrRange, "/") {
			targets, err = target.GenerateFromCIDR(*cidrOrRange)
			if err != nil {
				fmt.Printf("Error generating targets from CIDR %s: %v\n", *cidrOrRange, err)
				os.Exit(1)
			}
		} else if strings.Contains(*cidrOrRange, "-") {
			// IP range with dash notation (e.g., 192.168.1.1-192.168.1.10)
			parts := strings.Split(*cidrOrRange, "-")
			if len(parts) != 2 {
				fmt.Println("Error: Invalid IP range format. Use format: start-end (e.g., 192.168.1.1-192.168.1.10)")
				os.Exit(1)
			}
			targets, err = target.GenerateFromRange(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			if err != nil {
				fmt.Printf("Error generating targets from range %s: %v\n", *cidrOrRange, err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Error: -g requires either CIDR notation (x.x.x.x/y) or IP range (start-end)")
			os.Exit(1)
		}
	} else if *inputFile != "" {
		// Read targets from file
		targets, err = target.ReadFromFile(*inputFile)
		if err != nil {
			fmt.Printf("Error reading target file: %v\n", err)
			os.Exit(1)
		}
	} else if len(flag.Args()) > 0 {
		// Use command line arguments as targets
		targets = flag.Args()
	} else {
		// Check if there's data on stdin
		stdinTargets, err := target.ReadFromStdin()
		if err != nil {
			fmt.Printf("Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
		
		if len(stdinTargets) > 0 {
			targets = stdinTargets
		} else {
			fmt.Println("Error: No targets specified")
			fmt.Println("Usage: goping [options] <target1> <target2> ...")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	// Configure pinger
	pingerConfig := ping.Config{
		Count:           *count,
		Timeout:         time.Duration(*timeout) * time.Millisecond,
		Interval:        time.Duration(*interval) * time.Millisecond,
		Period:          time.Duration(*period) * time.Millisecond,
		AliveOnly:       *aliveOnly,
		UnreachableOnly: *unreachableOnly,
		Quiet:           *quiet,
		ShowStats:       *showStats,
	}

	// Run the pinger
	pinger := ping.NewPinger(targets, pingerConfig)
	err = pinger.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
} 