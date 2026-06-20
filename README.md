# hexagonal-payments

Payment processing service built with Go following Hexagonal Architecture (Ports and Adapters).

## Architecture

```
internal/
├── domain/       # Core business logic — no external imports
├── application/  # Use cases / application services
├── ports/        # Interface definitions (driven + driving ports)
└── adapters/     # Concrete implementations (DB, HTTP, messaging)
```

The domain package has zero external dependencies. All infrastructure concerns live in adapters and implement port interfaces.

## Stack

- Go 1.22
- PostgreSQL (pgx driver)
- Chi router
- testify

## Running

```bash
go mod download
cp .env.example .env
go run ./cmd/server
```

## Tests

```bash
go test ./...
```
