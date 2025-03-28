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
   git clone https://github.com/devurandom11/goping.git
   cd goping
   ```

2. Build the application:

   ```
   go build -o goping.exe
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
goping [options] <target1> <target2> <target3> ...
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
goping 192.168.1.1 192.168.1.2 192.168.1.3
```

Ping from a file containing a list of hostnames or IPs:

```
goping -f targets.txt
```

Ping a range of IPs:

```
goping -g 192.168.1.1-192.168.1.10
```

Ping a CIDR network:

```
goping -g 192.168.1.0/24
```

Send 5 pings to each target:

```
goping -c 5 192.168.1.1 192.168.1.2
```

Show only alive hosts:

```
goping -a 192.168.1.0/24
```

Show summary statistics:

```
goping -s 192.168.1.1 192.168.1.2 192.168.1.3
```

## Known Limitations

- Requires administrator privileges on Windows
- Only supports IPv4 (IPv6 support may be added in the future)
- Some advanced features from the original fping are not implemented

## License

This project is released under the MIT License.
