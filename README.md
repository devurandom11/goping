# Project Specification: GoPing - An fping Implementation for Windows

## 1. Introduction

fping is a popular command-line utility used to send ICMP echo probes to network hosts, similar to ping, but optimized for pinging multiple hosts concurrently. It's widely used for network monitoring and diagnostics. Currently, a native, high-performance fping equivalent specifically built for modern Windows systems using a modern language like Go is less common.

This project aims to create GoPing, a command-line tool written in Go that replicates the core functionality of fping on the Windows platform. Leveraging Go's concurrency model and standard library aims to provide a performant and easy-to-deploy alternative.

## Implementation

GoPing is implemented as a Go application with the following features:

- Efficiently ping multiple target hosts concurrently using ICMP echo requests
- Accepts targets via command-line arguments, stdin, file, or IP range generation
- Measures and reports Round-Trip Times (RTT)
- Identifies and reports reachable and unreachable hosts
- Provides output formats and command-line options similar to fping

## Building

### Prerequisites

- Go 1.16 or later
- Windows 10/11 or Windows Server 2016 or later
- Git (optional, for cloning the repository)

### Build Instructions

1. Clone the repository or download the source code:

   ```
   git clone https://github.com/windows-fping/goging.git
   cd goging
   ```

2. Build the application:

   ```
   go build -o goging.exe
   ```

   Or use the provided Makefile:

   ```
   make build
   ```

3. For an optimized release build:
   ```
   make release
   ```

## Usage

### Running GoPing

GoPing requires administrator privileges to send and receive ICMP packets. Run it from an elevated command prompt or PowerShell.

Basic usage:

```
goging [options] <target1> <target2> <target3> ...
```

### Command-Line Options

- `-c <count>`: Number of pings to send to each target (default: 1)
- `-t <timeout>`: Timeout in milliseconds (default: 500)
- `-i <interval>`: Interval in milliseconds between pings to the same target (default: 1000)
- `-p <period>`: Period in milliseconds between pings to consecutive targets (default: 25)
- `-a`: Show only alive hosts
- `-u`: Show only unreachable hosts
- `-q`: Quiet mode - only show summary
- `-s`: Show summary statistics
- `-f <file>`: Read targets from a file
- `-g`: Generate targets from IP range or CIDR notation

### Examples

Ping multiple hosts:

```
goging 192.168.1.1 192.168.1.2 192.168.1.3
```

Ping from a file containing a list of hostnames or IPs:

```
goging -f targets.txt
```

Ping a range of IPs:

```
goging -g 192.168.1.1 192.168.1.10
```

Ping a CIDR network:

```
goging -g 192.168.1.0/24
```

Send 5 pings to each target:

```
goging -c 5 192.168.1.1 192.168.1.2
```

Show only alive hosts:

```
goging -a 192.168.1.0/24
```

Show summary statistics:

```
goging -s 192.168.1.1 192.168.1.2 192.168.1.3
```

## Known Limitations

- Requires administrator privileges on Windows
- Only supports IPv4 (IPv6 support may be added in the future)
- Some advanced features from the original fping are not implemented

## License

This project is released under the MIT License.

## 2. Goals

- Develop a CLI application for Windows that mimics the essential features of fping
- Efficiently ping multiple target hosts (provided via arguments, stdin, file, or range generation) concurrently using ICMP echo requests
- Accurately measure and report Round-Trip Times (RTT)
- Identify and report reachable and unreachable hosts
- Provide output formats and command-line options familiar to fping users
- Compile into a single, statically linked executable for easy distribution (where feasible)
- Ensure compatibility with modern Windows versions (Windows 10/11, Windows Server 2016 and later)

## 3. Non-Goals

- Replicating every obscure or less-used command-line option from the original fping. The focus is on the most common and useful features
- A graphical user interface (GUI). This is strictly a command-line tool
- Support for legacy Windows versions (e.g., Windows 7, XP)
- Advanced features not present in the core fping toolset (e.g., TCP pings, path discovery)

## 4. Functional Requirements

### Target Input:

- Accept a list of target hostnames or IP addresses as command-line arguments
- Read a list of targets from standard input (stdin)
- Read a list of targets from a specified file (-f <file>)
- Generate a list of targets from an IP address range (-g <start_ip> <end_ip>) or CIDR notation (-g <network/mask>)

### Pinging Mechanism:

- Send ICMPv4 Echo Request packets concurrently to all specified targets
- Listen for ICMPv4 Echo Reply packets
- Handle timeouts for hosts that do not respond within a specified duration (-t <timeout>)

### Output:

- Default Mode: Print status (alive/unreachable) and RTT (if alive) for each target as responses are received or timeouts occur
- Alive Mode (-a): Only display hosts that are reachable
- Unreachable Mode (-u): Only display hosts that are unreachable
- Quiet Mode (-q): Suppress per-probe/per-target results, only showing the final summary
- Summary Statistics (-s): At the end of execution, print statistics including the number of targets, packets sent, packets received, packet loss percentage, and minimum, average, and maximum RTT for responsive hosts

