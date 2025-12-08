# GophKeeper Architecture

## System Overview

GophKeeper follows a client-server architecture with strong security principles:

```
┌─────────────────────────────────────────────────────────────┐
│                         Client Layer                        │
├─────────────────────────────────────────────────────────────┤
│  CLI Interface (Cobra)                                      │
│  ├── Commands: register, login, add, get, list, sync       │
│  └── Master Password Management                             │
├─────────────────────────────────────────────────────────────┤
│  Client Logic                                               │
│  ├── Local Config (~/.gophkeeper/config.json)              │
│  ├── Client-side Encryption (AES-256-GCM)                  │
│  └── gRPC Client Connection                                 │
└─────────────────────────────────────────────────────────────┘
                             ▼
                    [gRPC over TCP/TLS]
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                         Server Layer                        │
├─────────────────────────────────────────────────────────────┤
│  gRPC Server (port 50051)                                   │
│  └── Service Handlers                                       │
├─────────────────────────────────────────────────────────────┤
│  Business Logic                                             │
│  ├── Authentication (JWT)                                   │
│  ├── Authorization                                          │
│  └── Synchronization Logic                                  │
├─────────────────────────────────────────────────────────────┤
│  Storage Layer                                              │
│  └── PostgreSQL Interface                                   │
└─────────────────────────────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                     Database Layer                          │
├─────────────────────────────────────────────────────────────┤
│  PostgreSQL Database                                        │
│  ├── users table                                            │
│  └── data_items table                                       │
└─────────────────────────────────────────────────────────────┘
```

## Component Details

### Client Components

#### 1. CLI Interface (`cmd/client/`)
- **Technology**: Cobra framework
- **Responsibilities**:
  - Parse command-line arguments
  - Handle user input (including secure password prompts)
  - Format and display output
  - Manage version information

#### 2. Client Logic (`internal/client/`)
- **Responsibilities**:
  - Establish gRPC connection to server
  - Manage authentication tokens
  - Encrypt/decrypt data using master password
  - Synchronize data with server
  - Maintain local configuration

#### 3. Encryption (`internal/crypto/`)
- **Algorithms**:
  - AES-256-GCM for data encryption
  - PBKDF2 for key derivation (100,000 iterations)
  - bcrypt for password hashing
- **Features**:
  - Unique nonce for each encryption
  - Salt-based key derivation
  - Authenticated encryption (AEAD)

### Server Components

#### 1. gRPC Server (`cmd/server/`)
- **Technology**: Google gRPC
- **Configuration**:
  - Listening address (default :50051)
  - Database connection
  - JWT secret
  - Graceful shutdown handling

#### 2. Service Handlers (`internal/server/`)
- **Authentication Services**:
  - `Register`: Create new user account
  - `Login`: Authenticate and issue JWT token

- **Data Services**:
  - `AddItem`: Store encrypted data item
  - `GetItem`: Retrieve data item by ID or name
  - `ListItems`: List all user's items (with type filtering)
  - `UpdateItem`: Update existing item (with version control)
  - `DeleteItem`: Soft-delete item

- **Sync Services**:
  - `Sync`: Synchronize client and server state
  - Conflict detection and resolution

#### 3. Storage Layer (`internal/storage/`)
- **Interface**: Storage interface for flexibility
- **Implementation**: PostgreSQL
- **Features**:
  - Connection pooling
  - Prepared statements
  - Transaction support
  - Soft deletes
  - Version tracking for optimistic locking

#### 4. Authentication (`pkg/auth/`)
- **Technology**: JWT (JSON Web Tokens)
- **Token Structure**:
  ```json
  {
    "user_id": "uuid",
    "username": "email",
    "exp": "timestamp",
    "iat": "timestamp",
    "nbf": "timestamp"
  }
  ```
- **Token Lifetime**: 24 hours
- **Validation**: HMAC-SHA256 signing

### Data Models (`internal/models/`)

#### User Model
```go
type User struct {
    ID           string
    Username     string
    PasswordHash string    // bcrypt
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

#### DataItem Model
```go
type DataItem struct {
    ID            string
    UserID        string
    Type          DataType  // credential, text, binary, card
    Name          string
    EncryptedData []byte    // AES-256-GCM encrypted
    Metadata      map[string]string
    CreatedAt     time.Time
    UpdatedAt     time.Time
    Version       int64     // For conflict resolution
    Deleted       bool      // Soft delete flag
}
```

## Data Flow

### Registration Flow
```
Client                          Server                     Database
  │                               │                            │
  ├─1. Enter username/password────────────────────────────────▶│
  │                               │                            │
  │                               ├─2. Hash password───────────▶│
  │                               │    (bcrypt)                │
  │                               │                            │
  │                               ├─3. Store user──────────────▶│
  │                               │                            │
  │                               ├─4. Generate JWT────────────▶│
  │                               │                            │
  │◀─5. Return token──────────────┤                            │
  │                               │                            │
  ├─6. Save token to config───────────────────────────────────▶│
