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
	if apiKey == "" {
		log.Fatal("PREDICT_API_KEY environment variable is required")
	}

	// Create read-only client with API key
	client := predictclob.NewReadOnlyClient(constants.DefaultAPIHost, apiKey)

	// Get OPEN categories
	fmt.Println("=== Finding OPEN Markets ===")
	fmt.Println("Querying OPEN categories...")
	fmt.Println()

	openOpts := &types.GetCategoriesOptions{
		Status: types.CategoryStatusOpen,
	}

	// Get first page of OPEN categories (max 150 per API limit)
	first := 150
	openOpts.First = &first

	categoriesResp, err := client.GetCategories(openOpts)
	if err != nil {
		log.Fatalf("Failed to get OPEN categories: %v", err)
	}

	fmt.Printf("Found %d OPEN categories\n\n", len(categoriesResp.Data))

	// Track found markets
	var openMarkets []struct {
		CategorySlug string
		Market       types.CategoryMarket
	}

	// Iterate through categories and their markets
	for _, category := range categoriesResp.Data {
		for _, market := range category.Markets {
			// Check if market status is OPEN (REGISTERED means it's open for trading)
			if market.Status == types.MarketStatusRegistered || market.Status == types.MarketStatusOpen {
				openMarkets = append(openMarkets, struct {
					CategorySlug string
					Market       types.CategoryMarket
				}{
					CategorySlug: category.Slug,
					Market:       market,
				})
			}
		}
	}

	// Print found OPEN markets
	fmt.Printf("=== Found %d OPEN Markets ===\n\n", len(openMarkets))
	for i, item := range openMarkets {
		fmt.Printf("--- Market %d ---\n", i+1)
		fmt.Printf("Market ID: %s\n", item.Market.ID.String())
		fmt.Printf("Title: %s\n", item.Market.Title)
		fmt.Printf("Question: %s\n", item.Market.Question)
		fmt.Printf("Category Slug: %s\n", item.CategorySlug)
		fmt.Printf("Status: %s\n", item.Market.Status.String())
		fmt.Printf("Is Neg Risk: %v\n", item.Market.IsNegRisk)
		fmt.Printf("Is Yield Bearing: %v\n", item.Market.IsYieldBearing)
		fmt.Printf("Condition ID: %s\n", item.Market.ConditionID.Hex())
		fmt.Printf("Fee Rate Bps: %d\n", item.Market.FeeRateBps)
		fmt.Printf("Created At: %s\n", item.Market.CreatedAt.Format("2006-01-02 15:04:05"))
		if item.Market.Resolution != nil {
			fmt.Printf("Resolution: %s (IndexSet: %d)\n", item.Market.Resolution.Name, item.Market.Resolution.IndexSet)
		}
		fmt.Printf("Outcomes (%d):\n", len(item.Market.Outcomes))
		for j, outcome := range item.Market.Outcomes {
			fmt.Printf("  %d. %s (IndexSet: %d, OnChainID: %s)\n",
				j+1, outcome.Name, outcome.IndexSet, outcome.OnChainID)
		}
		fmt.Println()
	}

	// Print summary by type
	fmt.Println("=== Summary by Type ===")
	negRiskYB := 0
	negRiskNonYB := 0
	nonNegRiskYB := 0
	nonNegRiskNonYB := 0

	for _, item := range openMarkets {
		if item.Market.IsNegRisk && item.Market.IsYieldBearing {
			negRiskYB++
		} else if item.Market.IsNegRisk && !item.Market.IsYieldBearing {
			negRiskNonYB++
		} else if !item.Market.IsNegRisk && item.Market.IsYieldBearing {
			nonNegRiskYB++
		} else {
			nonNegRiskNonYB++
		}
	}

	fmt.Printf("NegRisk + YB: %d markets\n", negRiskYB)
	fmt.Printf("NegRisk + Non-YB: %d markets\n", negRiskNonYB)
	fmt.Printf("Non-NegRisk + YB: %d markets\n", nonNegRiskYB)
	fmt.Printf("Non-NegRisk + Non-YB: %d markets\n", nonNegRiskNonYB)
}
