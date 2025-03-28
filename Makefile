.PHONY: build clean test

# Build binary
build:
	go build -o goping.exe

# Clean build artifacts
clean:
	if exist goping.exe del goping.exe

# Run tests
test:
	go test ./...

# Build for release (optimized)
release:
	go build -ldflags="-s -w" -o goping.exe

# Default target
all: build 