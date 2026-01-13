package main

import (
	"fmt"
	"log"
	"os"
	"time"

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
	fmt.Printf("=== Getting orderbook for market ID: %s ===\n", marketID.String())
	orderbook, err := client.GetMarketOrderbook(marketID)
	if err != nil {
		log.Fatalf("Error getting orderbook: %v", err)
	}

	// Print result
	printOrderbook(orderbook)
}

// printOrderbook prints orderbook details with all fields clearly displayed
func printOrderbook(orderbook *types.Orderbook) {
	fmt.Printf("\n=== Orderbook Details ===\n")
	fmt.Printf("Market ID: %s\n", orderbook.MarketID.String())
	fmt.Printf("Update Timestamp (ms): %d\n", orderbook.UpdateTimestampMs)
	fmt.Printf("Update Timestamp: %s\n", time.Unix(orderbook.UpdateTimestampMs/1000, (orderbook.UpdateTimestampMs%1000)*1000000).Format("2006-01-02 15:04:05.000"))

	// Print LastOrderSettled
	fmt.Printf("\n=== Last Order Settled ===\n")
	if orderbook.LastOrderSettled != nil {
		fmt.Printf("ID: %s\n", orderbook.LastOrderSettled.ID)
		fmt.Printf("Price: %s\n", orderbook.LastOrderSettled.Price)
		fmt.Printf("Kind: %s\n", string(orderbook.LastOrderSettled.Kind))
		fmt.Printf("Market ID: %s\n", orderbook.LastOrderSettled.MarketID.String())
		fmt.Printf("Side: %s\n", orderbook.LastOrderSettled.Side.String())
		fmt.Printf("Outcome: %s\n", orderbook.LastOrderSettled.Outcome.String())
	} else {
		fmt.Println("  (no last order settled)")
	}

	// Print bids
	fmt.Printf("\n=== Bids (%d) ===\n", len(orderbook.Bids))
	if len(orderbook.Bids) == 0 {
		fmt.Println("  (no bids)")
	} else {
		for i, bid := range orderbook.Bids {
			fmt.Printf("  Bid %d:\n", i+1)
			fmt.Printf("    Price: %s (raw: %s)\n", bid.Price.String(), bid.RawPrice)
			fmt.Printf("    Amount: %s (raw: %s)\n", bid.Amount.String(), bid.RawAmount)
		}
	}

	// Print asks
	fmt.Printf("\n=== Asks (%d) ===\n", len(orderbook.Asks))
	if len(orderbook.Asks) == 0 {
		fmt.Println("  (no asks)")
	} else {
		for i, ask := range orderbook.Asks {
			fmt.Printf("  Ask %d:\n", i+1)
			fmt.Printf("    Price: %s (raw: %s)\n", ask.Price.String(), ask.RawPrice)
			fmt.Printf("    Amount: %s (raw: %s)\n", ask.Amount.String(), ask.RawAmount)
		}
	}

	// Print calculated fields
	fmt.Printf("\n=== Calculated Fields ===\n")
	if len(orderbook.Bids) > 0 {
		fmt.Printf("Best Bid: %s\n", orderbook.BestBid.String())
	} else {
		fmt.Printf("Best Bid: (no bids)\n")
	}
	if len(orderbook.Asks) > 0 {
		fmt.Printf("Best Ask: %s\n", orderbook.BestAsk.String())
	} else {
		fmt.Printf("Best Ask: (no asks)\n")
	}
	if len(orderbook.Bids) > 0 && len(orderbook.Asks) > 0 {
		fmt.Printf("Spread: %s\n", orderbook.Spread.String())
	} else {
		fmt.Printf("Spread: (cannot calculate - missing bids or asks)\n")
	}
}
