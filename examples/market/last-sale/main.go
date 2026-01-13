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

	// Get market last sale information
	fmt.Printf("=== Getting last sale for market ID: %s ===\n", marketID.String())
	lastSale, err := client.GetMarketLastSale(marketID)
	if err != nil {
		log.Fatalf("Failed to get last sale: %v", err)
	}

	// Print last sale details with all fields clearly displayed
	printLastSale(lastSale)
}

// printLastSale prints last sale details with all fields clearly displayed
func printLastSale(lastSale *types.MarketLastSale) {
	if lastSale == nil {
		fmt.Println("\n=== Last Sale Information ===")
		fmt.Println("No sale data available (null)")
		return
	}

	fmt.Println("\n=== Last Sale Information ===")
	fmt.Printf("Quote Type: %s\n", lastSale.QuoteType.String())
	fmt.Printf("Outcome: %s\n", lastSale.Outcome.String())
	fmt.Printf("Price In Currency: %s (raw: %s)\n", lastSale.PriceInCurrency.String(), lastSale.RawPriceInCurrency)
	fmt.Printf("Strategy: %s\n", string(lastSale.Strategy))
}
