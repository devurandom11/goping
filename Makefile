.PHONY: build clean test

# Build binary
build:
	go build -o goging.exe

# Clean build artifacts
clean:
	if exist goging.exe del goging.exe

# Run tests
test:
	go test ./...

# Build for release (optimized)
release:
	go build -ldflags="-s -w" -o goging.exe

# Default target
all: build 