# Database Backup Utility

[Go Version](https://go.dev/)
[License: MIT](LICENSE)
[Platform]()
[Docker]()

A cross-platform Go CLI for compressed full backups and restores of PostgreSQL. It streams database command output through gzip, supports local and S3-compatible storage, and can send webhook notifications on completion.

> **Note:** Only PostgreSQL is currently implemented. MySQL and MongoDB are underdeveloped. The MySQL and MongoDB adapters are placeholders and will return a "not implemented" error.

---

## Features

- **Compressed archives** — backup output is streamed directly through gzip
- **Timestamped backups** — each run produces a uniquely named archive
- **Flexible storage** — save locally or upload to any S3-compatible object store (AWS S3, MinIO, Cloudflare R2)
- **Webhook notifications** — receive a small JSON payload (no credentials included) when a job completes
- **Connection checks** — verify database connectivity before running a backup
- **Cross-platform** — runs on Windows, Linux, and macOS; Docker image available

## Prerequisites

- **Go 1.22+**
- **PostgreSQL client tools** available on your `PATH`:
  - `pg_dump`
  - `pg_restore`
  - `pg_isready`

## Installation

```sh
go build -o dbbackup.exe ./src/cmd/dbbackup
./dbbackup.exe --help
```

On Linux/macOS, drop the `.exe` suffix:

```sh
go build -o dbbackup ./src/cmd/dbbackup
./dbbackup --help
```

## Quick Start

### 1. Configure the database connection

The tool reads connection details from environment variables.

**PowerShell (Windows):**

```powershell
$env:PGHOST = "localhost"
$env:PGPORT = "5432"
$env:PGUSER = "postgres"
$env:PGPASSWORD = "your_password"
$env:PGDATABASE = "your_database_name"
```

**Linux/macOS:**

```sh
export PGHOST="localhost"
export PGPORT="5432"
export PGUSER="postgres"
export PGPASSWORD="your_password"
export PGDATABASE="your_database_name"
```

### 2. Check the connection

Verify the tool can reach your database before running a backup:

```sh
./dbbackup.exe check --db postgres
```

On success, this prints `connection ok`.

### 3. Run a backup

```sh
./dbbackup.exe backup --db postgres --mode full --storage local --output-dir ./backups
```

This creates a timestamped, compressed archive in the `./backups/postgres/` directory.

### 4. Restore a backup

> [!WARNING]
> The `--clean` flag drops existing objects in the target database before restoring. Use with caution.

```sh
./dbbackup.exe restore --db postgres --storage local --output-dir ./backups --key postgres/YOUR_BACKUP_FILE.dump.gz --clean
```

## S3-Compatible Storage (Optional)

To upload backups to AWS S3, MinIO, or Cloudflare R2, use `--storage s3` and set the following environment variables:

**PowerShell (Windows):**

```powershell
$env:AWS_REGION = "us-east-1"
$env:S3_BUCKET = "my-backups"
$env:S3_PREFIX = "prod/postgres"
# Optional, for MinIO/R2:
# $env:S3_ENDPOINT = "https://..."
```

**Linux/macOS:**

```sh
export AWS_REGION="us-east-1"
export S3_BUCKET="my-backups"
export S3_PREFIX="prod/postgres"
# Optional, for MinIO/R2:
# export S3_ENDPOINT="https://..."
```

## Webhook Notifications (Optional)

Set `WEBHOOK_URL` to receive a small JSON body on completion. No credentials are included in the payload.

```sh
export WEBHOOK_URL="https://example.com/hook"
```

## Environment Variables Reference

| Variable      | Required | Description                                         |
| ------------- | -------- | --------------------------------------------------- |
| `PGHOST`      | Yes      | Database host                                       |
| `PGPORT`      | Yes      | Database port (e.g. `5432`)                         |
| `PGUSER`      | Yes      | Database user                                       |
| `PGPASSWORD`  | Yes      | Database password                                   |
| `PGDATABASE`  | Yes      | Database name                                       |
| `AWS_REGION`  | S3 only  | S3 region                                           |
| `S3_BUCKET`   | S3 only  | Target bucket                                       |
| `S3_PREFIX`   | S3 only  | Key prefix inside the bucket (e.g. `prod/postgres`) |
| `S3_ENDPOINT` | No       | Custom endpoint for MinIO / Cloudflare R2           |
| `WEBHOOK_URL` | No       | URL that receives a JSON completion notification    |
| `PG_ISREADY`  | No       | Full path to `pg_isready` (see Troubleshooting)     |
| `PG_DUMP`     | No       | Full path to `pg_dump` (see Troubleshooting)        |
| `PG_RESTORE`  | No       | Full path to `pg_restore` (see Troubleshooting)     |

## Troubleshooting (Windows)

If you get an empty error like:

```text
Error: postgres connection check failed:
```

it means Go cannot find the PostgreSQL executables (`pg_dump`, `pg_restore`, `pg_isready`). Fix it in one of two ways:

**Option A — add the PostgreSQL** `bin` **directory to your system PATH:**

Add `C:\Program Files\PostgreSQL\17\bin` to your Windows System PATH, then restart your terminal.

**Option B — point the tool at the executables explicitly:**

```powershell
$env:PG_ISREADY = "C:\Program Files\PostgreSQL\17\bin\pg_isready.exe"
$env:PG_DUMP = "C:\Program Files\PostgreSQL\17\bin\pg_dump.exe"
$env:PG_RESTORE = "C:\Program Files\PostgreSQL\17\bin\pg_restore.exe"
```

## Docker (Optional)

You can build and run the tool inside a Docker container, which includes the PostgreSQL client tools automatically:

```sh
docker build -t dbbackup .
docker run --rm dbbackup --help
```

## License

This project is licensed under the [MIT License](LICENSE).
