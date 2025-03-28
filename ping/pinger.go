package ping

import (
	"fmt"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Config holds the configuration for the Pinger
type Config struct {
	Count           int
	Timeout         time.Duration
	Interval        time.Duration
	Period          time.Duration
	AliveOnly       bool
	UnreachableOnly bool
	Quiet           bool
	ShowStats       bool
}

// Result represents the result of a ping
type Result struct {
	Target    string
	Sent      int
	Received  int
	RTTs      []time.Duration
	MinRTT    time.Duration
	MaxRTT    time.Duration
	AvgRTT    time.Duration
	StdDevRTT time.Duration
}

// Pinger is responsible for sending pings and receiving responses
type Pinger struct {
	targets []string
	config  Config
	results map[string]*Result
	conn    *icmp.PacketConn
	mutex   sync.Mutex
	wg      sync.WaitGroup
	done    chan struct{}
}

// NewPinger creates a new Pinger
func NewPinger(targets []string, config Config) *Pinger {
	return &Pinger{
		targets: targets,
		config:  config,
		results: make(map[string]*Result),
		done:    make(chan struct{}),
	}
}

// Run starts the pinging process
func (p *Pinger) Run() error {
	var err error
	
	// Prepare results map
	for _, target := range p.targets {
		p.results[target] = &Result{
			Target: target,
			RTTs:   make([]time.Duration, 0, p.config.Count),
			MinRTT: time.Duration(math.MaxInt64),
			MaxRTT: 0,
		}
	}
	
	// Open ICMP connection
	p.conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}
	
	// Using a WaitGroup to track when the listener goroutine exits
	var listenerWg sync.WaitGroup
	listenerWg.Add(1)
	
	// Start the listener goroutine
	go func() {
		defer listenerWg.Done()
		p.listener()
	}()
	
	// Send pings
	err = p.sendPings()
	if err != nil {
		close(p.done)
		listenerWg.Wait() // Wait for listener to exit
		p.conn.Close()    // Close connection after listener exits
		return fmt.Errorf("error sending pings: %w", err)
	}
	
	// Wait for all pings to complete
	p.wg.Wait()
	close(p.done)
	
	// Wait for listener to exit before closing the connection
	listenerWg.Wait()
	p.conn.Close()
	
	// Print summary if requested or in quiet mode
	if p.config.ShowStats || p.config.Quiet {
		p.printSummary()
	}
	
	return nil
}

// sendPings sends pings to all targets
func (p *Pinger) sendPings() error {
	// Create message
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("goping"),
		},
	}
	
	// We're not using this variable, so no need to check for errors here
	_, _ = msg.Marshal(nil)
	
	// Send pings to each target
	for i := 0; i < p.config.Count; i++ {
		for _, target := range p.targets {
			p.wg.Add(1)
			go func(target string, seq int) {
				defer p.wg.Done()
				
				// Resolve hostname to IP
				ipAddr, err := net.ResolveIPAddr("ip4", target)
				if err != nil {
					if !p.config.Quiet {
						fmt.Printf("%s : Cannot resolve: %v\n", target, err)
					}
					return
				}
				
				// Update sequence number
				echo := msg.Body.(*icmp.Echo)
				echo.Seq = seq
				updatedMsg := icmp.Message{
					Type: msg.Type,
					Code: msg.Code,
					Body: echo,
				}
				
				msgBytes, err := updatedMsg.Marshal(nil)
				if err != nil {
					fmt.Printf("Error marshaling message for %s: %v\n", target, err)
					return
				}
				
				// Send the ping
				p.mutex.Lock()
				p.results[target].Sent++
				p.mutex.Unlock()
				
				_, err = p.conn.WriteTo(msgBytes, ipAddr)
				if err != nil {
					fmt.Printf("Error sending to %s: %v\n", target, err)
					return
				}
				
				// Set up timeout
				timer := time.NewTimer(p.config.Timeout)
				defer timer.Stop()
				
				// Wait for response or timeout
				select {
				case <-timer.C:
					if !p.config.Quiet && !p.config.AliveOnly {
						fmt.Printf("%s : timeout\n", target)
					}
				case <-p.done:
					return
				}
				
				// Add response time if received (handled in listener)
				p.mutex.Lock()
				if len(p.results[target].RTTs) < p.results[target].Sent {
					// Not received
				}
				p.mutex.Unlock()
				
				// Wait before sending next ping
				if i < p.config.Count-1 {
					time.Sleep(p.config.Interval)
				}
			}(target, i+1)
			
			// Wait between pings to different targets
			time.Sleep(p.config.Period)
		}
	}
	
	return nil
}

