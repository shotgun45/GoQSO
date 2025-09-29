# GoQSO

## Motivation

Amateur radio operators need reliable, modern tools to log their contacts (QSOs) and manage their station data. While many logging applications exist, most suffer from outdated interfaces, limited cross-platform support, or expensive licensing models.

GoQSO addresses these challenges by providing:

- **Modern Web Interface** - Built with React/TypeScript for a responsive, intuitive user experience
- **Cross-Platform Compatibility** - Works on any device with a web browser
- **Open Source** - Free to use, modify, and contribute to
- **Standard Compliance** - Full ADIF (Amateur Data Interchange Format) support for interoperability
- **Self-Hosted** - Complete control over your data with no cloud dependencies
- **Extensible Architecture** - RESTful API enables integration with other amateur radio tools

Whether you're a casual operator logging weekend contacts or a serious DXer managing thousands of QSOs, GoQSO provides the modern, reliable platform you need.

## Quick Start

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

## Usage

### Contact Management

**Adding Contacts:**
1. Navigate to the main page at http://localhost:3000
2. Click "Add New Contact" or use the contact form
3. Fill in the QSO details (callsign, date, time, band, mode, etc.)
4. Click "Save Contact" to add to your log

**Searching Contacts:**
- Use the search bar to find contacts by callsign or operator name
- Apply filters for date range, band, mode, country, or frequency
- Use advanced search for complex queries

**Exporting Data:**
1. Go to the Export page from the navigation menu
2. Choose to export all contacts or specify a date range
3. Download your log in ADIF format for use with other amateur radio software

### Administration

**System Monitoring:**
1. Navigate to the Admin page (http://localhost:3000/admin)
2. View application status, database statistics, and system health
3. Monitor contact counts and database size

**Database Management:**
- **Duplicate Detection**: The system automatically detects duplicate contacts and shows warnings
- **Merge Duplicates**: Use the merge tool to clean up duplicate records safely
- **Backup**: Export your complete log before performing maintenance operations

### API Usage

The backend provides a RESTful API accessible at `http://localhost:8080/api`:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/contacts` | List all contacts with optional search parameters |
| `POST` | `/api/contacts` | Add a new contact |
| `PUT` | `/api/contacts/:id` | Update an existing contact |
| `DELETE` | `/api/contacts/:id` | Delete a contact |
| `GET` | `/api/admin/system` | Get system information |
| `POST` | `/api/admin/merge-duplicates` | Merge duplicate contacts |
| `GET` | `/api/contacts/export` | Export contacts in ADIF format |
| `GET` | `/api/version` | Get API version information |

**Search Parameters:**
The `/api/contacts` endpoint supports advanced search:
- `search` - Search callsign or operator name
- `date_from` / `date_to` - Date range filter
- `band` - Amateur radio band filter
- `mode` - Communication mode filter
- `country` - Country filter
- `freq_min` / `freq_max` - Frequency range filter
- `confirmed` - Confirmation status filter

### Command Line Usage

```bash
./goqso                    # Start backend server
./goqso migrate up         # Run pending database migrations
./goqso migrate down       # Rollback last migration
./goqso migrate status     # Show migration status
./goqso db check           # Test database connection
./goqso help               # Show help message
```

### Manual Setup

If you prefer to run components individually:

**Backend Only:**
```bash
# Build the Go application
go build -o goqso .

# Run database migrations (if using PostgreSQL)
./goqso migrate up

# Start the backend server
./goqso
```

**Frontend Only:**
```bash
cd frontend
npm install
npm run dev
```

## Contributing

We welcome contributions from the amateur radio community! Whether you're fixing bugs, adding features, or improving documentation, your help makes GoQSO better for everyone.

### Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/yourusername/GoQSO.git
   cd GoQSO
   ```
3. **Create a feature branch:**
   ```bash
   git checkout -b feature/amazing-feature
   ```

### Development Guidelines

**Code Quality:**
- Run tests before submitting: `go test ./...`
- Follow Go formatting standards: `go fmt ./...`
- Use provided linting tools: `golangci-lint run`
- Run security scans: `gosec ./...` and `govulncheck ./...`

**Frontend Development:**
- Follow TypeScript best practices
- Ensure responsive design works on mobile devices
- Test in multiple browsers
- Build successfully: `npm run build`

**Database Changes:**
- Create migration files for schema changes
- Test migrations both up and down
- Document any breaking changes

### Testing

Run the comprehensive test suite:

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

### Submitting Changes

1. **Commit your changes** with descriptive messages:
   ```bash
   git commit -m 'Add support for FT8 mode detection'
   ```
2. **Push to your branch:**
   ```bash
   git push origin feature/amazing-feature
   ```
3. **Open a Pull Request** on GitHub with:
   - Clear description of changes
   - Screenshots for UI changes
   - Test results
   - Any breaking changes documented

### Areas for Contribution

**High Priority:**
- Contest logging features
- Mobile-responsive improvements
- Performance optimizations

**Medium Priority:**
- Additional export formats (Cabrillo, etc.)
- Real-time backup features
- Multi-operator support
- Custom field definitions
- Band/mode statistics visualization

**Getting Help:**
- Open an issue for questions or discussions
- Check existing issues before creating new ones
- Join the discussion on amateur radio forums
- Reach out via email for complex topics

### Code of Conduct

- Be respectful and professional in all interactions
- Focus on constructive feedback and solutions
- Help newcomers learn and contribute
- Follow amateur radio community values of experimentation and education

---

## ✨ Features

### Core Functionality
- **Contact Management** - Add, edit, and delete QSO contacts with comprehensive details
- **Advanced Search** - Search contacts by callsign, date range, band, mode, country, and more
- **Real-time Statistics** - View comprehensive QSO statistics and summaries
- **ADIF Export** - Export contacts to ADIF format for use with other amateur radio software
- **Duplicate Detection** - Automatic identification and safe merging of duplicate contacts
- **System Administration** - Monitor application health and database status

### Technical Features
- **Modern Web UI** - React/TypeScript frontend with responsive design
- **RESTful API** - Comprehensive HTTP API for all operations
- **Database Migrations** - Automatic PostgreSQL schema management with Goose
- **Environment Configuration** - Flexible configuration via .env files
- **CORS Support** - Cross-origin resource sharing for web applications
- **Security Hardening** - HTTP timeouts and proper error handling

## 📁 Project Structure

```
GoQSO/
├── main.go              # Application entry point
├── server.go            # HTTP server and API routes
├── database.go          # Database connection and operations
├── logger.go            # Contact logging and ADIF operations
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

For detailed database setup instructions, see [POSTGRESQL_SETUP.md](./documentation/POSTGRESQL_SETUP.md).

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

##  🏆 Amateur Radio Bands Supported

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

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 📞 Support

For questions, issues, or feature requests, please open an issue on GitHub.

---

**73 K3JIP** 📻