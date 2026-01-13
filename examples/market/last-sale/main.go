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
	sale, err := client.GetMarketLastSale(marketID)
	if err != nil {
		log.Fatalf("Error getting last sale: %v", err)
	}

	// Print result using %+v to show all fields
	fmt.Printf("Last Sale:\n%+v\n", sale)
}
