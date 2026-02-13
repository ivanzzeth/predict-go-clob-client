package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file from project root directory
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get API key from environment (optional for this endpoint)
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create read-only client
	client := predictclob.NewReadOnlyClient(constants.DefaultAPIHost, apiKey)

	// Get address from command line or environment
	var address common.Address
	if len(os.Args) > 1 {
		address = common.HexToAddress(os.Args[1])
	}
	if address == (common.Address{}) {
		if addrStr := os.Getenv("PREDICT_ADDRESS"); addrStr != "" {
			address = common.HexToAddress(addrStr)
		}
	}
	if address == (common.Address{}) {
		log.Fatal("Address is required (provide as command line argument or PREDICT_ADDRESS env var)")
	}

	// Query leaderboard stats
	account, err := client.GetLeaderboardUserStats(address)
	if err != nil {
		log.Fatalf("Error getting leaderboard stats: %v", err)
	}

	// Print result
	fmt.Println("=== Leaderboard User Stats ===")
	fmt.Printf("Name:                    %s\n", account.Name)
	fmt.Printf("Address:                 %s\n", account.Address.Hex())
	fmt.Printf("Image Status:            %s\n", account.ImageStatus)
	if account.Leaderboard != nil {
		fmt.Printf("Allocation Round Points: %.4f\n", account.Leaderboard.AllocationRoundPoints)
		fmt.Printf("Total Points:            %.4f\n", account.Leaderboard.TotalPoints)
		fmt.Printf("Rank:                    %d\n", account.Leaderboard.Rank)
	} else {
		fmt.Println("Leaderboard:             (no data)")
	}
	fmt.Println("==============================")
}
