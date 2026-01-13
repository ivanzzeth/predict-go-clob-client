package main

import (
	"fmt"
	"log"
	"os"

	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file from project root directory
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get API key from environment (required)
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create read-only client with API key
	client := predictclob.NewReadOnlyClient(constants.DefaultAPIHost, apiKey)

	// Get market ID from command line or environment
	var marketID types.MarketID
	if len(os.Args) > 1 {
		marketID = types.MustMarketIDFromString(os.Args[1])
	}
	if marketID.IsZero() {
		if marketIDStr := os.Getenv("MARKET_ID"); marketIDStr != "" {
			marketID = types.MustMarketIDFromString(marketIDStr)
		}
	}
	if marketID.IsZero() {
		log.Fatal("Market ID is required (provide as command line argument or MARKET_ID env var)")
	}

	// Call API
	stats, err := client.GetMarketStats(marketID)
	if err != nil {
		log.Fatalf("Error getting market stats: %v", err)
	}

	// Print result with all fields clearly displayed
	fmt.Println("=== Market Statistics ===")
	fmt.Printf("Total Liquidity (USD): %s\n", stats.TotalLiquidityUsd.String())
	fmt.Printf("Total Volume (USD): %s\n", stats.VolumeTotalUsd.String())
	fmt.Printf("24h Volume (USD): %s\n", stats.Volume24hUsd.String())
	fmt.Println("=========================")
}
