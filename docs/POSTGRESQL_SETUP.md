# GoQSO PostgreSQL Setup Guide

GoQSO has been updated to use PostgreSQL for storing amateur radio contacts instead of JSON files. This provides better performance, data integrity, and querying capabilities.

## Database Migration Management

GoQSO now uses **Goose** for database migration management, providing better control over database schema changes and versioning.

### Migration Commands

- **Run migrations**: `./goqso migrate up`
- **Rollback last migration**: `./goqso migrate down` 
- **Check migration status**: `./goqso migrate status`
- **Test database connection**: `./goqso db check`
- **Show help**: `./goqso help`

## Prerequisites

### 1. PostgreSQL Installation

#### macOS (using Homebrew):
```bash
brew install postgresql
brew services start postgresql
```

#### Ubuntu/Debian:
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

#### Docker (recommended for development):
```bash
docker run --name goqso-postgres \
  -e POSTGRES_USER=goqso \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=goqso \
  -p 5432:5432 \
  -d postgres:15
```

### 2. Database Setup

1. Create a database and user for GoQSO:

```sql
-- Connect as postgres user
sudo -u postgres psql

-- Create database and user
CREATE DATABASE goqso;
CREATE USER goqso WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE goqso TO goqso;
GRANT ALL ON SCHEMA public TO goqso;
GRANT CREATE ON SCHEMA public TO goqso;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO goqso;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO goqso;

-- For PostgreSQL 15+, you may also need:
\c goqso
GRANT ALL ON SCHEMA public TO goqso;
```

## Configuration

GoQSO uses environment variables for database configuration. Create a `.env` file or set these in your environment:

```bash
# Required
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=goqso
export POSTGRES_PASSWORD=your_password
export POSTGRES_DB=goqso

# Optional (defaults shown)
export POSTGRES_SSLMODE=disable
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_HOST` | `localhost` | PostgreSQL server hostname |
| `POSTGRES_PORT` | `5432` | PostgreSQL server port |
| `POSTGRES_USER` | `postgres` | Database username |
| `POSTGRES_PASSWORD` | (empty) | Database password |
| `POSTGRES_DB` | `goqso` | Database name |
| `POSTGRES_SSLMODE` | `disable` | SSL mode (disable/require/verify-full) |

## Building and Running

1. Build the application:
```bash
go build -o goqso .
```

2. Test database connection:
```bash
# Test connection before running migrations
./goqso db check
```

3. Run database migrations:
```bash
# Run all pending migrations
./goqso migrate up

# Check migration status
./goqso migrate status
```

4. Run the application:
```bash
# Using environment variables
export POSTGRES_PASSWORD=your_password
./goqso

# Or inline
POSTGRES_PASSWORD=your_password ./goqso
```

## Database Schema

GoQSO uses Goose migrations to manage database schema. The migrations are embedded in the application and automatically applied when you run `./goqso migrate up`.

### Migration Files

Migrations are stored in the `sql/schema/` directory:
- `001_contacts.sql` - Initial table creation and indexes

### Schema Structure

The main table structure created by migrations:

```sql
CREATE TABLE contacts (
    id SERIAL PRIMARY KEY,
    callsign VARCHAR(20) NOT NULL,
    contact_date DATE NOT NULL,
    time_on VARCHAR(4),
    time_off VARCHAR(4),
    frequency DECIMAL(10,6),
    band VARCHAR(10),
    mode VARCHAR(10),
    rst_sent VARCHAR(10),
    rst_received VARCHAR(10),
    operator_name VARCHAR(100),
    qth VARCHAR(200),
    country VARCHAR(100),
    grid_square VARCHAR(10),
    power_watts INTEGER,
    comment TEXT,
    confirmed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Indexes

The application creates indexes for optimal query performance:
- `idx_contacts_callsign` - for callsign searches
- `idx_contacts_date` - for date-based queries
- `idx_contacts_band` - for band statistics
- `idx_contacts_mode` - for mode statistics
- `idx_contacts_country` - for country statistics
- `idx_contacts_confirmed` - for QSL confirmation queries

## Features

### New PostgreSQL Features:
- **Faster searches**: Indexed callsign and date searches
- **Better data integrity**: Proper data types and constraints
- **Concurrent access**: Multiple users can access the same database
- **Backup/restore**: Standard PostgreSQL backup tools
- **Advanced queries**: Direct SQL access for custom reports

### Existing Features (unchanged):
- Add new QSO contacts
- List all contacts
- Search by callsign
- View comprehensive statistics
- Export to ADIF format

## Troubleshooting

### Connection Issues

1. **"connection refused"**: Check if PostgreSQL is running
   ```bash
   # Check status
   brew services list | grep postgresql  # macOS
   sudo systemctl status postgresql      # Linux
   ```

2. **"password authentication failed"**: Verify credentials
   ```bash
   psql -h localhost -U goqso -d goqso
   ```

3. **"database does not exist"**: Create the database
   ```sql
   sudo -u postgres psql -c "CREATE DATABASE goqso;"
   ```

### Permission Issues

If you get permission errors, ensure your user has proper database privileges:

```sql
-- Connect as postgres superuser
sudo -u postgres psql

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE goqso TO goqso;
\c goqso
GRANT ALL ON SCHEMA public TO goqso;
GRANT ALL ON ALL TABLES IN SCHEMA public TO goqso;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO goqso;
```

### Testing Connection

Test database connectivity using the CLI command:

```bash
./goqso db check
```

Or use the hidden `db` command in the interactive application menu:

```
Enter your choice: db
Database connection successful!
```

## Performance Tips

1. **Regular maintenance**: Run PostgreSQL's VACUUM and ANALYZE periodically
2. **Connection pooling**: For high-usage scenarios, consider pgbouncer
3. **Monitoring**: Use pg_stat_activity to monitor database performance

## Backup and Restore

### Backup
```bash
pg_dump -h localhost -U goqso goqso > goqso_backup.sql
```

### Restore
```bash
psql -h localhost -U goqso -d goqso < goqso_backup.sql
```

## Security Considerations

1. **Use strong passwords** for database users
2. **Enable SSL** in production environments
3. **Firewall rules** to restrict database access
4. **Regular updates** of PostgreSQL server

## Support

For issues with the PostgreSQL integration, please check:
1. PostgreSQL server logs
2. Application error messages
3. Network connectivity
4. User permissions

The application will provide detailed error messages to help diagnose connection and query issues.