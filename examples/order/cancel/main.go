package main

import (
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethsig"
	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get required environment variables
	privateKeyHex := os.Getenv("WALLET_PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("WALLET_PRIVATE_KEY environment variable is required")
	}

	orderID := os.Getenv("ORDER_ID")
	if orderID == "" {
		log.Fatal("ORDER_ID environment variable is required")
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Printf("EOA Address: %s", address.Hex())

	// Create EOA signer
	eoaSigner := ethsig.NewEthPrivateKeySigner(privateKey)

	// Get API key from environment
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create client
	client, err := predictclob.NewClient(
		predictclob.WithChainID(big.NewInt(56)),
		predictclob.WithAPIKey(apiKey),
		predictclob.WithEOATradingSigner(eoaSigner, address),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Authenticate
	jwtToken, _, err := client.Authenticate(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}
	client.SetJWTToken(jwtToken)
	log.Printf("Authenticated successfully\n")

	// Cancel order
	log.Println("=== Cancelling order ===")
	fmt.Printf("Order ID: %s\n", orderID)

	result, err := client.CancelOrder(orderID)
	if err != nil {
		log.Fatalf("Failed to cancel order: %v", err)
	}

	fmt.Printf("\nCancel result:\n")
	fmt.Printf("Success: %v\n", result.Success)
	if len(result.Removed) > 0 {
		fmt.Printf("Removed: %v\n", result.Removed)
	}
	if len(result.Noop) > 0 {
		fmt.Printf("Noop (already cancelled/filled): %v\n", result.Noop)
	}
}
