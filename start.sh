#!/bin/bash

# GoQSO Start Script
# Starts both the Go backend server and React frontend development server

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
    print_error "Go is not installed. Please install Go to run the backend server."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed. Please install Node.js to run the frontend server."
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    print_error "npm is not installed. Please install npm to run the frontend server."
    exit 1
fi

print_status "Starting GoQSO Amateur Radio Contact Logger..."

# Check if backend is already running
if check_port 8080; then
    print_warning "Backend server is already running on port 8080"
else
    print_status "Starting Go backend server on port 8080..."
    
    # Start the Go server in the background
    go run main.go server.go &
    BACKEND_PID=$!
    
    # Wait a moment for the server to start
    sleep 2
    
    # Check if the backend started successfully
    if check_port 8080; then
        print_success "Backend server started successfully on http://localhost:8080"
    else
        print_error "Failed to start backend server"
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
print_status "Backend API: http://localhost:8080"
print_status "Frontend App: http://localhost:3000"
print_status ""
print_status "Press Ctrl+C to stop both servers"

# Keep the script running and wait for user to stop it
wait