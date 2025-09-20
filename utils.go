package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// getUserInput prompts for user input and returns the trimmed string
func getUserInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

// printMainMenu displays the main menu options
func printMainMenu() {
	fmt.Println("\n=== MAIN MENU ===")
	fmt.Println("1. Add new QSO")
	fmt.Println("2. List all QSOs")
	fmt.Println("3. Search QSOs")
	fmt.Println("4. Show statistics")
	fmt.Println("5. Export to ADIF")
	fmt.Println("6. Exit")
}

// frequencyToBand converts frequency in MHz to amateur radio band
func frequencyToBand(freq float64) string {
	switch {
	case freq >= 1.8 && freq <= 2.0:
		return "160m"
	case freq >= 3.5 && freq <= 4.0:
		return "80m"
	case freq >= 5.3 && freq <= 5.4:
		return "60m"
	case freq >= 7.0 && freq <= 7.3:
		return "40m"
	case freq >= 10.1 && freq <= 10.15:
		return "30m"
	case freq >= 14.0 && freq <= 14.35:
		return "20m"
	case freq >= 18.068 && freq <= 18.168:
		return "17m"
	case freq >= 21.0 && freq <= 21.45:
		return "15m"
	case freq >= 24.89 && freq <= 24.99:
		return "12m"
	case freq >= 28.0 && freq <= 29.7:
		return "10m"
	case freq >= 50.0 && freq <= 54.0:
		return "6m"
	case freq >= 144.0 && freq <= 148.0:
		return "2m"
	case freq >= 420.0 && freq <= 450.0:
		return "70cm"
	default:
		return "Unknown"
	}
}
