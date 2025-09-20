package main

import "fmt"

func main() {
	fmt.Printf("GoQSO - Amateur Radio Contact Logger v%s\n", version)
	fmt.Println("========================================")

	logger := NewQSOLogger()

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
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
