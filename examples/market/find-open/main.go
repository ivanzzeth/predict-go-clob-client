package main

import (
	"fmt"
	"log"
	"os"

	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
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

	// Track found markets with price information
	type MarketWithPrice struct {
		CategorySlug string
		Market       types.CategoryMarket
		MidPrice     *decimal.Decimal // nil if price cannot be determined
		BestBid      *decimal.Decimal
		BestAsk      *decimal.Decimal
	}

	var openMarkets []MarketWithPrice

	// Price range filter: 0.3 to 0.7
	minPrice := decimal.NewFromFloat(0.3)
	maxPrice := decimal.NewFromFloat(0.7)

	fmt.Println("Filtering markets by price range [0.3, 0.7]...")
	fmt.Println()

	// Iterate through categories and their markets
	for _, category := range categoriesResp.Data {
		for _, market := range category.Markets {
			// Check if market status is OPEN (REGISTERED means it's open for trading)
			if market.Status == types.MarketStatusRegistered || market.Status == types.MarketStatusOpen {
				// Get orderbook to check price
				orderbook, err := client.GetMarketOrderbook(market.ID)
				if err != nil {
					log.Printf("Warning: Failed to get orderbook for market %s: %v", market.ID.String(), err)
					continue
				}

				// Calculate mid price if both bid and ask exist
				var midPrice *decimal.Decimal
				var bestBid, bestAsk *decimal.Decimal

				if len(orderbook.Bids) > 0 && len(orderbook.Asks) > 0 {
					bid := orderbook.BestBid
					ask := orderbook.BestAsk
					bestBid = &bid
					bestAsk = &ask
					mid := bid.Add(ask).Div(decimal.NewFromInt(2))
					midPrice = &mid

					// Check if mid price is in range [0.3, 0.7]
					if mid.GreaterThanOrEqual(minPrice) && mid.LessThanOrEqual(maxPrice) {
						openMarkets = append(openMarkets, MarketWithPrice{
							CategorySlug: category.Slug,
							Market:       market,
							MidPrice:     midPrice,
							BestBid:      bestBid,
							BestAsk:      bestAsk,
						})
					}
				} else {
					// No price data available, skip this market
					log.Printf("Warning: Market %s has no bid/ask data", market.ID.String())
				}
			}
		}
	}

	// Print found OPEN markets with price in range [0.3, 0.7]
	fmt.Printf("=== Found %d OPEN Markets with Price in [0.3, 0.7] ===\n\n", len(openMarkets))
	for i, item := range openMarkets {
		fmt.Printf("--- Market %d ---\n", i+1)
		fmt.Printf("Market ID: %s\n", item.Market.ID.String())
		fmt.Printf("Title: %s\n", item.Market.Title)
		fmt.Printf("Question: %s\n", item.Market.Question)
		fmt.Printf("Category Slug: %s\n", item.CategorySlug)
		fmt.Printf("Status: %s\n", item.Market.Status.String())
		if item.MidPrice != nil {
			fmt.Printf("Mid Price: %s\n", item.MidPrice.String())
		}
		if item.BestBid != nil {
			fmt.Printf("Best Bid: %s\n", item.BestBid.String())
		}
		if item.BestAsk != nil {
			fmt.Printf("Best Ask: %s\n", item.BestAsk.String())
		}
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
