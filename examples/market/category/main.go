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

	// Get category slug from command line or environment
	var slug string
	if len(os.Args) > 1 {
		slug = os.Args[1]
	}
	if slug == "" {
		slug = os.Getenv("CATEGORY_SLUG")
	}
	if slug == "" {
		log.Fatal("Category slug is required (provide as command line argument or CATEGORY_SLUG env var)")
	}

	// Get category by slug
	fmt.Printf("=== Getting category by slug: %s ===\n", slug)
	category, err := client.GetCategory(slug)
	if err != nil {
		log.Fatalf("Failed to get category: %v", err)
	}

	// Print category details with all fields clearly displayed
	printCategory(category)
}

// printCategory prints category details with all fields clearly displayed
func printCategory(category *types.Category) {
	fmt.Printf("\n=== Category Details ===\n")
	fmt.Printf("ID: %s\n", category.ID.String())
	fmt.Printf("Slug: %s\n", category.Slug)
	fmt.Printf("Title: %s\n", category.Title)
	fmt.Printf("Description: %s\n", category.Description)
	fmt.Printf("Image URL: %s\n", category.ImageURL)
	fmt.Printf("Is Neg Risk: %v\n", category.IsNegRisk)
	fmt.Printf("Is Yield Bearing: %v\n", category.IsYieldBearing)
	fmt.Printf("Market Variant: %s\n", category.MarketVariant.String())
	fmt.Printf("Status: %s\n", category.Status.String())
	fmt.Printf("Created At: %s\n", category.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Published At: %s\n", category.PublishedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Starts At: %s\n", category.StartsAt.Format("2006-01-02 15:04:05"))
	if category.EndsAt != nil {
		fmt.Printf("Ends At: %s\n", category.EndsAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("Ends At: (null)\n")
	}

	// Print tags
	fmt.Printf("\n=== Tags (%d) ===\n", len(category.Tags))
	if len(category.Tags) == 0 {
		fmt.Println("  (no tags)")
	} else {
		for i, tag := range category.Tags {
			fmt.Printf("  Tag %d:\n", i+1)
			fmt.Printf("    ID: %s\n", tag.ID.String())
			fmt.Printf("    Name: %s\n", tag.Name)
		}
	}

	// Print markets
	fmt.Printf("\n=== Markets (%d) ===\n", len(category.Markets))
	if len(category.Markets) == 0 {
		fmt.Println("  (no markets)")
	} else {
		for i, market := range category.Markets {
			fmt.Printf("\n  --- Market %d ---\n", i+1)
			fmt.Printf("    ID: %s\n", market.ID.String())
			fmt.Printf("    Image URL: %s\n", market.ImageURL)
			fmt.Printf("    Title: %s\n", market.Title)
			fmt.Printf("    Question: %s\n", market.Question)
			fmt.Printf("    Description: %s\n", market.Description)
			fmt.Printf("    Status: %s\n", market.Status.String())
			fmt.Printf("    Is Neg Risk: %v\n", market.IsNegRisk)
			fmt.Printf("    Is Yield Bearing: %v\n", market.IsYieldBearing)
			fmt.Printf("    Fee Rate Bps: %d\n", market.FeeRateBps)
			fmt.Printf("    Oracle Question ID: %s\n", market.OracleQuestionID.Hex())
			fmt.Printf("    Condition ID: %s\n", market.ConditionID.Hex())
			fmt.Printf("    Resolver Address: %s\n", market.ResolverAddress.Hex())
			fmt.Printf("    Category Slug: %s\n", market.CategorySlug)
			fmt.Printf("    Created At: %s\n", market.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("    Decimal Precision: %d\n", market.DecimalPrecision)
			fmt.Printf("    Spread Threshold: %f\n", market.SpreadThreshold)
			fmt.Printf("    Share Threshold: %f\n", market.ShareThreshold)

			// Question Index (nullable)
			if market.QuestionIndex != nil {
				fmt.Printf("    Question Index: %d\n", *market.QuestionIndex)
			} else {
				fmt.Printf("    Question Index: (null)\n")
			}

			// Kalshi Market Ticker (nullable)
			if market.KalshiMarketTicker != nil {
				fmt.Printf("    Kalshi Market Ticker: %s\n", *market.KalshiMarketTicker)
			} else {
				fmt.Printf("    Kalshi Market Ticker: (null)\n")
			}

			// Polymarket Condition IDs
			fmt.Printf("    Polymarket Condition IDs (%d):\n", len(market.PolymarketConditionIDs))
			if len(market.PolymarketConditionIDs) == 0 {
				fmt.Printf("      (none)\n")
			} else {
			for j, id := range market.PolymarketConditionIDs {
				fmt.Printf("      %d: %s\n", j+1, id.Hex())
			}
			}

			// Resolution (nullable)
			if market.Resolution != nil {
				fmt.Printf("    Resolution:\n")
				fmt.Printf("      Name: %s\n", market.Resolution.Name)
				fmt.Printf("      Index Set: %d\n", market.Resolution.IndexSet)
				fmt.Printf("      On Chain ID: %s\n", market.Resolution.OnChainID)
				if market.Resolution.Status != nil {
					fmt.Printf("      Status: %s\n", *market.Resolution.Status)
				} else {
					fmt.Printf("      Status: (null)\n")
				}
			} else {
				fmt.Printf("    Resolution: (null)\n")
			}

			// Outcomes
			fmt.Printf("    Outcomes (%d):\n", len(market.Outcomes))
			if len(market.Outcomes) == 0 {
				fmt.Printf("      (no outcomes)\n")
			} else {
				for j, outcome := range market.Outcomes {
					fmt.Printf("      Outcome %d:\n", j+1)
					fmt.Printf("        Name: %s\n", outcome.Name)
					fmt.Printf("        Index Set: %d\n", outcome.IndexSet)
					fmt.Printf("        On Chain ID: %s\n", outcome.OnChainID)
					if outcome.Status != nil {
						fmt.Printf("        Status: %s\n", *outcome.Status)
					} else {
						fmt.Printf("        Status: (null)\n")
					}
				}
			}
		}
	}
}
