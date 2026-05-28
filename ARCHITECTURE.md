# Nova Architecture

## Overview

Nova is an internal framework for building enterprise applications with a consistent pattern: master-detail views with grids, tabs, and related entities.

## Tech Stack

| Layer      | Technology      | Version |
| ---------- | --------------- | ------- |
| Frontend   | Angular         | 17+     |
| State      | Angular Signals | -       |
| Styling    | Tailwind CSS    | 3.x     |
| Tables     | TanStack Table  | 8.x     |
| Backend    | Go + Fiber      | 1.22    |
| DB Access  | pgx/v5          | 5.5     |
| Database   | PostgreSQL      | 16      |
| Migrations | golang-migrate  | 4.16    |
| Validation | Zod (shared)    | -       |

## Directory Structure

```
nova/
├── frontend/                    # Angular application
│   ├── src/
│   │   ├── app/
│   │   │   ├── core/           # Singleton services, guards, interceptors
│   │   │   │   ├── services/   # API communication
│   │   │   │   ├── guards/     # Route guards
│   │   │   │   └── interceptors/
│   │   │   ├── shared/         # Reusable components, directives, pipes
│   │   │   │   ├── components/
│   │   │   │   ├── directives/
│   │   │   │   └── pipes/
│   │   │   └── features/      # Feature modules (screen-specific)
│   │   │       └── [entity]/
│   │   │           ├── screens/
│   │   │           ├── services/
│   │   │           └── models/
│   │   ├── environments/
│   │   └── styles/
│   └── angular.json
│
├── backend/                    # Go + Fiber API (Hexagonal Architecture)
│   ├── cmd/                   # Binaries (server, migrate, setup, check)
│   ├── internal/
│   │   ├── domain/           # Core business logic (pure, no external deps)
│   │   │   └── [entity]/
│   │   │       ├── entity.go    # Domain entity
│   │   │       ├── service.go   # Business logic
│   │   │       ├── ports.go     # Repository interfaces (ports)
│   │   │       └── dto.go       # Request/Response DTOs
│   │   ├── adapters/          # Implementations of ports
│   │   │   ├── api/            # Driving adapters (HTTP handlers)
│   │   │   │   └── [entity]/
│   │   │   │       └── handler.go
│   │   │   └── db/             # Driven adapters (DB repositories)
│   │   │       └── [entity]/
│   │   │           └── *Repository.go
│   │   ├── infrastructure/    # External concerns
│   │   │   ├── config/        # Configuration
│   │   │   ├── db/           # Database connection
│   │   │   ├── middleware/    # Auth, tenant middleware
│   │   │   └── wire/         # Dependency injection
│   │   └── handler/          # Shared DTOs
│   ├── migrations/           # SQL migrations (golang-migrate)
│   │   ├── global/
│   │   └── tenant/
│   ├── go.mod / go.sum
│   └── config.yaml
│
├── docs/
│   ├── decisions/            # Architecture Decision Records
│   └── assets/
│
└── .devcontainer/           # Dev environment definition
```

## Hexagonal Architecture

The backend follows **Hexagonal Architecture** (aka Ports & Adapters):

```
┌─────────────────────────────────────────────────────────────────┐
│                        DRIVING ADAPTERS                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │   Auth   │  │  Users   │  │   Orgs   │  │  Parts   │   ...   │
│  │ Handler  │  │ Handler  │  │ Handler  │  │ Handler  │         │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘         │
└───────┼─────────────┼─────────────┼─────────────┼───────────────┘
        │             │             │             │
        ▼             ▼             ▼             ▼
┌─────────────────────────────────────────────────────────────────┐
│                         DOMAIN (Core)                            │
│  ┌────────────────────────────────────────────────────────┐     │
│  │  AuthService ←─ UserRepository, SessionRepository       │     │
│  │  UserService ←─ UserRepository                          │     │
│  │  OrgService  ←─ OrganizationRepository                 │     │
│  │  ...                                                    │     │
│  │                                                         │     │
│  │  [entity]/ports.go  → Interface definitions (ports)    │     │
│  │  [entity]/service.go → Business logic                   │     │
│  │  [entity]/dto.go    → Request/Response DTOs             │     │
│  └────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
        │             │             │             │
        ▼             ▼             ▼             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      DRIVEN ADAPTERS                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │   Auth   │  │  Users   │  │   Orgs   │  │  Parts   │   ...   │
│  │   Repo   │  │   Repo   │  │   Repo   │  │   Repo   │         │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘         │
│      ↓              ↓              ↓              ↓              │
│   PostgreSQL    PostgreSQL     PostgreSQL     PostgreSQL         │
└─────────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

| Layer              | Description                                                                                                |
| ------------------ | ---------------------------------------------------------------------------------------------------------- |
| **Domain**         | Pure business logic, zero external dependencies. Contains entities, services, ports (interfaces), and DTOs |
| **API Adapters**   | Drive the application (HTTP handlers). Transform HTTP requests to domain calls                             |
| **DB Adapters**    | Driven by the application (DB repositories). Implement repository ports                                    |
| **Infrastructure** | Cross-cutting concerns: config, DI (wire), middleware, database connection                                 |

### Port/Adapter Pattern per Domain

```
internal/domain/[entity]/
├── entity.go       → Domain entity (struct + business methods)
├── service.go      → Application service (orchestrates domain logic)
├── ports.go       → Repository interface(s) defined by the domain
└── dto.go          → Request/Response DTOs

internal/adapters/api/[entity]/
└── handler.go     → HTTP adapter (Fiber handler)

internal/adapters/db/[entity]/
└── [entity]Repository.go → PostgreSQL adapter implementing the port
```

## Core Patterns

### Screen Schema Pattern

Screens are defined declaratively, not built screen-by-screen:

```typescript
// Schema definition
export const UserListSchema: ScreenSchema = {
  entity: 'User',
  list: {
    columns: ['id', 'name', 'email', 'status'],
    filters: ['name', 'status'],
    sortable: true,
    paginated: true,
  },
};

export const UserDetailSchema: ScreenSchema = {
  entity: 'User',
  tabs: [
    { id: 'general', label: 'General', type: 'form' },
    { id: 'orders', label: 'Orders', type: 'grid', relatedEntity: 'Order' },
    { id: 'history', label: 'History', type: 'grid', relatedEntity: 'AuditLog' },
  ],
};
```

### Shared Components

| Component     | Purpose                                               |
| ------------- | ----------------------------------------------------- |
| `NovaGrid`    | Generic data grid with sorting, filtering, pagination |
| `NovaForm`    | Dynamic form from schema                              |
| `NovaTabs`    | Tab container for detail views                        |
| `NovaSubgrid` | Grid for related entities                             |
| `NovaToolbar` | Action buttons (create, edit, delete, export)         |

## Environment Variables

| Variable       | Description                          | Default     |
| -------------- | ------------------------------------ | ----------- |
| `DATABASE_URL` | PostgreSQL connection string         | -           |
| `PORT`         | API server port                      | 4000        |
| `JWT_SECRET`   | Secret for JWT signing               | -           |
| `NODE_ENV`     | Environment: development, production | development |

## Development Commands

```bash
make dev          # Start all services
make db:migrate   # Run database migrations (golang-migrate)
make db:seed      # Run database seeds
make test         # Run all tests
make lint         # Run linting
```

## Further Reading

- [Setup Guide for Windows/WSL](docs/setup/windows-wsl.md)
- [Setup Guide for macOS/Linux](docs/setup/macos-linux.md)
- [ADR Index](docs/decisions/README.md)
