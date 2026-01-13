package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

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

	// Get API key from environment (optional)
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Get RPC URL from environment (required for balance query)
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		// Default to BNB Mainnet public RPC
		rpcURL = "https://bsc-dataseed1.binance.org"
		log.Printf("Using default RPC URL: %s", rpcURL)
	}

	// Parse private key and create signer
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}
	signer := ethsig.NewEthPrivateKeySigner(privateKey)

	// Create client with signer and RPC URL (auto-authenticates if Signer, APIKey are set and JWTToken is not)
	client, err := predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
		predictclob.WithEOATradingSigner(signer),
		predictclob.WithChainID(big.NewInt(56)), // BNB Mainnet
		predictclob.WithRPCURL(rpcURL),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Call API
	ctx := context.Background()
	balance, err := client.GetBalance(ctx)
	if err != nil {
		log.Fatalf("Error getting balance: %v", err)
	}

	// Print result
	fmt.Printf("=== Collateral Balance ===\n")
	fmt.Printf("Total: %s\n", balance.Total.String())
	fmt.Printf("Locked: %s\n", balance.Locked.String())
	fmt.Printf("Available: %s\n", balance.Available.String())
}
