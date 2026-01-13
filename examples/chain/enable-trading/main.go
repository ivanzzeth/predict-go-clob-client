package main

import (
	"context"
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
	// Load .env file from project root directory
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get required environment variables
	privateKeyHex := os.Getenv("WALLET_PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("WALLET_PRIVATE_KEY environment variable is required")
	}

	// Get RPC URL from environment (required for chain operations)
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
	address := signer.GetAddress()
	log.Printf("EOA Address: %s", address.Hex())

	// Create client with signer and RPC URL
	// Note: EnableTrading only requires chain operations, no API authentication needed
	client, err := predictclob.NewClient(
		predictclob.WithEOATradingSigner(signer),
		predictclob.WithChainID(big.NewInt(56)), // BNB Mainnet
		predictclob.WithRPCURL(rpcURL),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	log.Printf("Client created successfully\n")

	// Enable trading by approving necessary tokens on both YB and NYB contracts
	fmt.Println("=== Enabling Trading ===")
	fmt.Println("This will approve USDT tokens for both Yield Bearing (YB) and Non-Yield Bearing (NYB) contracts")
	fmt.Println()

	ctx := context.Background()
	txHashes, err := client.EnableTrading(ctx)
	if err != nil {
		log.Fatalf("Failed to enable trading: %v", err)
	}

	fmt.Println("\n✅ Trading successfully enabled!")
	fmt.Printf("EOA Address: %s\n", address.Hex())
	fmt.Printf("\nTransaction hashes (%d):\n", len(txHashes))
	for i, txHash := range txHashes {
		fmt.Printf("  %d. %s\n", i+1, txHash.Hex())
		fmt.Printf("     View transaction: https://bscscan.com/tx/%s\n", txHash.Hex())
	}
	fmt.Println()
	fmt.Println("Note: After these transactions are confirmed, you can use Split/Merge/Redeem operations")
}
