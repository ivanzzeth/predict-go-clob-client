package main

import (
	"fmt"
	"log"
	"os"

	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
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

	// Create client with API key
	client, err := predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Get private key from environment (required for authentication)
	privateKey := os.Getenv("WALLET_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("WALLET_PRIVATE_KEY environment variable is required")
	}

	// Authenticate with private key
	token, address, err := client.Authenticate(privateKey)
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}
	client.SetJWTToken(token)
	fmt.Printf("Authenticated as: %s\n\n", address.Hex())

	// Get order hash from environment
	orderHash := os.Getenv("ORDER_HASH")
	if orderHash == "" {
		log.Fatal("ORDER_HASH environment variable is required (order hash, not order ID)")
	}

	// Get order by hash
	fmt.Printf("=== Getting order by hash: %s ===\n", orderHash)
	order, err := client.GetOrderByHash(orderHash)
	if err != nil {
		log.Fatalf("Failed to get order: %v", err)
	}

	// Print order details
	fmt.Printf("\nOrder Details:\n")
	fmt.Printf("  ID: %s\n", order.ID)
	fmt.Printf("  Market ID: %s\n", order.MarketID)
	fmt.Printf("  Currency: %s\n", order.Currency)
	fmt.Printf("  Strategy: %s\n", order.Strategy)
	fmt.Printf("  Status: %s\n", order.Status)
	fmt.Printf("  Amount: %s (raw: %s)\n", order.Amount.String(), order.RawAmount)
	fmt.Printf("  Amount Filled: %s (raw: %s)\n", order.AmountFilled.String(), order.RawAmountFilled)
	fmt.Printf("  Is Neg Risk: %v\n", order.IsNegRisk)
	fmt.Printf("  Is Yield Bearing: %v\n", order.IsYieldBearing)
	fmt.Printf("  Reward Earning Rate: %f\n", order.RewardEarningRate)

	// Print order data details
	if order.OrderData.Salt != "" {
		fmt.Printf("\nOrder Data:\n")
		fmt.Printf("  Hash: %s\n", order.OrderData.Hash.Hex())
		fmt.Printf("  Salt: %s\n", order.OrderData.Salt)
		fmt.Printf("  Maker: %s\n", order.OrderData.Maker.Hex())
		fmt.Printf("  Signer: %s\n", order.OrderData.Signer.Hex())
		fmt.Printf("  Taker: %s\n", order.OrderData.Taker.Hex())
		fmt.Printf("  Token ID: %s\n", order.OrderData.TokenID)
		fmt.Printf("  Maker Amount: %s\n", order.OrderData.MakerAmount.String())
		fmt.Printf("  Taker Amount: %s\n", order.OrderData.TakerAmount.String())
		expirationTime := order.OrderData.Expiration.Time()
		fmt.Printf("  Expiration: %s\n", expirationTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Nonce: %s\n", order.OrderData.Nonce)
		fmt.Printf("  Fee Rate Bps: %s\n", order.OrderData.FeeRateBps)
		fmt.Printf("  Side: %d\n", order.OrderData.Side)
		fmt.Printf("  Signature Type: %d\n", order.OrderData.SignatureType)
	}
}
