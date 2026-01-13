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

	// Get market by ID
	// useCache=false for detail view as we want real-time data
	fmt.Printf("=== Getting market by ID: %s ===\n", marketID.String())
	market, err := client.GetMarket(marketID, false)
	if err != nil {
		log.Fatalf("Failed to get market: %v", err)
	}

	// Print market details with all fields clearly displayed
	printMarket(market)
}

// printMarket prints market details with all fields clearly displayed
func printMarket(market *types.Market) {
	fmt.Printf("\n=== Market Details ===\n")
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
	fmt.Printf("\n=== Polymarket Condition IDs (%d) ===\n", len(market.PolymarketConditionIDs))
	if len(market.PolymarketConditionIDs) == 0 {
		fmt.Println("  (none)")
	} else {
		for i, id := range market.PolymarketConditionIDs {
			fmt.Printf("  %d: %s\n", i+1, id.Hex())
		}
	}

	// Resolution (nullable)
	fmt.Printf("\n=== Resolution ===\n")
	if market.Resolution != nil {
		fmt.Printf("  Name: %s\n", market.Resolution.Name)
		fmt.Printf("  Index Set: %d\n", market.Resolution.IndexSet)
		fmt.Printf("  On Chain ID: %s\n", market.Resolution.OnChainID)
		if market.Resolution.Status != nil {
			fmt.Printf("  Status: %s\n", *market.Resolution.Status)
		} else {
			fmt.Printf("  Status: (null)\n")
		}
	} else {
		fmt.Println("  (null)")
	}

	// Outcomes
	fmt.Printf("\n=== Outcomes (%d) ===\n", len(market.Outcomes))
	if len(market.Outcomes) == 0 {
		fmt.Println("  (no outcomes)")
	} else {
		for i, outcome := range market.Outcomes {
			fmt.Printf("  Outcome %d:\n", i+1)
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
