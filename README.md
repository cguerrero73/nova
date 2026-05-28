# Nova

Internal framework for rapid enterprise application development.

## Quick Start

### Prerequisites

- VS Code with Dev Containers extension
- Docker Desktop

### Setup

1. Open in VS Code
2. Click "Reopen in Container" when prompted
3. Run migrations: `make db:migrate`
4. Start dev servers: `make dev`

### Services

| Service | URL |
|---------|-----|
| Frontend | http://localhost:4200 |
| Backend API | http://localhost:4000 |

## Commands

```bash
make dev          # Start all services
make db:migrate   # Run database migrations
make db:seed      # Run database seeds
make test         # Run tests
make lint         # Lint code
make lint:fix     # Auto-fix linting issues
```

## Project Structure

```
nova/
├── frontend/      Angular 17 application
├── backend/       Go + Fiber API with pgx
├── docs/          Architecture & ADRs
└── .devcontainer/ Dev environment
```

## Tech Stack

- **Frontend**: Angular 17, Signals, Tailwind CSS
- **Backend**: Go, Fiber, pgx, PostgreSQL
- **Migrations**: golang-migrate
- **Patterns**: Schema-driven screens

See [ARCHITECTURE.md](ARCHITECTURE.md) for details.
