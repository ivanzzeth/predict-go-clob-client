package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
)

func main() {
	// Load .env file from project root
	_ = godotenv.Load(".env")

	// Get configuration from environment
	apiKey := os.Getenv("PREDICT_API_KEY")
	privateKey := os.Getenv("WALLET_PRIVATE_KEY")

	// Create client
	client, err := predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Predict.fun SDK Demo ===\n")

	// 1. Check API health (may not be available)
	fmt.Println("1. Checking API health...")
	ok, err := client.GetOk()
	if err != nil {
		fmt.Printf("   ⚠ API health check failed (endpoint may not exist): %v\n", err)
		fmt.Println("   Continuing with other API calls...\n")
	} else {
		fmt.Printf("   ✓ API is healthy: %v\n\n", ok)
	}

	// 2. Get categories
	fmt.Println("2. Fetching categories...")
	categories, err := client.GetCategories(nil)
	if err != nil {
		log.Fatalf("Failed to get categories: %v", err)
	}
	fmt.Printf("   ✓ Found %d categories\n", len(categories))
	if len(categories) > 0 {
		fmt.Printf("   Example: %s (slug: %s)\n\n", categories[0].Title, categories[0].Slug)
	}

	// 3. Get markets (prefer OPEN status)
	fmt.Println("3. Fetching markets...")
	markets, err := client.GetMarkets(&types.GetMarketsOptions{
		Limit:  10,
		Status: types.MarketStatusOpen,
	})
	if err != nil {
		log.Fatalf("Failed to get markets: %v", err)
	}
	fmt.Printf("   ✓ Found %d OPEN markets\n", len(markets))
	if len(markets) > 0 {
		marketID := markets[0].ID
		fmt.Printf("   Example: %s (ID: %s, Status: %s)\n\n", markets[0].Title, marketID, markets[0].Status)

		// 4. Get market details
		fmt.Println("4. Fetching market details...")
		market, err := client.GetMarket(marketID)
		if err != nil {
			log.Fatalf("Failed to get market: %v", err)
		}
		fmt.Printf("   ✓ Market: %s\n", market.Title)
		fmt.Printf("   Question: %s\n", market.Question)
		fmt.Printf("   Status: %s\n", market.Status)
		fmt.Printf("   Fee Rate BPS: %s\n\n", market.FeeRateBps)

		// 5. Get market stats
		fmt.Println("5. Fetching market stats...")
		stats, err := client.GetMarketStats(marketID)
		if err != nil {
			log.Fatalf("Failed to get market stats: %v", err)
		}
		fmt.Printf("   ✓ Volume: %s\n", stats.Volume)
		fmt.Printf("   Open Interest: %s\n", stats.OpenInterest)
		fmt.Printf("   Last Price: %s\n\n", stats.LastPrice)

		// 6. Get orderbook
		fmt.Println("6. Fetching orderbook...")
		orderbook, err := client.GetMarketOrderbook(marketID)
		if err != nil {
			log.Fatalf("Failed to get orderbook: %v", err)
		}
		fmt.Printf("   ✓ Bids: %d levels\n", len(orderbook.Bids))
		fmt.Printf("   ✓ Asks: %d levels\n\n", len(orderbook.Asks))
	}

	// 7. Authentication (if private key is provided)
	if privateKey != "" {
		fmt.Println("7. Authenticating...")
		token, address, err := client.Authenticate(privateKey)
		if err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}
		fmt.Printf("   ✓ Authenticated as: %s\n", address.Hex())
		fmt.Printf("   ✓ JWT Token: %s...\n\n", token[:50])

		// 8. Get account info
		fmt.Println("8. Fetching account info...")
		account, err := client.GetAccount()
		if err != nil {
			log.Fatalf("Failed to get account: %v", err)
		}
		fmt.Printf("   ✓ Name: %s\n", account.Name)
		fmt.Printf("   Address: %s\n", account.Address)
		if account.Referral != nil {
			fmt.Printf("   Referral Code: %s\n", account.Referral.Code)
		}
		fmt.Println()

		// 9. Get positions
		fmt.Println("9. Fetching positions...")
		positions, err := client.GetPositions(&types.GetPositionsOptions{
			First: 10,
		})
		if err != nil {
			log.Fatalf("Failed to get positions: %v", err)
		}
		fmt.Printf("   ✓ Found %d positions\n", len(positions))
		for i, pos := range positions {
			if i >= 3 {
				fmt.Printf("   ... and %d more\n", len(positions)-3)
				break
			}
			fmt.Printf("   - Position %d: Market=%s, Amount=%s\n", i+1, pos.Market.Title, pos.Amount)
		}
	} else {
		fmt.Println("7. Skipping authentication (WALLET_PRIVATE_KEY not set)")
		fmt.Println("   Set WALLET_PRIVATE_KEY to test authenticated endpoints")
	}

	fmt.Println("\n=== Demo Complete ===")
}
