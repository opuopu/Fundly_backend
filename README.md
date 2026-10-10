# Fundly

Fundly is a financial application built with Go following **Domain-Driven Design (DDD)** and Clean / Hexagonal Architecture principles.

---

## 🏛️ Architecture Overview

The codebase is structured into clear architectural boundaries adhering to DDD principles.
* 📖 **REST API Specification:** Detailed API design and contract specification is available in [`docs/rest-api-design.md`](docs/rest-api-design.md).

```text
fundly/
├── cmd/
│   └── fundly/
│       └── main.go                  # Composition root (bootstraps server, wires dependencies)
├── internal/
│   ├── config/
│   │   └── config.go                # Application environment configuration
│   ├── domain/                      # Enterprise business logic & core rules (pure Go, 0 external deps)
│   │   ├── user/                    # User aggregate, repository interface, domain errors
│   │   ├── account/                 # Account aggregate, balance logic, repository interface
│   │   ├── transaction/             # Transaction entity, statuses, repository interface
│   │   ├── terms/                   # Terms and conditions domain model
│   │   └── stat/                    # Analytics and reporting stats domain models
│   ├── application/                 # Use case orchestration
│   │   └── transfer/                # Cross-account transfer application service
│   └── infrastructure/              # Adapters and external implementations
│       ├── http/                    # Primary adapter: HTTP handlers & routing
│       └── persistence/
│           └── postgres/            # Secondary adapter: PostgreSQL repository implementations
├── migrations/                      # Database schema migration files
├── go.mod
└── README.md
```

---

## 🔄 Dependency Rules

1. **Domain (`internal/domain`)**:
   - Contains pure business entities, value objects, and repository **interfaces**.
   - Has **no dependencies** on outer layers (no SQL drivers, no HTTP frameworks).
2. **Application (`internal/application`)**:
   - Orchestrates domain models and repository interfaces to fulfill use cases (e.g. money transfers).
   - Depends only on the domain layer.
3. **Infrastructure (`internal/infrastructure`)**:
   - Implements interfaces defined in the domain layer (e.g., PostgreSQL persistence).
   - Handles transport protocols (HTTP handlers, gRPC, etc.).
4. **Command / Entrypoint (`cmd/fundly`)**:
   - Acts as the composition root. Loads configs, establishes database pools, and injects concrete infrastructure into application services.

---

## 🚀 Getting Started

### Run Tests
```bash
go test -v ./...
```

### Build Project
```bash
go build ./...
```

### Start Server
```bash
go run ./cmd/fundly
```

The health check endpoint is accessible at:
```bash
curl http://localhost:8080/health
```
