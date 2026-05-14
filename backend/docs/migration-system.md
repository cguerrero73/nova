# Migration System Architecture

## Overview

Nova EAM uses [golang-migrate](https://github.com/golang-migrate/migrate) as the core migration engine, with a thin orchestration layer (`cmd/migrate/main.go`) that handles tenant-specific workflows.

## Key Design Decision: Seeds as Versioned Migrations

Seeds (reference data like syscodes and default roles) are versioned as part of migrations, not separate files. This ensures:
- **Unified version tracking**: golang-migrate tracks all schema + data migrations
- **Automatic ordering**: Seeds run in correct sequence after schema creation
- **Rollback support**: Each seed migration has a `.down.sql` stub (typically no-op for reference data)

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
│  • Build connection string (search_path)    │
│  • Create tenant schema (if needed)         │
│  • Run migrations (schema + data)           │
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
│   └── tenant/                    # Tenant schema + data migrations
│       ├── YYYYMMDDHHMMSS_init_*.up.sql      # Schema creation
│       ├── YYYYMMDDHHMMSS_init_*.down.sql   # Schema rollback
│       ├── YYYYMMDDHHMMSS_seed_*.up.sql     # Reference data
│       └── YYYYMMDDHHMMSS_seed_*.down.sql   # Data rollback (no-op)
```

### Migration Ordering

Tenant migrations run in timestamp order:

```
000 - init_tenant.up.sql     → Creates tables (eamorganizations, eamusers, etc.)
001 - seed_syscodes.up.sql   → Populates system codes (OBTP, OBST, JBTP, JBST, etc.)
002 - seed_defaults.up.sql   → Populates default roles and common org
```

## Commands

| Command | Description |
|---------|-------------|
| `-type=global up` | Apply all global migrations |
| `-type=global down` | Rollback last global migration |
| `-type=global status` | Show global migration version |
| `-type=tenant -tenant=CODE up` | Apply tenant schema + data migrations |
| `-type=tenant -tenant=CODE down` | Rollback last tenant migration |
| `-type=tenant -tenant=CODE status` | Show tenant migration version |
| `-type=tenant -tenant=CODE bootstrap` | Create schema + apply migrations |

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
- **Bootstrap orchestration**: Sequences schema → migrate (which includes seeds)