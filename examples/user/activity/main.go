package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethsig"
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
	if apiKey == "" {
		log.Fatal("PREDICT_API_KEY environment variable is required")
	}

	// Get private key from environment (required for authentication)
	privateKeyHex := os.Getenv("WALLET_PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("WALLET_PRIVATE_KEY environment variable is required")
	}

	// Parse private key and create signer
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}
	signer := ethsig.NewEthPrivateKeySigner(privateKey)
	address := signer.GetAddress()

	// Create client with signer (auto-authenticates if Signer, APIKey are set and JWTToken is not)
	client, err := predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
		predictclob.WithEOATradingSigner(signer),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Client created and authenticated as: %s\n\n", address.Hex())

	// Example 1: Get all activity (first 10)
	fmt.Println("=== Example 1: Get account activity (first 10) ===")
	first := 10
	opts := &types.GetActivityOptions{
		First: &first,
	}
	response, err := client.GetActivity(opts)
	if err != nil {
		log.Fatalf("Failed to get activity: %v", err)
	}

	fmt.Printf("Found %d activity items\n", len(response.Data))
	if response.Cursor != nil {
		fmt.Printf("Next page cursor: %s\n", *response.Cursor)
	}

	// Print activity items
	for i, activity := range response.Data {
		fmt.Printf("\n--- Activity %d ---\n", i+1)
		fmt.Printf("Type: %s\n", activity.Name)
		fmt.Printf("Created At: %s\n", activity.CreatedAt.Format("2006-01-02 15:04:05"))
		if activity.TransactionHash != nil {
			fmt.Printf("Transaction Hash: %s\n", activity.TransactionHash.Hex())
		}
		if activity.RawAmountFilled != nil {
			fmt.Printf("Amount Filled: %s (raw: %s)\n", activity.AmountFilled.String(), *activity.RawAmountFilled)
		}
		if activity.RawPriceExecuted != nil {
			fmt.Printf("Price Executed: %s (raw: %s)\n", activity.PriceExecuted.String(), *activity.RawPriceExecuted)
		}
		if activity.Order != nil {
			fmt.Printf("Order: QuoteType=%s, Amount=%s, Price=%s\n",
				activity.Order.QuoteType, activity.Order.Amount.String(), activity.Order.Price.String())
		}
		fmt.Printf("Market: %s (ID: %s)\n", activity.Market.Title, activity.Market.ID)
		if activity.Outcome != nil {
			fmt.Printf("Outcome: %s (IndexSet: %d)\n", activity.Outcome.Name, activity.Outcome.IndexSet)
		}
	}

	// Example 2: Get next page if cursor exists
	if response.Cursor != nil && *response.Cursor != "" {
		fmt.Println("\n=== Example 2: Get next page ===")
		nextOpts := &types.GetActivityOptions{
			First: &first,
			After: *response.Cursor,
		}
		nextResponse, err := client.GetActivity(nextOpts)
		if err != nil {
			log.Fatalf("Failed to get next page: %v", err)
		}
		fmt.Printf("Found %d more activity items\n", len(nextResponse.Data))
	}
}
