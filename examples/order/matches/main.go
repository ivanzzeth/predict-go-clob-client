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

	// Create client with API key
	client, err := predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
	)
	if err != nil {
		log.Fatal(err)
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
	client, err = predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
		predictclob.WithEOATradingSigner(signer),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Client created and authenticated as: %s\n", address.Hex())

	// Get matches with optional filters
	first := 10
	opts := &types.GetOrderMatchesOptions{
		First: &first, // Get first 10 matches
		// CategoryID:     types.CategoryID("your-category-id"),
		// MarketID:       types.MarketID("your-market-id"),
		SignerAddress: address.Hex(),
		// IsSignerMaker:  &[]bool{true}[0], // Filter for matches where signer is maker
	}

	response, err := client.GetOrderMatches(opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d match events\n", len(response.Data))
	if response.Cursor != nil {
		fmt.Printf("Next page cursor: %s\n", *response.Cursor)
	}

	// Print match events
	for i, match := range response.Data {
		fmt.Printf("\n--- Match Event %d ---\n", i+1)
		fmt.Printf("Market: %s (ID: %s)\n", match.Market.Title, match.Market.ID)
		fmt.Printf("Transaction Hash: %s\n", match.TransactionHash.Hex())
		fmt.Printf("Executed At: %s\n", match.ExecutedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Amount Filled: %s\n", match.AmountFilled.String())
		fmt.Printf("Price Executed: %s\n", match.PriceExecuted.String())
		fmt.Printf("Taker Quote Type: %s, Amount: %s, Price: %s\n",
			match.Taker.QuoteType, match.Taker.Amount.String(), match.Taker.Price.String())
		fmt.Printf("Number of Makers: %d\n", len(match.Makers))
		for j, maker := range match.Makers {
			fmt.Printf("  Maker %d: Quote Type: %s, Amount: %s, Price: %s\n",
				j+1, maker.QuoteType, maker.Amount.String(), maker.Price.String())
		}
	}
}
