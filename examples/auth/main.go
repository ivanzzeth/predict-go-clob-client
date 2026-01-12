package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
)

func main() {
	// Load .env file from project root
	_ = godotenv.Load(".env")

	// Get private key from environment variable
	privateKey := os.Getenv("WALLET_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("WALLET_PRIVATE_KEY environment variable is required")
	}

	// Get API key from environment variable (optional for testnet)
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create client with API key
	client, err := predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Predict.fun Authentication ===")

	// Perform authentication
	token, address, err := client.Authenticate(privateKey)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Printf("\n✓ Authentication successful!\n")
	fmt.Printf("Wallet Address: %s\n", address.Hex())
	fmt.Printf("JWT Token: %s...\n", token[:50])

	// Verify token is set in client
	if client.GetJWTToken() == "" {
		log.Fatal("JWT token was not set in client")
	}

	fmt.Println("\n✓ JWT token is set in client and will be used for authenticated requests")
}
