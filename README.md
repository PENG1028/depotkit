# Depotly

**Docker-first local data service control tool** for small web projects and lightweight self-hosted development.

Depotly provides a repeatable local control layer for starting, checking, connecting, backing up, and lightly migrating common data services.

## What Depotly Is

- A CLI tool to manage local data services running in Docker
- A configuration-driven approach to PostgreSQL, Redis, MinIO/S3, and MongoDB
- A simple migration runner for PostgreSQL with checksum verification and dirty-state detection
- A tool for backups, schema dumps, and connection information

## What Depotly Is Not

- Not an ORM
- Not a cloud database platform
- Not a Prisma replacement
- Not a Neon or Supabase replacement
- Not a Kubernetes operator
- Not a high-availability or replication management tool

## Supported Services

| Service    | Type                     | Docker Image  |
|------------|--------------------------|---------------|
| PostgreSQL | Relational database      | postgres:16   |
| Redis      | Cache / key-value store  | redis:7       |
| MinIO      | S3-compatible object store | minio/minio |
| MongoDB    | Document database        | mongo:7       |

## Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/engine/install/) (Docker Desktop or Docker Engine)
- Docker Compose (built-in Docker plugin)
- `pg_dump` / `pg_restore` (PostgreSQL client tools — optional, for backup/restore)
- `mongodump` / `mongorestore` (MongoDB database tools — optional, for backup/restore)

### Initialize a Project

```bash
depotly init
```

This creates:

```
depotly.yaml              # Project configuration
db/
  migrations/postgres/     # SQL migration files
  seeds/postgres/          # Seed data (optional)
  schema/postgres/         # Schema output
.datadock/
  runtime/                 # Docker Compose files
  backups/
    postgres/              # PostgreSQL dump files
    mongo/                 # MongoDB dump files
  reports/                 # Operation logs
```

### Check Prerequisites

```bash
depotly doctor
```

Checks: Docker binary, Docker daemon, Docker Compose, file permissions, and port availability.

### Start Services

```bash
depotly up
```

Generates a `docker-compose.yml` from your config and starts all enabled services.

### Verify Services

```bash
depotly status     # Container states
depotly check      # Service connectivity
depotly connect    # Connection strings
```

## Connection Strings

```bash
depotly connect postgres
# DATABASE_URL=postgres://app:app_password@localhost:5432/app

depotly connect redis
# REDIS_URL=redis://localhost:6379

depotly connect object
# S3_ENDPOINT=http://localhost:9000
# S3_ACCESS_KEY=minio
# S3_SECRET_KEY=minio_password
# S3_BUCKET=app

depotly connect mongo
# MONGO_URL=mongodb://localhost:27017/app
```

## PostgreSQL Migration Workflow

### 1. Create Migration Files

Create SQL migration files in `db/migrations/postgres/`:

```
202606231200_create_users.up.sql
202606231200_create_users.down.sql
202606231300_add_status.up.sql
202606231300_add_status.down.sql
```

Filename format: `YYYYMMDDHHMM_description.up.sql` / `YYYYMMDDHHMM_description.down.sql`.

### 2. Check Migration Status

```bash
depotly pg migrate status
```

Shows applied migrations, pending migrations, dirty state, and checksum mismatch warnings.

### 3. Apply Migrations

```bash
depotly pg migrate up
```

Executes pending migrations in order. Each migration runs in a transaction:

1. Inserts a `dirty=true` record in `schema_migrations`
2. Executes the SQL
3. Updates the record to `dirty=false` with execution time

If a migration fails, the dirty state is preserved for investigation.

### 4. Dump Schema

```bash
depotly pg schema-dump
```

Runs `pg_dump --schema-only` and saves to `db/schema/postgres/latest.sql`.

### 5. Backup

```bash
depotly pg backup
```

Creates a compressed dump in `.datadock/backups/postgres/`.

## Redis Key Versioning Strategy

Redis is treated as a cache / temporary state system — no schema migration.

Use versioned keys with TTL:

```
user:123:profile:v1
user:123:profile:v2
```

**Migration workflow:**

1. New code writes `v2` keys
2. Reads prefer `v2`; fall back to `v1` or regenerate from PostgreSQL
3. Old `v1` keys expire by TTL
4. Optionally delete old keys:

```bash
depotly redis flush-namespace --pattern "user:*:profile:v1"
```

### Commands

```bash
depotly redis ping                          # Test connectivity
depotly redis scan --pattern "cache:*"      # List matching keys
depotly redis ttl-check --pattern "cache:*" # Find keys without TTL
depotly redis flush-namespace --pattern "cache:old:*"  # Delete keys (requires confirmation)
```

