# GoQSO

Amateur Radio Contact Logger with PostgreSQL backend

## Features

- Add and manage QSO contacts
- Search contacts by callsign
- View comprehensive statistics
- Export to ADIF format
- PostgreSQL database backend with migrations

## Quick Start

1. **Set up PostgreSQL database** (see [POSTGRESQL_SETUP.md](POSTGRESQL_SETUP.md))

2. **Build the application:**
   ```bash
   go build -o goqso .
   ```

3. **Configure environment variables:**
   ```bash
   export POSTGRES_PASSWORD=your_password
   # See POSTGRESQL_SETUP.md for all options
   ```

4. **Run database migrations:**
   ```bash
   ./goqso migrate up
   ```

5. **Start the application:**
   ```bash
   ./goqso
   ```

## Command Line Usage

```bash
./goqso                    # Start interactive QSO logger
./goqso migrate up         # Run pending database migrations
./goqso migrate down       # Rollback last migration
./goqso migrate status     # Show migration status
./goqso db check           # Test database connection
./goqso help               # Show help message
```

## Migration Management

GoQSO uses [Goose](https://github.com/pressly/goose) for database migration management. Migrations are stored in the `sql/schema/` directory and are embedded in the application binary for easy deployment.

For detailed setup instructions, see [POSTGRESQL_SETUP.md](POSTGRESQL_SETUP.md).