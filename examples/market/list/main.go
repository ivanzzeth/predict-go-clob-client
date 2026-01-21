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

// Environment variables:
// - PREDICT_API_KEY: API key (required)
// - MARKET_FIRST: limit number of results (optional)
// - MARKET_AFTER: pagination cursor (optional)
// - MARKET_STATUS: client-side filter by status - UNPAUSED, RESOLVED, etc. (optional)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	apiKey := os.Getenv("PREDICT_API_KEY")
	client := predictclob.NewReadOnlyClient(constants.DefaultAPIHost, apiKey)

	opts := &types.GetMarketsOptions{}

	if first := os.Getenv("MARKET_FIRST"); first != "" {
		opts.First = &first
	}
	if after := os.Getenv("MARKET_AFTER"); after != "" {
		opts.After = &after
	}

	statusFilter := types.MarketStatus(os.Getenv("MARKET_STATUS"))

	resp, err := client.GetMarkets(opts)
	if err != nil {
		log.Fatalf("Failed to get markets: %v", err)
	}

	// Client-side filter by status if specified
	var markets []types.Market
	if statusFilter != "" {
		for _, m := range resp.Data {
			if m.Status == statusFilter {
				markets = append(markets, m)
			}
		}
	} else {
		markets = resp.Data
	}

	fmt.Printf("Found %d markets", len(markets))
	if statusFilter != "" {
		fmt.Printf(" (filtered by status=%s from %d total)", statusFilter, len(resp.Data))
	}
	fmt.Println()
	if resp.Cursor != nil {
		fmt.Printf("Next cursor: %s\n", *resp.Cursor)
	}

	for i, m := range markets {
		fmt.Printf("\n--- Market %d ---\n", i+1)
		fmt.Printf("ID: %s\n", m.ID.String())
		fmt.Printf("Title: %s\n", m.Title)
		fmt.Printf("Status: %s\n", m.Status.String())
		fmt.Printf("Is Yield Bearing: %v\n", m.IsYieldBearing)
		fmt.Printf("Fee Rate Bps: %d\n", m.FeeRateBps)
		fmt.Printf("Outcomes (%d):\n", len(m.Outcomes))
		for j, o := range m.Outcomes {
			fmt.Printf("  %d. %s (OnChainID: %s)\n", j+1, o.Name, o.OnChainID)
		}
	}
}
