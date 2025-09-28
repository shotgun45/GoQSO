# GoQSO

Modern Amateur Radio Contact Logger with React Frontend and PostgreSQL Backend

A comprehensive amateur radio contact logging application featuring a modern React/TypeScript frontend, robust Go backend API, and PostgreSQL database with automatic migrations.

## ✨ Features

### Core Functionality
- **Contact Management** - Add, edit, and delete QSO contacts with comprehensive details
- **Advanced Search** - Search contacts by callsign, date range, band, mode, country, and more
- **Real-time Statistics** - View comprehensive QSO statistics and summaries
- **ADIF Export** - Export contacts to ADIF format for use with other amateur radio software

### Technical Features
- **Modern Web UI** - React/TypeScript frontend with responsive design
- **RESTful API** - Comprehensive HTTP API for all operations
- **Database Migrations** - Automatic PostgreSQL schema management with Goose
- **Environment Configuration** - Flexible configuration via .env files
- **CORS Support** - Cross-origin resource sharing for web applications
- **Security Hardening** - HTTP timeouts and proper error handling

## 🚀 Quick Start

### Prerequisites
- **Go 1.23+** - Backend server and CLI tools
- **Node.js 20.19+** - Frontend development and build tools (Vite requirement)
- **PostgreSQL** (optional) - Database backend (can run without for development)

### Development Setup

1. **Clone and setup the repository:**
   ```bash
   git clone <repository-url>
   cd GoQSO
   ```

2. **Configure environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your PostgreSQL configuration
   ```

3. **Start the development environment:**
   ```bash
   ./start.sh
   ```

   The start script will:
   - Load environment variables from `.env` file
   - Check system dependencies (Go, Node.js)
   - Test PostgreSQL connectivity (if configured)
   - Start the Go backend server on port 8080
   - Start the React frontend development server on port 3000

### Access Your Application

After running `./start.sh`, the application will be available at:

- **🌐 Frontend Application**: http://localhost:3000
- **📊 Backend API**: http://localhost:8080

## 🛠️ Manual Setup

If you prefer to run components individually:

### Backend Only
```bash
# Build the Go application
go build -o goqso .

# Run database migrations (if using PostgreSQL)
./goqso migrate up

# Start the backend server
./goqso
```

### Frontend Only
```bash
cd frontend
npm install
npm run dev
```

## 📝 Command Line Usage

```bash
./goqso                    # Start backend server
./goqso migrate up         # Run pending database migrations
./goqso migrate down       # Rollback last migration
./goqso migrate status     # Show migration status
./goqso db check           # Test database connection
./goqso help               # Show help message
```

## 🌐 API Endpoints

The backend provides a RESTful API accessible at `http://localhost:8080/api`:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/contacts` | List all contacts with optional search parameters |
| `POST` | `/api/contacts` | Add a new contact |
| `PUT` | `/api/contacts/:id` | Update an existing contact |
| `DELETE` | `/api/contacts/:id` | Delete a contact |
| `GET` | `/api/version` | Get API version information |

### Search Parameters
The `/api/contacts` endpoint supports advanced search:
- `search` - Search callsign or operator name
- `date_from` / `date_to` - Date range filter
- `band` - Amateur radio band filter
- `mode` - Communication mode filter
- `country` - Country filter
- `freq_min` / `freq_max` - Frequency range filter
- `confirmed` - Confirmation status filter

## 📁 Project Structure

```
GoQSO/
├── main.go              # Application entry point
├── server.go            # HTTP server and API routes
├── database.go          # Database connection and operations
├── logger.go            # Logging utilities
├── utils.go             # Utility functions (band calculations, etc.)
├── start.sh             # Development startup script
├── frontend/            # React/TypeScript frontend
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── api/         # API client functions
│   │   └── types/       # TypeScript type definitions
│   └── package.json     # Frontend dependencies
├── sql/                 # Database schema and migrations
├── docs/                # Documentation
└── README.md           # This file
```

## 🗄️ Database Setup

GoQSO uses PostgreSQL with automatic schema migrations powered by [Goose](https://github.com/pressly/goose).

For detailed database setup instructions, see [POSTGRESQL_SETUP.md](./docs/POSTGRESQL_SETUP.md).

### Environment Variables

Configure your database connection in `.env`:

```bash
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=goqso
POSTGRES_USER=your_username
POSTGRES_PASSWORD=your_password
POSTGRES_SSLMODE=disable
```

## 🧪 Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test files
go test -v ./logger_test.go
go test -v ./server_test.go
go test -v ./utils_test.go
```

## 🔧 Development

### Code Quality Tools

The project includes several code quality tools:

```bash
# Lint code
golangci-lint run

# Security scanning
gosec ./...
govulncheck ./...

# Format code
go fmt ./...
```

### Building for Production

```bash
# Build backend
go build -o goqso .

# Build frontend
cd frontend
npm run build
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🏆 Amateur Radio Bands Supported

| Band | Frequency Range | Notes |
|------|----------------|-------|
| 160m | 1.8 - 2.0 MHz | Long wave |
| 80m | 3.5 - 4.0 MHz | Medium wave |
| 60m | 5.3 - 5.4 MHz | Channel allocations |
| 40m | 7.0 - 7.3 MHz | International shortwave |
| 30m | 10.1 - 10.15 MHz | Digital modes preferred |
| 20m | 14.0 - 14.35 MHz | Primary DX band |
| 17m | 18.068 - 18.168 MHz | WARC band |
| 15m | 21.0 - 21.45 MHz | High frequency |
| 12m | 24.89 - 24.99 MHz | WARC band |
| 10m | 28.0 - 29.7 MHz | Sporadic E propagation |
| 6m | 50.0 - 54.0 MHz | Magic band |
| 2m | 144.0 - 148.0 MHz | VHF |
| 70cm | 420.0 - 450.0 MHz | UHF |

## 📞 Support

For questions, issues, or feature requests, please open an issue on GitHub.

---

**73 K3JIP** 📻