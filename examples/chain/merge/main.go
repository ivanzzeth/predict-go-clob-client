package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethsig"
	"github.com/joho/godotenv"
	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
	"github.com/shopspring/decimal"
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

	marketIDStr := os.Getenv("MARKET_ID")
	if marketIDStr == "" {
		log.Fatal("MARKET_ID environment variable is required")
	}

	amountStr := os.Getenv("AMOUNT")
	if amountStr == "" {
		log.Fatal("AMOUNT environment variable is required")
	}
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		log.Fatalf("Invalid AMOUNT: %v", err)
	}

	// Get API key from environment
	apiKey := os.Getenv("PREDICT_API_KEY")

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
	// Set cache TTL for market data (e.g., 5 minutes) to reduce API calls
	client, err := predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
		predictclob.WithEOATradingSigner(signer),
		predictclob.WithChainID(big.NewInt(56)), // BNB Mainnet
		predictclob.WithRPCURL(rpcURL),
		predictclob.WithCacheTTL(5 * time.Minute), // 5 minutes cache TTL
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	log.Printf("Client created successfully\n")

	// Parse market ID
	marketID := types.MustMarketIDFromString(marketIDStr)

	// Get market info to display details
	market, err := client.GetMarket(marketID, true)
	if err != nil {
		log.Fatalf("Failed to get market: %v", err)
	}

	fmt.Printf("=== Market Information ===\n")
	fmt.Printf("Market ID: %s\n", marketID.String())
	fmt.Printf("Title: %s\n", market.Title)
	fmt.Printf("Question: %s\n", market.Question)
	fmt.Printf("Condition ID: %s\n", market.ConditionID.Hex())
	fmt.Printf("Is Neg Risk: %v\n", market.IsNegRisk)
	fmt.Printf("Is Yield Bearing: %v\n", market.IsYieldBearing)
	fmt.Printf("Status: %s\n", market.Status.String())
	fmt.Printf("\n")

	// Merge outcome tokens back into collateral
	fmt.Printf("=== Merging Outcome Tokens ===\n")
	fmt.Printf("Amount: %s shares\n", amount.String())
	fmt.Printf("Market ID: %s\n", marketID.String())

	ctx := context.Background()
	txHash, err := client.Merge(ctx, marketID, amount)
	if err != nil {
		log.Fatalf("Failed to merge: %v", err)
	}

	fmt.Printf("\nMerge transaction successful!\n")
	fmt.Printf("Transaction Hash: %s\n", txHash.Hex())
	fmt.Printf("\n")
	fmt.Printf("Note: The transaction will automatically use MergeNegRisk if market.IsNegRisk is true\n")
	fmt.Printf("Note: Market data is cached (useCache=true) to reduce API calls\n")
}
