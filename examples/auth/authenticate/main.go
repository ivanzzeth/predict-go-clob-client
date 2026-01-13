package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethsig"
	"github.com/joho/godotenv"
	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
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
	address := signer.GetAddress()

	// Get API key from environment (optional)
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create client
	// Note: Client will auto-authenticate if Signer, APIKey are set and JWTToken is not.
	// This example demonstrates manual authentication for educational purposes.
	client, err := predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
		predictclob.WithEOATradingSigner(signer),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Call API
	token, address, err := client.Authenticate()
	if err != nil {
		log.Fatalf("Error authenticating: %v", err)
	}

	// Print result
	fmt.Printf("Authentication successful!\n")
	fmt.Printf("Wallet Address: %s\n", address.Hex())
	fmt.Printf("JWT Token: %s\n", token)
	fmt.Printf("\nAddress struct:\n%+v\n", address)
}
