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

	// Get private key from environment
	privateKey := os.Getenv("WALLET_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("WALLET_PRIVATE_KEY environment variable is required")
	}

	// Get API key from environment (optional)
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create client
	client, err := predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Authenticate first
	_, _, err = client.Authenticate(privateKey)
	if err != nil {
		log.Fatalf("Error authenticating: %v", err)
	}

	// Parse options from command line or environment
	opts := &types.GetPositionsOptions{}

	// Market ID
	if len(os.Args) > 1 {
		opts.MarketID = types.MarketID(os.Args[1])
	}
	if opts.MarketID == "" {
		if marketIDStr := os.Getenv("MARKET_ID"); marketIDStr != "" {
			opts.MarketID = types.MarketID(marketIDStr)
		}
	}

	// First (limit)
	if len(os.Args) > 2 {
		if first, err := strconv.Atoi(os.Args[2]); err == nil {
			opts.First = first
		}
	}
	if opts.First == 0 {
		if firstStr := os.Getenv("POSITION_LIMIT"); firstStr != "" {
			if first, err := strconv.Atoi(firstStr); err == nil {
				opts.First = first
			}
		}
	}

	// After (cursor)
	if len(os.Args) > 3 {
		opts.After = os.Args[3]
	}
	if opts.After == "" {
		opts.After = os.Getenv("POSITION_CURSOR")
	}

	// Call API
	positions, err := client.GetPositions(opts)
	if err != nil {
		log.Fatalf("Error getting positions: %v", err)
	}

	// Print result using %+v to show all fields
	fmt.Printf("Total positions: %d\n\n", len(positions))
	for i, position := range positions {
		fmt.Printf("Position [%d]:\n%+v\n\n", i+1, position)
	}
}