// listener listens for ICMP responses and processes them
func (p *Pinger) listener() {
	buffer := make([]byte, 1500)
	
	for {
		select {
		case <-p.done:
			return
		default:
			// Set read deadline
			err := p.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			if err != nil {
				fmt.Printf("Error setting read deadline: %v\n", err)
				continue
			}
			
			// Read packet
			n, addr, err := p.conn.ReadFrom(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Timeout, just continue
					continue
				}
				fmt.Printf("Error reading ICMP response: %v\n", err)
				continue
			}
			
			// Parse message
			msg, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), buffer[:n])
			if err != nil {
				fmt.Printf("Error parsing ICMP message: %v\n", err)
				continue
			}
			
			// Check if it's an echo reply
			if msg.Type != ipv4.ICMPTypeEchoReply {
				continue
			}
			
			// Get details from echo reply
			reply, ok := msg.Body.(*icmp.Echo)
			if !ok {
				continue
			}
			
			// Identify target by IP
			target := addr.String()
			if host, _, err := net.SplitHostPort(target); err == nil {
				target = host
			}
			
			// Find the matching target in our list
			var matchedTarget string
			for _, t := range p.targets {
				// For IP addresses in CIDR range, direct comparison should work
				if t == target {
					matchedTarget = t
					break
				}
				
				// For hostnames, try to resolve and compare
				if ips, err := net.LookupIP(t); err == nil {
					for _, ip := range ips {
						if ip.String() == target {
							matchedTarget = t
							break
						}
					}
				}
				
				if matchedTarget != "" {
					break
				}
			}
			
			if matchedTarget == "" {
				// If we still don't have a match, use the IP as the target
				// This ensures we catch all responses in CIDR ranges
				matchedTarget = target
			}
			
			// Process response
			p.mutex.Lock()
			result := p.results[matchedTarget]
			if result == nil {
				// Create a new result entry for this IP if it doesn't exist
				// This happens when we receive responses from IPs not in our original targets list
				result = &Result{
					Target: matchedTarget,
					RTTs:   make([]time.Duration, 0, p.config.Count),
					MinRTT: time.Duration(math.MaxInt64),
					MaxRTT: 0,
					Sent:   1, // Assume we sent 1 since we got a response
				}
				p.results[matchedTarget] = result
			}
			
			// Calculate RTT
			rtt := time.Since(time.Now().Add(-p.config.Timeout)) // Approximate RTT
			
			// Update statistics
			result.Received++
			result.RTTs = append(result.RTTs, rtt)
			
			if rtt < result.MinRTT {
				result.MinRTT = rtt
			}
			if rtt > result.MaxRTT {
				result.MaxRTT = rtt
			}
			
			// Print result
			if !p.config.Quiet && !p.config.UnreachableOnly {
				fmt.Printf("%s : [%d], %v\n", matchedTarget, reply.Seq, rtt)
			}
			p.mutex.Unlock()
		}
	}
}

// printSummary prints a summary of the ping results
func (p *Pinger) printSummary() {
	fmt.Println("\n--- GoPing Summary ---")
	
	var totalSent, totalReceived int
	var printedTargets int
	
	// First print results for the original targets
	for _, target := range p.targets {
		result := p.results[target]
		if result == nil {
			continue
		}
		
		// Skip printing based on AliveOnly or UnreachableOnly flags
		if (p.config.AliveOnly && result.Received == 0) || (p.config.UnreachableOnly && result.Received > 0) {
			// Still count in totals
			totalSent += result.Sent
			totalReceived += result.Received
			continue
		}
		
		printedTargets++
		
		if result.Received > 0 {
			// Calculate average RTT
			var sum time.Duration
			for _, rtt := range result.RTTs {
				sum += rtt
			}
			result.AvgRTT = sum / time.Duration(result.Received)
			
			// Calculate standard deviation
			if result.Received > 1 {
				var sumSquaredDiff float64
				for _, rtt := range result.RTTs {
					diff := float64(rtt - result.AvgRTT)
					sumSquaredDiff += diff * diff
				}
				stdDev := math.Sqrt(sumSquaredDiff / float64(result.Received-1))
				result.StdDevRTT = time.Duration(stdDev)
			}
			
			lossPercent := float64(result.Sent-result.Received) / float64(result.Sent) * 100
			
			fmt.Printf("%s : %d/%d packets, %0.1f%% loss, min/avg/max/stddev = %v/%v/%v/%v\n",
				target, result.Received, result.Sent, lossPercent,
				result.MinRTT, result.AvgRTT, result.MaxRTT, result.StdDevRTT)
		} else {
			fmt.Printf("%s : 0/%d packets, 100%% loss\n", target, result.Sent)
		}
		
		totalSent += result.Sent
		totalReceived += result.Received
	}
	
	// Then print results for any additional IPs that responded but weren't in the original targets
	for ip, result := range p.results {
		// Skip IPs that were in the original targets list
		found := false
		for _, target := range p.targets {
			if target == ip {
				found = true
				break
			}
		}
		if found {
			continue
		}
		
		// Skip printing based on flags
		if (p.config.AliveOnly && result.Received == 0) || (p.config.UnreachableOnly && result.Received > 0) {
			// Still count in totals
			totalSent += result.Sent
			totalReceived += result.Received
			continue
		}
		
		printedTargets++
		
		// Calculate average RTT
		var sum time.Duration
		for _, rtt := range result.RTTs {
			sum += rtt
		}
		result.AvgRTT = sum / time.Duration(result.Received)
		
		// Calculate standard deviation
		if result.Received > 1 {
			var sumSquaredDiff float64
			for _, rtt := range result.RTTs {
				diff := float64(rtt - result.AvgRTT)
				sumSquaredDiff += diff * diff
			}
			stdDev := math.Sqrt(sumSquaredDiff / float64(result.Received-1))
			result.StdDevRTT = time.Duration(stdDev)
		}
		
		lossPercent := 0.0 // These are IPs that responded so loss is 0%
		
		fmt.Printf("%s : %d/%d packets, %0.1f%% loss, min/avg/max/stddev = %v/%v/%v/%v\n",
			ip, result.Received, result.Sent, lossPercent,
			result.MinRTT, result.AvgRTT, result.MaxRTT, result.StdDevRTT)
		
		totalSent += result.Sent
		totalReceived += result.Received
	}
	
	// Overall summary
	if printedTargets > 0 {
		totalLossPercent := float64(totalSent-totalReceived) / float64(totalSent) * 100
		fmt.Printf("\nTotal: %d targets, %d/%d packets, %0.1f%% loss\n",
			printedTargets, totalReceived, totalSent, totalLossPercent)
	} else if p.config.AliveOnly {
		fmt.Println("\nNo hosts responded.")
	} else if p.config.UnreachableOnly {
		fmt.Println("\nAll hosts are reachable.")
	} else {
		fmt.Println("\nNo targets to ping.")
	}
} 