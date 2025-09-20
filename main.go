package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	// Check if running migration commands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			if len(os.Args) > 2 {
				switch os.Args[2] {
				case "up":
					if err := MigrateUp(); err != nil {
						log.Fatalf("Migration failed: %v", err)
					}
					return
				case "down":
					if err := MigrateDown(); err != nil {
						log.Fatalf("Migration rollback failed: %v", err)
					}
					return
				case "status":
					if err := MigrateStatus(); err != nil {
						log.Fatalf("Migration status failed: %v", err)
					}
					return
				default:
					fmt.Println("Usage: goqso migrate [up|down|status]")
					os.Exit(1)
				}
			} else {
				fmt.Println("Usage: goqso migrate [up|down|status]")
				os.Exit(1)
			}
		case "db":
			if len(os.Args) > 2 && os.Args[2] == "check" {
				if err := CheckDatabaseConnection(); err != nil {
					log.Fatalf("Database connection failed: %v", err)
				}
				return
			}
		case "help", "--help", "-h":
			printHelp()
			return
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			printHelp()
			os.Exit(1)
		}
	}

	fmt.Printf("GoQSO - Amateur Radio Contact Logger v%s\n", version)
	fmt.Println("========================================")

	// Initialize database connection
	logger, err := NewQSOLogger()
	if err != nil {
		log.Fatalf("Failed to initialize QSO logger: %v", err)
	}
	defer logger.Close()

	fmt.Println("Connected to PostgreSQL database successfully!")

	for {
		printMainMenu()
		choice := getUserInput("Enter your choice: ")

		switch choice {
		case "1":
			logger.AddContact()
		case "2":
			logger.ListContacts()
		case "3":
			logger.SearchContacts()
		case "4":
			logger.ShowStatistics()
		case "5":
			logger.ExportADIF()
		case "6":
			fmt.Println("\n73! Thanks for using GoQSO!")
			return
		case "db":
			// Hidden option to test database connection
			if err := CheckDatabaseConnection(); err != nil {
				fmt.Printf("Database connection failed: %v\n", err)
			} else {
				fmt.Println("Database connection successful!")
			}
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
