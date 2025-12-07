# GophKeeper - Secure Password Manager

GophKeeper is a client-server password manager system that allows users to securely store and synchronize sensitive data across multiple devices.

## Features

- **Secure Storage**: Store login/password pairs, text notes, binary data, and bank card information
- **End-to-End Encryption**: All data is encrypted using AES-256-GCM
- **Multi-Device Sync**: Synchronize data across multiple authorized clients
- **Metadata Support**: Add custom metadata to any stored item
- **Cross-Platform CLI**: Works on Windows, Linux, and macOS
- **gRPC Communication**: Efficient binary protocol for client-server communication
- **JWT Authentication**: Secure token-based authentication

## Architecture

### Server Components
- **Authentication Service**: User registration and authentication
- **Storage Service**: Encrypted data storage with PostgreSQL
- **Sync Service**: Multi-device data synchronization

### Client Components
- **CLI Interface**: Command-line interface using Cobra
- **Local Cache**: Local encrypted data cache
- **Sync Manager**: Automatic synchronization with server

## Installation

### Prerequisites
- Go 1.24+
- PostgreSQL 14+
- Protocol Buffers compiler (protoc)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/ar11/gophkeeper.git
cd gophkeeper

# Install dependencies
go mod download

# Generate protobuf files (requires protoc and plugins)
# See SETUP_PROTOBUF.md for detailed instructions
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/api/proto/gophkeeper.proto

# Build server and client
make build

# Or build for all platforms
make build-all
```

**Note**: If you encounter issues with protobuf generation, see `SETUP_PROTOBUF.md` for detailed setup instructions.

## Usage

### Server

1. Start PostgreSQL database
2. Set environment variables:
```bash
export DB_DSN="postgres://user:password@localhost:5432/gophkeeper?sslmode=disable"
export JWT_SECRET="your-secret-key"
export SERVER_ADDRESS=":50051"
```

3. Run migrations:
```bash
make migrate-up
```

4. Start server:
```bash
./bin/server
```

### Client

#### Register a new user
```bash
./bin/client register --username user@example.com --password secretpass
```

#### Login
```bash
./bin/client login --username user@example.com --password secretpass
```

#### Store credentials
```bash
./bin/client add credential --name "GitHub" --login myuser --password mypass --metadata "website=github.com"
```

#### Store text data
```bash
./bin/client add text --name "Secret Note" --data "My secret information" --metadata "category=personal"
```

#### Store binary data
```bash
./bin/client add binary --name "SSH Key" --file ~/.ssh/id_rsa --metadata "server=production"
```

#### Store bank card
```bash
./bin/client add card --name "Visa Card" --number "4111111111111111" --holder "John Doe" --cvv "123" --expiry "12/25"
```

#### List all items
```bash
./bin/client list
```

#### Get item details
```bash
./bin/client get --name "GitHub"
```

#### Sync with server
```bash
./bin/client sync
```

#### Check version
```bash
./bin/client version
```

## Data Types

### Credentials
- Login/password pairs
- Associated metadata (website, notes, etc.)

### Text Data
- Arbitrary text information
- Secure notes
- Recovery codes

### Binary Data
- Files
- Encryption keys
- Certificates

### Bank Cards
- Card number
- Cardholder name
- CVV
- Expiration date

## Security

- All data is encrypted at rest using AES-256-GCM
- Transport layer security via gRPC with TLS
- Password hashing using bcrypt
- JWT tokens for authentication
- Master password never leaves the client

## Testing

Run unit tests:
```bash
make test
```

View coverage report:
```bash
make coverage
```

## Development

### Project Structure
```
gophkeeper/
├── cmd/
│   ├── server/          # Server entry point
│   └── client/          # Client entry point
├── internal/
│   ├── server/          # Server implementation
│   ├── client/          # Client implementation
│   ├── models/          # Data models
│   ├── storage/         # Storage layer
│   └── crypto/          # Encryption utilities
├── pkg/
│   ├── api/             # API definitions (protobuf)
│   └── auth/            # Authentication utilities
├── migrations/          # Database migrations
└── tests/              # Integration tests
```

## License

MIT License

## Contributing

Contributions are welcome! Please read CONTRIBUTING.md for details.

