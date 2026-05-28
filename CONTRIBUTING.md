# Contributing to Nova

## Welcome

This document outlines the conventions and guidelines for contributing to Nova.

## Getting Started

### Prerequisites

- VS Code with Dev Containers extension
- Docker Desktop running

### Setup

1. Clone the repository
2. Open in VS Code
3. Click "Reopen in Container" when prompted
4. Wait for container to build (first time: ~5 minutes)
5. Run `make db:migrate` to setup the database
6. Run `make dev` to start development servers

## Branch Naming

Format: `<type>/<ticket>-<short-description>`

| Type | Use case |
|------|----------|
| `feat/` | New feature |
| `fix/` | Bug fix |
| `chore/` | Maintenance, dependencies |
| `docs/` | Documentation only |
| `refactor/` | Code restructure |
| `test/` | Tests only |

Examples:
- `feat/001-add-user-crud`
- `fix/042-grid-pagination-bug`
- `chore/003-upgrade-angular-18`

## Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation |
| `style` | Formatting, no code change |
| `refactor` | Code restructure |
| `test` | Adding tests |
| `chore` | Build, deps, config |

### Examples

```
feat(users): add CRUD operations for users
fix(grid): resolve pagination reset on filter
docs(api): add endpoint documentation
```

## Code Style

- 2 spaces for indentation
- Single quotes for strings
- Semicolons required
- Max line length: 100 chars
- Run `make lint` before committing
- Auto-fix: `make lint:fix`

## Pull Request Process

1. Create branch from `main`
2. Make changes following these guidelines
3. Ensure all tests pass: `make test`
4. Ensure linting passes: `make lint`
5. Update documentation if needed
6. Open PR with description following template
7. Request review from team

## API Conventions

### REST Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/{entity}` | List all (with pagination) |
| GET | `/api/v1/{entity}/:id` | Get one by ID |
| POST | `/api/v1/{entity}` | Create new |
| PUT | `/api/v1/{entity}/:id` | Update existing |
| DELETE | `/api/v1/{entity}/:id` | Delete |

### Response Format

```json
{
  "success": true,
  "data": { },
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 100
  }
}
```

### Error Format

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human readable message",
    "details": []
  }
}
```

## Testing

- Backend: Go testing (built-in)
- Frontend: Jest + Testing Library
- Run tests: `make test`

## Documentation

- Update `ARCHITECTURE.md` for architectural decisions
- Add ADRs in `docs/decisions/` for significant changes
- Keep JSDoc comments on public methods

## Questions?

Open an issue or reach out to the team.