## Object Storage (MinIO/S3) Strategy

Object storage manages buckets, object keys, prefixes, metadata, and signed URLs.

**Recommended pattern — store references in PostgreSQL:**

| PostgreSQL                    | MinIO/S3                |
|-------------------------------|-------------------------|
| `files.object_key`            | Object body             |
| `files.metadata`              | Object metadata         |
| Source of truth for paths     | Source of truth for content |

**Key migration strategy:**

- Never hardcode object paths in application code
- Use `files.object_key` as the single source of truth
- If a prefix needs to change, write new objects to the new prefix
- Keep old objects until all references are removed

### Commands

```bash
depotly object status                          # Check bucket status
depotly object put-test                        # Upload test object
depotly object list --prefix "uploads/"        # List objects
depotly object signed-url --key "uploads/a.png"  # Generate presigned URL
depotly object clean-test                      # Remove test objects
```

## MongoDB Schema Versioning

MongoDB documents should include a `schemaVersion` field for migration tracking.

### Commands

```bash
depotly mongo status                    # Container status
depotly mongo shell                     # Open mongosh
depotly mongo collections               # List collections
depotly mongo backup                    # Run mongodump
depotly mongo restore                   # Run mongorestore
depotly mongo versions --collection projects  # Show schemaVersion distribution
```

### Checking Schema Versions

```bash
depotly mongo versions --collection projects
```

Output:

```
Collection: projects
Total documents: 478

schemaVersion:
  1: 120 documents
  2: 340 documents
  missing: 18 documents
```

### Migration Script Example

Future migration scripts live in `db/migrations/mongo/`:

```javascript
db.projects.updateMany(
  { schemaVersion: 1 },
  [
    {
      $set: {
        profile: { displayName: "$name" },
        schemaVersion: 2
      }
    }
  ]
)
```

## Configuration

See `depotly.yaml` after running `depotly init`:

```yaml
project: demo-app

runtime:
  mode: docker
  compose_file: .datadock/runtime/docker-compose.yml

services:
  postgres:
    enabled: true
    image: postgres:16
    container_name: demo-postgres
    port: 5432
    database: app
    user: app
    password: app_password
    volume: demo_postgres_data

  redis:
    enabled: true
    image: redis:7
    container_name: demo-redis
    port: 6379
    volume: demo_redis_data

  object:
    enabled: true
    provider: minio
    image: minio/minio
    container_name: demo-minio
    port: 9000
    console_port: 9001
    access_key: minio
    secret_key: minio_password
    bucket: app
    volume: demo_minio_data

  mongo:
    enabled: true
    image: mongo:7
    container_name: demo-mongo
    port: 27017
    database: app
    volume: demo_mongo_data

postgres:
  migrations: db/migrations/postgres
  schema: db/schema/postgres/latest.sql
  backups: .datadock/backups/postgres

mongo:
  backups: .datadock/backups/mongo

object:
  backup_prefix: backups/
```

> **⚠ Security Note:** Plaintext credentials in `depotly.yaml` are acceptable for local development, but not recommended for production environments. For production, use a secrets manager or environment-variable-based configuration.

## All Commands

### Global

| Command | Description |
|---------|-------------|
| `init` | Initialize a new Depotly project |
| `doctor` | Check Docker and prerequisites |
| `up` | Start all services |
| `down` | Stop all services (preserves volumes) |
| `restart` | Restart all services |
| `status` | Show service states |
| `check` | Verify services are reachable |
| `connect [service]` | Print connection strings |
| `reset` | Stop services and delete volumes (requires project name confirmation) |

### PostgreSQL

| Command | Description |
|---------|-------------|
| `pg status` | Show PostgreSQL container status |
| `pg shell` | Open interactive psql session |
| `pg backup` | Create pg_dump backup |
| `pg restore [file]` | Restore from backup |
| `pg schema-dump` | Run pg_dump --schema-only |
| `pg migrate status` | Show migration status |
| `pg migrate up` | Apply pending migrations |

### Redis

| Command | Description |
|---------|-------------|
| `redis status` | Show Redis container status |
| `redis ping` | Ping Redis server |
| `redis scan --pattern "..."` | Scan matching keys |
| `redis ttl-check --pattern "..."` | Check TTL for keys |
| `redis flush-namespace --pattern "..."` | Delete matching keys (requires confirmation) |

### Object Storage

