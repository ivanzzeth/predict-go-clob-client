package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethsig"
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
	privateKeyHex := os.Getenv("WALLET_PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("WALLET_PRIVATE_KEY environment variable is required")
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// Create signer
	signer := ethsig.NewEthPrivateKeySigner(privateKey)

	// Get API key from environment (optional)
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create client (auto-authenticates if Signer, APIKey are set and JWTToken is not)
	client, err := predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
		predictclob.WithEOATradingSigner(signer),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
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
