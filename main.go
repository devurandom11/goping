package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/windows-fping/goging/ping"
	"github.com/windows-fping/goging/target"
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
	generateRange := flag.Bool("g", false, "Generate targets from IP range or CIDR notation")

	flag.Parse()

	if *aliveOnly && *unreachableOnly {
		fmt.Println("Error: Cannot use both -a and -u options simultaneously")
		os.Exit(1)
	}

	var targets []string
	var err error

	// Handle target input
	if *generateRange {
		args := flag.Args()
		if len(args) < 1 {
			fmt.Println("Error: -g requires IP range arguments")
			os.Exit(1)
		}

		if strings.Contains(args[0], "/") {
			// CIDR notation
			targets, err = target.GenerateFromCIDR(args[0])
		} else if len(args) >= 2 {
			// IP range
			targets, err = target.GenerateFromRange(args[0], args[1])
		} else {
			fmt.Println("Error: -g requires either CIDR notation or start/end IP addresses")
			os.Exit(1)
		}

		if err != nil {
			fmt.Printf("Error generating targets: %v\n", err)
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
			fmt.Println("Usage: goging [options] <target1> <target2> ...")
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