# GophKeeper Makefile

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X 'main.Version=$(VERSION)' -X 'main.BuildDate=$(BUILD_DATE)'

.PHONY: all build clean test proto server client

all: proto build

# Generate protobuf files
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/api/proto/gophkeeper.proto

# Build both server and client
build: server client

# Build server
server:
	go build -ldflags "$(LDFLAGS)" -o bin/server cmd/server/main.go

# Build client
client:
	go build -ldflags "$(LDFLAGS)" -o bin/client cmd/client/main.go

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/client-linux-amd64 cmd/client/main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/client-darwin-amd64 cmd/client/main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/client-darwin-arm64 cmd/client/main.go
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/client-windows-amd64.exe cmd/client/main.go

# Run tests
test:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Show test coverage
coverage: test
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Run server
run-server:
	./bin/server

# Run client
run-client:
	./bin/client

# Database migrations up
migrate-up:
	migrate -path migrations -database "postgres://localhost:5432/gophkeeper?sslmode=disable" up

# Database migrations down
migrate-down:
	migrate -path migrations -database "postgres://localhost:5432/gophkeeper?sslmode=disable" down

