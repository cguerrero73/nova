# Nova Architecture

## Overview

Nova is an internal framework for building enterprise applications with a consistent pattern: master-detail views with grids, tabs, and related entities.

## Tech Stack

| Layer | Technology | Version |
|-------|------------|---------|
| Frontend | Angular | 17+ |
| State | Angular Signals | - |
| Styling | Tailwind CSS | 3.x |
| Tables | TanStack Table | 8.x |
| Backend | Fastify | 4.x |
| ORM | Prisma | 5.x |
| Database | PostgreSQL | 16 |
| Validation | Zod | - |

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
├── backend/                    # Fastify API
│   ├── src/
│   │   ├── modules/           # Feature modules
│   │   │   └── [entity]/
│   │   │       ├── routes.ts
│   │   │       ├── service.ts
│   │   │       └── schema.ts  # Zod schemas
│   │   ├── plugins/
│   │   ├── utils/
│   │   └── app.ts
│   └── prisma/
│       ├── schema.prisma
│       └── migrations/
│
├── docs/
│   ├── decisions/            # Architecture Decision Records
│   └── assets/
│
└── .devcontainer/           # Dev environment definition
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

| Component | Purpose |
|-----------|---------|
| `NovaGrid` | Generic data grid with sorting, filtering, pagination |
| `NovaForm` | Dynamic form from schema |
| `NovaTabs` | Tab container for detail views |
| `NovaSubgrid` | Grid for related entities |
| `NovaToolbar` | Action buttons (create, edit, delete, export) |

### API Structure

Each entity module follows this pattern:

```
backend/src/modules/[entity]/
├── routes.ts      # Fastify route definitions
├── service.ts     # Business logic
├── schema.ts      # Zod validation schemas
└── types.ts       # TypeScript interfaces
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | - |
| `NODE_ENV` | Environment: development, production | development |
| `PORT` | API server port | 4000 |
| `FRONTEND_URL` | Frontend URL for CORS | http://localhost:3000 |

## Development Commands

```bash
make dev          # Start all services
make db:migrate   # Run database migrations
make db:studio    # Open Prisma Studio
make test         # Run all tests
make lint         # Run linting
```

## Further Reading

- [Setup Guide for Windows/WSL](docs/setup/windows-wsl.md)
- [Setup Guide for macOS/Linux](docs/setup/macos-linux.md)
- [ADR Index](decisions/README.md)
- [Frontend Guidelines](docs/frontend-guidelines.md)
- [Backend Guidelines](docs/backend-guidelines.md)