```

### Add Item Flow
```
Client                          Server                     Database
  │                               │                            │
  ├─1. Get master password from user──────────────────────────▶│
  │                               │                            │
  ├─2. Encrypt data locally───────────────────────────────────▶│
  │    (AES-256-GCM)              │                            │
  │                               │                            │
  ├─3. Send encrypted data + JWT──────────────────────────────▶│
  │                               │                            │
  │                               ├─4. Validate JWT────────────▶│
  │                               │                            │
  │                               ├─5. Store encrypted data────▶│
  │                               │                            │
  │◀─6. Confirmation──────────────┤                            │
```

### Sync Flow
```
Client                          Server                     Database
  │                               │                            │
  ├─1. Send last sync timestamp───────────────────────────────▶│
  │                               │                            │
  │                               ├─2. Query items modified────▶│
  │                               │    after timestamp         │
  │                               │                            │
  │                               │◀─3. Return modified items──┤
  │                               │                            │
  │◀─4. Send items to client─────┤                            │
  │                               │                            │
  ├─5. Detect conflicts───────────────────────────────────────▶│
  │    (version comparison)       │                            │
  │                               │                            │
  ├─6. Update local state─────────────────────────────────────▶│
```

## Security Architecture

### Defense in Depth

1. **Client-Side Encryption**
   - Master password never leaves the client
   - Data encrypted before transmission
   - Key derivation with PBKDF2 (100k iterations)

2. **Transport Security**
   - gRPC with TLS support (configure in production)
   - Certificate validation
   - Secure cipher suites

3. **Server-Side Security**
   - JWT-based authentication
   - Token expiration (24 hours)
   - bcrypt password hashing (cost factor 10)
   - SQL injection prevention (prepared statements)

4. **Database Security**
   - Encrypted data at rest
   - User data isolation (foreign keys)
   - Soft deletes for data recovery

### Encryption Details

#### Data Encryption (AES-256-GCM)
```
Plaintext → [PBKDF2] → Key (256-bit)
                         ↓
Plaintext + Key → [AES-GCM] → Salt + Nonce + Ciphertext + Tag
```

#### Password Hashing (bcrypt)
```
Password → [bcrypt] → Hash (60 bytes)
                      ↓
                   Stored in DB
```

## API Protocol (gRPC)

### Message Format
- **Serialization**: Protocol Buffers (protobuf)
- **Transport**: HTTP/2
- **RPC Style**: Unary (request-response)

### API Endpoints

#### Authentication
- `Register(RegisterRequest) → RegisterResponse`
- `Login(LoginRequest) → LoginResponse`

#### Data Management
- `AddItem(AddItemRequest) → AddItemResponse`
- `GetItem(GetItemRequest) → GetItemResponse`
- `ListItems(ListItemsRequest) → ListItemsResponse`
- `UpdateItem(UpdateItemRequest) → UpdateItemResponse`
- `DeleteItem(DeleteItemRequest) → DeleteItemResponse`

#### Synchronization
- `Sync(SyncRequest) → SyncResponse`

## Scalability Considerations

### Current Design
- Single server instance
- Single database instance
- Suitable for: Small to medium deployments (< 10k users)

### Scaling Options

#### Horizontal Scaling
```
Load Balancer
    ├── Server Instance 1
    ├── Server Instance 2
    └── Server Instance N
            ↓
    PostgreSQL (Primary)
        ├── Read Replica 1
        └── Read Replica 2
```

#### Optimizations
- Connection pooling (already implemented)
- Read replicas for list/get operations
- Caching layer (Redis) for JWT validation
- CDN for client distribution

## Monitoring and Observability

### Logging
- Structured logging with contextual information
- Log levels: DEBUG, INFO, WARN, ERROR
- Sensitive data never logged

### Metrics (Future)
- Request count by endpoint
- Request latency percentiles
- Active connections
- Database query performance
- Authentication success/failure rate

### Health Checks (Future)
- `/health` endpoint
- Database connectivity check
- Disk space monitoring

## Deployment Architecture

### Docker Deployment
```
docker-compose.yml
    ├── postgres (database)
    └── server (application)
```

### Production Deployment (Recommended)
```
Internet
    ↓
[Reverse Proxy: nginx/Traefik]
    ↓
[GophKeeper Server] ← [PostgreSQL]
```

## Error Handling

### Client Errors
- Input validation
- Network error recovery
- Authentication failure handling
- Graceful degradation

### Server Errors
- Input sanitization
- Database error handling
- Rate limiting (future)
- Circuit breaker pattern (future)

## Future Architecture Enhancements

1. **Microservices Split**
   - Auth Service
   - Data Service
   - Sync Service

2. **Event Sourcing**
   - Audit log for all operations
   - Replay capability

3. **Real-time Sync**
   - WebSocket/gRPC streaming
   - Push notifications

4. **Multi-region**
   - Geographic data replication
   - Conflict-free replicated data types (CRDTs)

## Testing Strategy

### Unit Tests
- Coverage: 70%+
- Focus: crypto, auth, models
- Mocking: database, gRPC

### Integration Tests (Future)
- End-to-end flows
- Database operations
- gRPC communication

### Performance Tests (Future)
- Load testing
- Stress testing
- Benchmark comparisons

---

For implementation details, see the source code in respective directories.

