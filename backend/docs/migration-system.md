# Migration System Architecture

## Overview

Nova EAM uses [golang-migrate](https://github.com/golang-migrate/migrate) as the core migration engine, with a thin orchestration layer (`cmd/migrate/main.go`) that handles tenant-specific workflows.

## Architecture Diagram

```
nova-migrate CLI
      │
      ▼
┌─────────────────────────────────────────────┐
│  Runner (cmd/migrate/main.go)                │
│                                             │
│  Responsibilities:                          │
│  • Parse CLI flags                          │
│  • Build connection string (search_path)   │
│  • Create tenant schema (if needed)         │
│  • Orchestrate: migrate + seeds             │
│  • Bootstrap workflow                        │
└──────────────────────┬──────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────┐
│  golang-migrate (library)                   │
│                                             │
│  Responsibilities:                          │
│  • Read .up.sql / .down.sql files           │
│  • Execute SQL against PostgreSQL           │
│  • Track version in schema_migrations      │
│  • Checksum verification                     │
│  • Dirty state detection                    │
└──────────────────────┬──────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────┐
│  PostgreSQL                                 │
│                                             │
│  public schema    → global tables           │
│  tenant_{code}    → tenant tables           │
│  schema_migrations → migration tracking      │
└─────────────────────────────────────────────┘
```

## File Structure

```
backend/
├── cmd/migrate/main.go           # Orchestration runner
├── migrations/
│   ├── global/                    # Global schema migrations
│   │   ├── YYYYMMDDHHMMSS_name.up.sql
│   │   └── YYYYMMDDHHMMSS_name.down.sql
│   └── tenant/                    # Tenant schema migrations
│       ├── YYYYMMDDHHMMSS_name.up.sql
│       └── YYYYMMDDHHMMSS_name.down.sql
└── seeds/
    └── tenant/                    # Tenant reference data
        ├── 001_*.sql
        └── 002_*.sql
```

## Commands

| Command | Description |
|---------|-------------|
| `-type=global up` | Apply global migrations |
| `-type=global down` | Rollback last global migration |
| `-type=global status` | Show global migration version |
| `-type=global seed` | Run global seeds |
| `-type=tenant -tenant=CODE up` | Apply tenant migrations |
| `-type=tenant -tenant=CODE down` | Rollback last tenant migration |
| `-type=tenant -tenant=CODE status` | Show tenant migration version |
| `-type=tenant -tenant=CODE seed` | Run tenant seeds |
| `-type=tenant -tenant=CODE bootstrap` | Create schema + migrate + seed |

## Connection String & Search Path

**Global migrations:**
```
postgres://user:pass@host:port/nova?search_path=public
```

**Tenant migrations:**
```
postgres://user:pass@host:port/nova?search_path=tenant_{code},public
```

Note: `public` is included in tenant search_path to allow access to `uuid_generate_v4()` function defined in the global schema.

## golang-migrate Responsibilities

- **Version tracking**: `schema_migrations` table in each schema
- **Checksum verification**: Detects modified migration files
- **Dirty state**: Set when migration fails mid-way
- **Up/Down/Steps**: All migration operations

## Runner Responsibilities

- **Schema creation**: `CREATE SCHEMA IF NOT EXISTS tenant_{code}` (golang-migrate doesn't do this)
- **Search path configuration**: Sets correct schema context
- **Bootstrap orchestration**: Sequences schema → migrate → seeds
- **Seed execution**: Runs reference data scripts separately from migrations