| Command | Description |
|---------|-------------|
| `object status` | Check bucket status |
| `object put-test` | Upload test object |
| `object list --prefix "..."` | List objects with prefix |
| `object signed-url --key "..."` | Generate presigned URL |
| `object clean-test` | Remove test objects |

### MongoDB

| Command | Description |
|---------|-------------|
| `mongo status` | Show MongoDB container status |
| `mongo shell` | Open interactive mongosh session |
| `mongo collections` | List collections |
| `mongo backup` | Run mongodump |
| `mongo restore [dir]` | Run mongorestore |
| `mongo versions --collection ...` | Show schemaVersion distribution |

## Endpoint Exposure Model

StorePilot manages direct endpoints for local database instances. The **Endpoint Exposure Model** adds an optional declaration layer for future routed access.

### Concepts

- **Direct endpoint**: The real, currently-usable connection to your database instance. Used by `connect`, `check`, `migrate`, `backup`, and all other data commands.
- **Exposure**: An optional manifest-only declaration that describes how the database *could* be exposed through a router (e.g., Aegis). Enabling exposure does NOT open ports, start proxies, or modify any database credentials.
- **Provider**: The routing system that would handle the exposed endpoint. Currently only `aegis` is recognized, and it is manifest-only in this version.

### Important

- StorePilot does **not** implement a TCP proxy.
- StorePilot does **not** call any Aegis API in this version.
- StorePilot does **not** guarantee routed endpoint connectivity.
- Enabling exposure does **not** change your `DATABASE_URL`.
- A manifest being present does **not** mean the route has been applied.
- Exposure being enabled does **not** mean the routed endpoint is reachable.
- Without exposure configuration, all existing database features work exactly as before.
- Exposure directory is determined by `runtime.work_dir` in config (defaults to `.datadock`). If both `.storepilot` and `.datadock` exist and `work_dir` is not configured, the command will refuse with a clear error.

### Commands

```bash
depotly endpoint show postgres
depotly endpoint direct postgres
depotly endpoint manifest postgres
depotly endpoint expose postgres --provider aegis
depotly endpoint test postgres
depotly endpoint unexpose postgres
```

### Endpoint Commands

| Command | Description |
|---------|-------------|
| `endpoint show <instance>` | Show endpoint status (direct + exposure) |
| `endpoint direct <instance>` | Print direct connection string (hides password by default) |
| `endpoint manifest <instance>` | Generate exposure manifest to stdout (no files written) |
| `endpoint expose <instance> --provider aegis` | Enable exposure, update config, write manifest file |
| `endpoint test <instance>` | Test direct endpoint connectivity |
| `endpoint unexpose <instance>` | Disable exposure (does not affect database) |

### Example Workflow

```bash
# Show current endpoint status
depotly endpoint show postgres

# Print direct connection info (password masked)
depotly endpoint direct postgres

# Enable exposure
depotly endpoint expose postgres --provider aegis

# View generated manifest
depotly endpoint manifest postgres

# Test direct connectivity
depotly endpoint test postgres

# Disable exposure
depotly endpoint unexpose postgres
```

### Security

- `endpoint direct` hides passwords by default (`--show-secret` to reveal; prints a warning when used).
- `endpoint manifest` never writes database passwords to the manifest.
- For local managed instances with plaintext passwords, the manifest sets `credentials.omitted: true` — no credential reference is recorded.
- `endpoint expose` generates a manifest file under `<work_dir>/exposures/` — this file is a route declaration, not a credential file.
- `--show-secret` output goes to stdout only. Ensure it is not captured in logs or shared.

### MVP Scope

- Exposure is only fully supported for **PostgreSQL** in this version.
- `endpoint show`, `endpoint direct`, and `endpoint test` work for all instance types (PostgreSQL, Redis, MinIO, MongoDB).
- `endpoint manifest` and `endpoint expose` reject non-PostgreSQL instances with a clear error message.
- Aegis provider is **manifest-only** — no API calls, no routing, no proxy.
- `endpoint test` always tests only the direct endpoint. Routed endpoint tests are not available.

## Safety

- **Never** deletes data without confirmation
- **Never** resets Docker volumes without typing the project name
- **Never** runs destructive Redis deletions without confirmation
- **Never** overwrites existing backup files
- **Never** silently ignores failed commands
- **Always** prints clear error messages
- PostgreSQL migration refuses checksum mismatches
- PostgreSQL migration refuses to proceed if a dirty migration exists
- Operation logs are saved in `.datadock/reports/`

## License

MIT

---

**Depotly v0.1** — For local development and lightweight self-hosting. Not intended for high-availability production operations.