### Control Options:

- Count (-c <count>): Send a specific number of pings to each target. Default is typically 1 unless combined with interval/period options
- Interval (-i <interval>): Wait interval milliseconds between sending subsequent pings to the same target. Default: ~1000ms
- Period (-p <period>): Wait period milliseconds between sending pings to consecutive targets in the list/round-robin fashion. Default: ~25ms
- Timeout (-t <timeout>): Specify the time in milliseconds to wait for a reply for a single ping before considering it timed out. Default: ~500ms

## 5. Technical Requirements

### Language and Platform:

- Go (latest stable version)
- Windows 10, Windows 11, Windows Server 2016, Windows Server 2019, Windows Server 2022 (64-bit)
- Protocol: ICMPv4 (ICMPv6 support is a potential future enhancement)

### Privileges:

Sending/receiving raw ICMP packets on Windows requires Administrator privileges. The application must:

- Check if running with sufficient privileges
- If not, clearly inform the user that elevation is required
- Optionally, attempt to self-elevate via UAC prompt (though this can be complex and might be deferred)

### Dependencies:

- Minimize external dependencies
- Utilize the Go standard library (net, os, time, flag, etc.)
- Consider using golang.org/x/net/icmp and golang.org/x/net/ipv4 for ICMP packet construction and handling, as this simplifies dealing with raw sockets

### Concurrency:

Use Goroutines extensively for concurrent sending, listening, and timeout management. Channels should be used for communication between goroutines.

## 6. High-Level Architecture

### Components:

- **CLI Parser**: Parses command-line arguments and flags (using Go's flag package or a library like cobra/urfave/cli). Validates input
- **Target Manager**: Resolves hostnames (if necessary), generates targets from ranges/CIDR, and manages the list of targets to be pinged. Handles input from files or stdin

### ICMP Engine:

- **Listener**: Opens a raw ICMP socket (requires privileges), listens for incoming ICMP Echo Replies, parses them, and matches them to sent requests (using ID/Sequence numbers). Sends results (RTT, success/failure) back via channels
- **Sender**: Creates and sends ICMP Echo Request packets to targets based on timing parameters (-i, -p). Uses goroutines for concurrency. Tracks sent packets
- **Timeout Manager**: Manages timers for each sent packet/target, signaling timeouts via channels if no reply is received within the specified duration (-t)

### Results Processing:

- **Results Aggregator**: Receives results (success, failure, RTT, timeout) from the ICMP Engine via channels. Stores per-target statistics
- **Output Formatter**: Formats and prints results to stdout in real-time (unless in -q mode) based on the selected output mode (-a, -u)
- **Statistics Calculator**: Calculates the final summary statistics (-s) after all pings are complete or the process is interrupted

## 7. Implementation Considerations & Challenges

- **Windows Raw Sockets**: This is the primary technical hurdle. Requires careful handling of Windows APIs or leveraging appropriate Go packages (golang.org/x/net/...) to correctly open raw sockets and construct/parse ICMP packets. Ensure proper privilege checks are implemented
- **Packet Identification**: Reliably matching incoming replies to outgoing requests, especially with many concurrent pings. Use unique ICMP identifiers and sequence numbers
- **Concurrency Scaling**: Efficiently managing potentially thousands of goroutines and timers without excessive resource consumption or deadlocks. Use worker pools or bounded concurrency if necessary
- **Timer Accuracy**: Achieving reasonably accurate RTT measurements using Go's time package. Be mindful of system load and scheduling delays
- **Hostname Resolution**: Handle DNS resolution efficiently and potentially cache results for repeated targets

## 8. Testing Strategy

### Test Types:

- **Unit Tests**: Test individual components like argument parsing, target generation, IP range/CIDR logic, and statistics calculation in isolation. Mocking network interactions where possible
- **Integration Tests**: Test the core pinging loop by pinging localhost (127.0.0.1) or known local network IPs. Verify basic packet sending/receiving and RTT measurement. Test different command-line flag combinations
- **Manual Testing**: Execute the compiled binary on target Windows versions. Test against various real-world targets (local and remote), large target lists, files, and generated ranges. Test privilege checks and error handling. Compare output and behavior against the original fping on Linux where applicable

## 9. Deliverables

- Source code managed in a Git repository (e.g., GitHub)
- README.md detailing:
  - Project description
  - Installation/Build instructions
  - Usage examples for all supported options
  - Known limitations or differences from original fping
  - Privilege requirements
- Compiled GoPing.exe binary for Windows (amd64)
- A suite of automated tests (unit and potentially integration)

## 10. Future Considerations

- Add ICMPv6 support
- Implement timestamp (-D) or elapsed time (-e) options
- Add support for setting the source IP address (-S)
- Add TOS (Type of Service) / DSCP support (-O)
- Provide output options like CSV
- Cross-compilation support for Linux and macOS
