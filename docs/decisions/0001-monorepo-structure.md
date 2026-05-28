# ADR-0001: Monorepo Structure with Workspace

## Status

Accepted

## Context

Nova will be used to build multiple applications internally. We need:
- Shared code that can be reused across applications
- Consistent tooling and configuration
- Easy onboarding for new developers
- Independent deployment of frontend and backend

## Decision

We will use npm workspaces in a single repository with this structure:

```
nova/
├── frontend/        # Angular application
├── backend/         # Go + Fiber API
├── docs/            # Project documentation
└── package.json     # Root package with workspaces
```

## Consequences

### Positive
- Shared tooling configuration (ESLint, Prettier)
- Easy imports between frontend and backend if needed
- Single repo for all code
- Devcontainer defines entire environment

### Negative
- Repository grows with all projects
- Need discipline to keep modules independent

### Neutral
- Requires npm 7+ or yarn workspaces
- CI/CD needs to handle workspace structure

## Alternatives Considered

### Option 1: True Monorepo (Nx/Turborepo)

**Rejected** because:
- Added complexity for initial setup
- Steeper learning curve for new team members
- Overkill for 2-3 applications
- Can migrate later if needed

### Option 2: Separate Repositories

**Rejected** because:
- Duplicated configuration
- Harder to maintain consistency
- More overhead managing multiple repos
- Difficult to share internal packages

## Related Decisions

- ADR-0002: Angular Signals for State Management
