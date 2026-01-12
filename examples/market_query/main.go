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
	// Load .env file from project root
	_ = godotenv.Load(".env")

	// Get API key from environment (optional for some endpoints)
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create client with optional API key
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
		// Use read-only client if no API key
		client = predictclob.NewReadOnlyClient(constants.DefaultAPIHost)
	}

	// Check API health (may require API key)
	ok, err := client.GetOk()
	if err != nil {
		fmt.Printf("Warning: API health check failed (may require API key): %v\n", err)
		fmt.Println("Continuing with other API calls...")
	} else {
		fmt.Printf("API is healthy: %v\n\n", ok)
	}

	// Get categories
	fmt.Println("=== Fetching Categories ===")
	categories, err := client.GetCategories(nil)
	if err != nil {
		log.Fatalf("Failed to get categories: %v", err)
	}
	fmt.Printf("Found %d categories\n", len(categories))
	for i, cat := range categories {
		if i >= 3 {
			fmt.Printf("  ... and %d more\n", len(categories)-3)
			break
		}
		fmt.Printf("  - %s (slug: %s, status: %s)\n", cat.Title, cat.Slug, cat.Status)
	}

	// Get OPEN categories
	fmt.Println("\n=== Fetching OPEN Categories ===")
	openCategories, err := client.GetCategories(&types.GetCategoriesOptions{
		Status: types.CategoryStatusOpen,
	})
	if err != nil {
		log.Fatalf("Failed to get OPEN categories: %v", err)
	}
	fmt.Printf("Found %d OPEN categories\n", len(openCategories))

	// Get markets (prefer OPEN status)
	fmt.Println("\n=== Fetching Markets ===")
	markets, err := client.GetMarkets(&types.GetMarketsOptions{
		Limit:  20,
		Status: types.MarketStatusOpen,
	})
	if err != nil {
		log.Fatalf("Failed to get markets: %v", err)
	}
	fmt.Printf("Found %d OPEN markets\n", len(markets))
	for i, market := range markets {
		if i >= 3 {
			fmt.Printf("  ... and %d more\n", len(markets)-3)
			break
		}
		fmt.Printf("  - %s (ID: %s, Status: %s)\n", market.Title, market.ID, market.Status)
	}

	// Get a specific market for detailed queries
	var testMarketID types.MarketID
	if len(markets) > 0 {
		testMarketID = markets[0].ID
	}

	if testMarketID != "" {
		fmt.Printf("\n=== Fetching Market Details: %s ===\n", testMarketID)

		market, err := client.GetMarket(testMarketID)
		if err != nil {
			log.Fatalf("Failed to get market: %v", err)
		}
		fmt.Printf("Market: %s\n", market.Title)
		fmt.Printf("Question: %s\n", market.Question)
		fmt.Printf("Status: %s\n", market.Status)
		fmt.Printf("Fee Rate BPS: %s\n", market.FeeRateBps)

		// Get market stats
		fmt.Printf("\n=== Fetching Market Stats ===\n")
		stats, err := client.GetMarketStats(testMarketID)
		if err != nil {
			log.Fatalf("Failed to get market stats: %v", err)
		}
		fmt.Printf("Volume: %s\n", stats.Volume)
		fmt.Printf("Open Interest: %s\n", stats.OpenInterest)
		fmt.Printf("Bid Price: %s\n", stats.BidPrice)
		fmt.Printf("Ask Price: %s\n", stats.AskPrice)
		fmt.Printf("Last Price: %s\n", stats.LastPrice)

		// Get orderbook
		fmt.Printf("\n=== Fetching Orderbook ===\n")
		orderbook, err := client.GetMarketOrderbook(testMarketID)
		if err != nil {
			log.Fatalf("Failed to get orderbook: %v", err)
		}
		fmt.Printf("Bids: %d levels\n", len(orderbook.Bids))
		if len(orderbook.Bids) > 0 {
			fmt.Printf("  Top bid: Price=%s, Amount=%s\n", orderbook.Bids[0].Price, orderbook.Bids[0].Amount)
		}
		fmt.Printf("Asks: %d levels\n", len(orderbook.Asks))
		if len(orderbook.Asks) > 0 {
			fmt.Printf("  Top ask: Price=%s, Amount=%s\n", orderbook.Asks[0].Price, orderbook.Asks[0].Amount)
		}

		// Get last sale (may not exist for new markets)
		fmt.Printf("\n=== Fetching Last Sale ===\n")
		sale, err := client.GetMarketLastSale(testMarketID)
		if err != nil {
			fmt.Printf("   ⚠ Last sale not available (market may have no trades yet): %v\n", err)
		} else {
			fmt.Printf("Price: %s\n", sale.Price)
			fmt.Printf("Amount: %s\n", sale.Amount)
			fmt.Printf("Transaction Hash: %s\n", sale.TransactionHash)
		}
	}
}
