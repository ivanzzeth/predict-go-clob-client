package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

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

	// Parse options from command line or environment
	opts := &types.GetMarketsOptions{}

	// Category ID
	if len(os.Args) > 1 {
		opts.CategoryID = types.CategoryID(os.Args[1])
	}
	if opts.CategoryID == "" {
		if categoryIDStr := os.Getenv("CATEGORY_ID"); categoryIDStr != "" {
			opts.CategoryID = types.CategoryID(categoryIDStr)
		}
	}

	// Limit
	if len(os.Args) > 2 {
		if limit, err := strconv.Atoi(os.Args[2]); err == nil {
			opts.Limit = limit
		}
	}
	if opts.Limit == 0 {
		if limitStr := os.Getenv("MARKET_LIMIT"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil {
				opts.Limit = limit
			}
		}
	}

	// Status (default to OPEN)
	if len(os.Args) > 3 {
		opts.Status = types.MarketStatus(os.Args[3])
	}
	if opts.Status == "" {
		if statusStr := os.Getenv("MARKET_STATUS"); statusStr != "" {
			opts.Status = types.MarketStatus(statusStr)
		} else {
			opts.Status = types.MarketStatusOpen
		}
	}

	// Call API
	markets, err := client.GetMarkets(opts)
	if err != nil {
		log.Fatalf("Error getting markets: %v", err)
	}

	// Print result using %+v to show all fields
	fmt.Printf("Total markets: %d\n\n", len(markets))
	for i, market := range markets {
		fmt.Printf("Market [%d]:\n%+v\n\n", i+1, market)
	}
}
