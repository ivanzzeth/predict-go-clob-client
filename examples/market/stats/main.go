package main

import (
	"fmt"
	"log"
	"os"

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

	// Get API key from environment
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create client
	var client *predictclob.Client
	var err error
	if apiKey != "" {
		client, err = predictclob.NewClient(
			predictclob.WithAPIHost(constants.DefaultAPIHost),
			predictclob.WithAPIKey(apiKey),
		)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
	} else {
		client = predictclob.NewReadOnlyClient(constants.DefaultAPIHost)
	}

	// Get market ID from command line or environment
	var marketID types.MarketID
	if len(os.Args) > 1 {
		marketID = types.MarketID(os.Args[1])
	}
	if marketID == "" {
		if marketIDStr := os.Getenv("MARKET_ID"); marketIDStr != "" {
			marketID = types.MarketID(marketIDStr)
		}
	}
	if marketID == "" {
		log.Fatal("Market ID is required (provide as command line argument or MARKET_ID env var)")
	}

	// Call API
	stats, err := client.GetMarketStats(marketID)
	if err != nil {
		log.Fatalf("Error getting market stats: %v", err)
	}

	// Print result using %+v to show all fields
	fmt.Printf("Market Stats:\n%+v\n", stats)
}
