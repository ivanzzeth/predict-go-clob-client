package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethsig"
	"github.com/joho/godotenv"
	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
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

	// Get market info to display details and check if neg-risk
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
	fmt.Printf("Outcomes: %d\n", len(market.Outcomes))
	for i, outcome := range market.Outcomes {
		fmt.Printf("  Outcome %d: %s (IndexSet: %d, OnChainID: %s)\n", i+1, outcome.Name, outcome.IndexSet, outcome.OnChainID)
	}
	fmt.Printf("\n")

	// Prepare amounts for neg-risk markets
	var amounts []*big.Int
	if market.IsNegRisk {
		// For neg-risk markets, amounts parameter is required
		// Parse AMOUNTS from environment variable (comma-separated list)
		amountsStr := os.Getenv("AMOUNTS")
		if amountsStr == "" {
			log.Fatal("AMOUNTS environment variable is required for neg-risk markets (comma-separated list of amounts in wei)")
		}

		amountStrs := strings.Split(amountsStr, ",")
		amounts = make([]*big.Int, 0, len(amountStrs))
		for i, amountStr := range amountStrs {
			amountStr = strings.TrimSpace(amountStr)
			amount, ok := new(big.Int).SetString(amountStr, 10)
			if !ok {
				log.Fatalf("Invalid amount at index %d: %s", i, amountStr)
			}
			amounts = append(amounts, amount)
		}

		if len(amounts) != len(market.Outcomes) {
			log.Fatalf("Number of amounts (%d) must match number of outcomes (%d)", len(amounts), len(market.Outcomes))
		}

		fmt.Printf("=== Redeeming Neg-Risk Market ===\n")
		fmt.Printf("Amounts (wei):\n")
		for i, amount := range amounts {
			fmt.Printf("  Outcome %d (%s): %s wei\n", i+1, market.Outcomes[i].Name, amount.String())
		}
	} else {
		// For regular markets, amounts can be nil
		fmt.Printf("=== Redeeming Regular Market ===\n")
		fmt.Printf("Note: No amounts parameter needed for regular markets\n")
	}

	// Redeem positions for resolved market
	fmt.Printf("\n=== Redeeming Positions ===\n")
	fmt.Printf("Market ID: %s\n", marketID.String())

	ctx := context.Background()
	txHash, err := client.Redeem(ctx, marketID, amounts)
	if err != nil {
		log.Fatalf("Failed to redeem: %v", err)
	}

	fmt.Printf("\nRedeem transaction successful!\n")
	fmt.Printf("Transaction Hash: %s\n", txHash.Hex())
	fmt.Printf("\n")
	fmt.Printf("Note: The transaction will automatically use RedeemNegRisk if market.IsNegRisk is true\n")
	fmt.Printf("Note: Market data is cached (useCache=true) to reduce API calls\n")
}
