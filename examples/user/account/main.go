package main

import (
	"fmt"
	"log"
	"os"

	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file from project root directory
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get private key from environment
	privateKey := os.Getenv("WALLET_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("WALLET_PRIVATE_KEY environment variable is required")
	}

	// Get API key from environment (optional)
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create client
	client, err := predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Authenticate first
	_, _, err = client.Authenticate(privateKey)
	if err != nil {
		log.Fatalf("Error authenticating: %v", err)
	}

	// Call API
	account, err := client.GetAccount()
	if err != nil {
		log.Fatalf("Error getting account: %v", err)
	}

	// Print account information
	fmt.Println("=== Account Information ===")
	fmt.Printf("Name: %s\n", account.Name)
	fmt.Printf("Address: %s\n", account.Address)
	if account.ImageURL != nil {
		fmt.Printf("Image URL: %s\n", *account.ImageURL)
	} else {
		fmt.Printf("Image URL: (null)\n")
	}

	if account.Referral != nil {
		fmt.Println("\n=== Referral Information ===")
		if account.Referral.Code != nil {
			fmt.Printf("Code: %s\n", *account.Referral.Code)
		} else {
			fmt.Printf("Code: (null)\n")
		}
		fmt.Printf("Status: %s\n", account.Referral.Status)
	} else {
		fmt.Println("\nReferral: (not available)")
	}
}
