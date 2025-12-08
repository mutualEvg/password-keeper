# Contributing to GophKeeper

Thank you for your interest in contributing to GophKeeper! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 14 or higher
- Protocol Buffers compiler (protoc)
- Docker and Docker Compose (optional)

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/ar11/gophkeeper.git
cd gophkeeper
```

2. Install dependencies:
```bash
go mod download
```

3. Generate protobuf files:
```bash
make proto
```

4. Start PostgreSQL (using Docker):
```bash
docker-compose up -d postgres
```

5. Run database migrations:
```bash
make migrate-up
```

6. Build the project:
```bash
make build
```

## Code Guidelines

### Code Style

- Follow standard Go code style and conventions
- Use `gofmt` to format your code
- Run `go vet` to check for common errors
- Use meaningful variable and function names

### Documentation

- All exported functions, types, and variables must have documentation comments
- Documentation should start with the name of the thing being documented
- Use complete sentences in documentation

Example:
```go
// GenerateToken creates a JWT token for the specified user.
// It returns the signed token string or an error if generation fails.
func GenerateToken(userID, username, secret string) (string, error) {
    // ...
}
```

### Testing

- Write unit tests for all new functionality
- Aim for at least 70% code coverage
- Use table-driven tests where appropriate
- Include both positive and negative test cases

Run tests:
```bash
make test
```

Check coverage:
```bash
make coverage
```

### Commit Messages

Follow the conventional commits format:

- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation changes
- `test:` adding or updating tests
- `refactor:` code refactoring
- `chore:` maintenance tasks

Example:
```
feat: add support for OTP data type
fix: resolve race condition in sync handler
docs: update API documentation for new endpoints
```

## Pull Request Process

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Make your changes
4. Write or update tests
5. Ensure all tests pass: `make test`
6. Update documentation if needed
7. Commit your changes with a meaningful commit message
8. Push to your fork: `git push origin feature/your-feature-name`
9. Create a Pull Request

### PR Requirements

- All tests must pass
- Code coverage should not decrease
- Code must be formatted with `gofmt`
- No linter warnings
- Include description of changes
- Reference any related issues

## Project Structure

```
gophkeeper/
├── cmd/                    # Application entry points
│   ├── server/            # Server binary
│   └── client/            # Client binary
├── internal/              # Private application code
│   ├── server/           # Server implementation
│   ├── client/           # Client implementation
│   ├── models/           # Data models
│   ├── storage/          # Storage layer
│   └── crypto/           # Encryption utilities
├── pkg/                   # Public packages
│   ├── api/              # API definitions (protobuf)
│   └── auth/             # Authentication utilities
├── migrations/            # Database migrations
└── tests/                # Integration tests
```

## Questions or Issues?

If you have questions or encounter issues:

1. Check existing issues on GitHub
2. Read the documentation
3. Ask in discussions
4. Create a new issue with detailed information

## License

By contributing to GophKeeper, you agree that your contributions will be licensed under the MIT License.

