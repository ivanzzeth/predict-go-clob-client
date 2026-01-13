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

	// Example 1: Get all markets
	fmt.Println("=== Example 1: Get all markets ===")
	allMarketsResp, err := client.GetMarkets(nil)
	if err != nil {
		log.Fatalf("Failed to get all markets: %v", err)
	}
	fmt.Printf("Found %d markets\n", len(allMarketsResp.Data))
	if allMarketsResp.Cursor != nil {
		fmt.Printf("Next page cursor: %s\n", *allMarketsResp.Cursor)
	}
	printMarkets(allMarketsResp.Data)

	// Example 2: Get markets with pagination (first 10)
	fmt.Println("\n=== Example 2: Get markets with pagination (first 10) ===")
	first := "10"
	paginationOpts := &types.GetMarketsOptions{
		First: &first,
	}
	paginatedResp, err := client.GetMarkets(paginationOpts)
	if err != nil {
		log.Fatalf("Failed to get paginated markets: %v", err)
	}
	fmt.Printf("Found %d markets (requested first 10)\n", len(paginatedResp.Data))
	if paginatedResp.Cursor != nil {
		fmt.Printf("Next page cursor: %s\n", *paginatedResp.Cursor)
	}
	printMarkets(paginatedResp.Data)

	// Example 3: Get next page if cursor exists
	if paginatedResp.Cursor != nil && *paginatedResp.Cursor != "" {
		fmt.Println("\n=== Example 3: Get next page using cursor ===")
		nextOpts := &types.GetMarketsOptions{
			First: &first,
			After: paginatedResp.Cursor,
		}
		nextResp, err := client.GetMarkets(nextOpts)
		if err != nil {
			log.Fatalf("Failed to get next page: %v", err)
		}
		fmt.Printf("Found %d more markets\n", len(nextResp.Data))
		printMarkets(nextResp.Data)
	}

	// Example 4: Get markets with custom filters from command line or environment
	fmt.Println("\n=== Example 4: Get markets with custom filters ===")
	opts := buildOptionsFromArgs()
	if opts != nil {
		customResp, err := client.GetMarkets(opts)
		if err != nil {
			log.Fatalf("Failed to get markets with custom filters: %v", err)
		}
		fmt.Printf("Found %d markets with custom filters\n", len(customResp.Data))
		if customResp.Cursor != nil {
			fmt.Printf("Next page cursor: %s\n", *customResp.Cursor)
		}
		printMarkets(customResp.Data)
	} else {
		fmt.Println("No custom filters provided (use command line args or env vars)")
	}
}

// buildOptionsFromArgs builds GetMarketsOptions from command line arguments or environment variables
func buildOptionsFromArgs() *types.GetMarketsOptions {
	opts := &types.GetMarketsOptions{}

	// First (limit) from args or env
	if len(os.Args) > 1 {
		opts.First = &os.Args[1]
	}
	if opts.First == nil {
		if firstStr := os.Getenv("MARKET_FIRST"); firstStr != "" {
			opts.First = &firstStr
		}
	}

	// After (cursor) from args or env
	if len(os.Args) > 2 {
		opts.After = &os.Args[2]
	}
	if opts.After == nil {
		if afterStr := os.Getenv("MARKET_AFTER"); afterStr != "" {
			opts.After = &afterStr
		}
	}

	// Return nil if no options are set
	if opts.First == nil && opts.After == nil {
		return nil
	}

	return opts
}

// printMarkets prints market details with all fields clearly displayed
func printMarkets(markets []types.Market) {
	for i, market := range markets {
		fmt.Printf("\n--- Market %d ---\n", i+1)
		fmt.Printf("ID: %s\n", market.ID.String())
		fmt.Printf("Image URL: %s\n", market.ImageURL)
		fmt.Printf("Title: %s\n", market.Title)
		fmt.Printf("Question: %s\n", market.Question)
		fmt.Printf("Description: %s\n", market.Description)
		fmt.Printf("Status: %s\n", market.Status.String())
		fmt.Printf("Is Neg Risk: %v\n", market.IsNegRisk)
		fmt.Printf("Is Yield Bearing: %v\n", market.IsYieldBearing)
		fmt.Printf("Fee Rate Bps: %d\n", market.FeeRateBps)
		fmt.Printf("Oracle Question ID: %s\n", market.OracleQuestionID.Hex())
		fmt.Printf("Condition ID: %s\n", market.ConditionID.Hex())
		fmt.Printf("Resolver Address: %s\n", market.ResolverAddress.Hex())
		fmt.Printf("Category Slug: %s\n", market.CategorySlug)
		fmt.Printf("Created At: %s\n", market.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Decimal Precision: %d\n", market.DecimalPrecision)
		fmt.Printf("Spread Threshold: %f\n", market.SpreadThreshold)
		fmt.Printf("Share Threshold: %f\n", market.ShareThreshold)

		// Question Index (nullable)
		if market.QuestionIndex != nil {
			fmt.Printf("Question Index: %d\n", *market.QuestionIndex)
		} else {
			fmt.Printf("Question Index: (null)\n")
		}

		// Kalshi Market Ticker (nullable)
		if market.KalshiMarketTicker != nil {
			fmt.Printf("Kalshi Market Ticker: %s\n", *market.KalshiMarketTicker)
		} else {
			fmt.Printf("Kalshi Market Ticker: (null)\n")
		}

		// Polymarket Condition IDs
		fmt.Printf("Polymarket Condition IDs (%d):\n", len(market.PolymarketConditionIDs))
		if len(market.PolymarketConditionIDs) == 0 {
			fmt.Printf("  (none)\n")
		} else {
			for j, id := range market.PolymarketConditionIDs {
				fmt.Printf("  %d: %s\n", j+1, id.Hex())
			}
		}

		// Resolution (nullable)
		if market.Resolution != nil {
			fmt.Printf("Resolution:\n")
			fmt.Printf("  Name: %s\n", market.Resolution.Name)
			fmt.Printf("  Index Set: %d\n", market.Resolution.IndexSet)
			fmt.Printf("  On Chain ID: %s\n", market.Resolution.OnChainID)
			if market.Resolution.Status != nil {
				fmt.Printf("  Status: %s\n", *market.Resolution.Status)
			} else {
				fmt.Printf("  Status: (null)\n")
			}
		} else {
			fmt.Printf("Resolution: (null)\n")
		}

		// Outcomes
		fmt.Printf("Outcomes (%d):\n", len(market.Outcomes))
		if len(market.Outcomes) == 0 {
			fmt.Printf("  (no outcomes)\n")
		} else {
			for j, outcome := range market.Outcomes {
				fmt.Printf("  Outcome %d:\n", j+1)
				fmt.Printf("    Name: %s\n", outcome.Name)
				fmt.Printf("    Index Set: %d\n", outcome.IndexSet)
				fmt.Printf("    On Chain ID: %s\n", outcome.OnChainID)
				if outcome.Status != nil {
					fmt.Printf("    Status: %s\n", *outcome.Status)
				} else {
					fmt.Printf("    Status: (null)\n")
				}
			}
		}
	}
}
