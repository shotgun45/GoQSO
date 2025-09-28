package main

import (
	"log"

	goqso "goqso/internal"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Start the web server
	goqso.StartServer()
}
