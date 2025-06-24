.PHONY: build test install clean run

# Build the binary
build:
	go build -o pool .

# Run tests
test:
	go test -v ./...

# Install the binary
install: build
	go install .

# Clean build artifacts
clean:
	rm -f pool
	go clean

# Run the application
run:
	go run .

# Build for all platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o pool-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o pool-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o pool-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o pool-windows-amd64.exe .

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...