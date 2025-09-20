#!/bin/bash

# GoQSO Start Script
# Starts both the Go backend server and React frontend development server
# 
# Requirements:
#   - Go 1.23+ 
#   - Node.js 18+
#   - PostgreSQL (optional, but recommended for full functionality)
#
# Usage: ./start.sh
# 
# The script will:
#   1. Load environment variables from .env file
#   2. Start the Go backend server on port 8080
#   3. Start the React frontend development server on port 3000
#   4. Display helpful information about available endpoints
#
# Press Ctrl+C to stop both servers

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[GoQSO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[GoQSO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[GoQSO]${NC} $1"
}

print_error() {
    echo -e "${RED}[GoQSO]${NC} $1"
}

# Function to check if a port is in use
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Function to check PostgreSQL connectivity
check_postgres() {
    if command -v psql &> /dev/null; then
        if psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c '\q' 2>/dev/null; then
            return 0
        fi
    fi
    return 1
}

# Function to stop servers on script exit
cleanup() {
    print_warning "Stopping servers..."
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
    fi
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null || true
    fi
    exit 0
}

# Set up signal handlers for cleanup
trap cleanup SIGINT SIGTERM

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.23+ to run the backend server."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
print_status "Using Go version: $GO_VERSION"

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed. Please install Node.js to run the frontend server."
    exit 1
fi

# Check Node.js version
NODE_VERSION=$(node --version)
print_status "Using Node.js version: $NODE_VERSION"

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    print_error "npm is not installed. Please install npm to run the frontend server."
    exit 1
fi

print_status "Starting GoQSO Amateur Radio Contact Logger..."

# Load environment variables from .env file
if [ -f ".env" ]; then
    print_status "Loading environment variables from .env file..."
    export $(grep -v '^#' .env | xargs)
fi

# Check PostgreSQL connection requirements
print_status "Checking PostgreSQL configuration..."
if [ -z "$POSTGRES_HOST" ] || [ -z "$POSTGRES_DB" ] || [ -z "$POSTGRES_USER" ]; then
    print_warning "PostgreSQL environment variables not fully configured."
    print_warning "Some features may not work without a database connection."
else
    if check_postgres; then
        print_success "PostgreSQL connection successful"
    else
        print_warning "Cannot connect to PostgreSQL at $POSTGRES_HOST:$POSTGRES_PORT"
        print_warning "Database features will not be available until connection is established"
        print_status "Continuing with server startup..."
    fi
fi

# Check if backend is already running
if check_port 8080; then
    print_warning "Backend server is already running on port 8080"
else
    print_status "Starting Go backend server on port 8080..."
    
    # Start the Go server in the background (using current directory structure)
    go run . &
    BACKEND_PID=$!
    
    # Wait a moment for the server to start
    sleep 3
    
    # Check if the backend started successfully
    if check_port 8080; then
        print_success "Backend server started successfully on http://localhost:8080"
        print_status "API endpoints available at: http://localhost:8080/api"
    else
        print_error "Failed to start backend server"
        print_error "Check that PostgreSQL is running and accessible"
        exit 1
    fi
fi

# Check if frontend is already running
if check_port 3000; then
    print_warning "Frontend server is already running on port 3000"
else
    print_status "Starting React frontend development server on port 3000..."
    
    # Navigate to frontend directory and start the development server
    cd frontend
    
    # Check if node_modules exists, if not run npm install
    if [ ! -d "node_modules" ]; then
        print_status "Installing frontend dependencies..."
        npm install
    fi
    
    # Start the frontend server in the background
    npm run dev &
    FRONTEND_PID=$!
    
    # Return to root directory
    cd ..
    
    # Wait a moment for the server to start
    sleep 3
    
    # Check if the frontend started successfully
    if check_port 3000; then
        print_success "Frontend server started successfully on http://localhost:3000"
    else
        print_error "Failed to start frontend server"
        cleanup
        exit 1
    fi
fi

print_success "GoQSO is now running!"
print_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
print_status "🌐 Backend API: http://localhost:8080"
print_status "📱 Frontend App: http://localhost:3000"
print_status "📊 API Documentation: http://localhost:8080/api"
print_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
print_status ""
print_status "📋 Available API endpoints:"
print_status "  GET    /api/contacts     - List all contacts"
print_status "  POST   /api/contacts     - Add new contact"
print_status "  PUT    /api/contacts/:id - Update contact"
print_status "  DELETE /api/contacts/:id - Delete contact"
print_status "  GET    /api/version      - API version info"
print_status ""
print_warning "Press Ctrl+C to stop both servers"

# Keep the script running and wait for user to stop it
wait