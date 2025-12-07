# Protocol Buffers Setup Guide

## Issue

The protobuf generated files (`gophkeeper.pb.go` and `gophkeeper_grpc.pb.go`) need to be properly generated using the `protoc` compiler.

## Solution

### Step 1: Install protoc-gen-go and protoc-gen-go-grpc

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Step 2: Add to PATH

Make sure `$GOPATH/bin` is in your PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Step 3: Generate protobuf files

```bash
# From project root
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/api/proto/gophkeeper.proto
```

### Step 4: Build the project

```bash
make build
# or
go build -o bin/server cmd/server/main.go
go build -o bin/client cmd/client/main.go
```

## Alternative: Use Docker

If you have issues with protoc, you can use Docker to generate the files:

```bash
docker run --rm -v $(pwd):/workspace -w /workspace \
    namely/protoc-all:latest \
    -f pkg/api/proto/gophkeeper.proto \
    -l go \
    -o .
```

## Verifying Installation

```bash
# Check if protoc-gen-go is in PATH
which protoc-gen-go
# Should output: /Users/[username]/go/bin/protoc-gen-go

# Check if protoc-gen-go-grpc is in PATH
which protoc-gen-go-grpc
# Should output: /Users/[username]/go/bin/protoc-gen-go-grpc

# Test protoc
protoc --version
# Should output: libprotoc 3.x.x or higher
```

## Common Issues

### Issue: "protoc-gen-go: program not found"
**Solution**: Add `$(go env GOPATH)/bin` to your PATH

### Issue: "Cannot find import"
**Solution**: Run `go mod tidy` before building

### Issue: Module not found
**Solution**: Ensure you're running commands from the project root directory

## After Generation

Once the protobuf files are generated successfully, you can:

1. Build the project: `make build`
2. Run tests: `make test`
3. Start the server: `./bin/server`
4. Use the client: `./bin/client --help`

## Notes

The proto file is located at: `pkg/api/proto/gophkeeper.proto`

This file defines all the gRPC services and message types for the GophKeeper API.

