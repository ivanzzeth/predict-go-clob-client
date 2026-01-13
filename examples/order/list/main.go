package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethsig"
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
	fmt.Printf("Client created and authenticated as: %s\n\n", address.Hex())

	// Example 1: Get all orders
	fmt.Println("=== Example 1: Get all orders ===")
	allOrdersResp, err := client.GetOrders(nil)
	if err != nil {
		log.Fatalf("Failed to get orders: %v", err)
	}
	fmt.Printf("Found %d orders\n", len(allOrdersResp.Data))
	if allOrdersResp.Cursor != nil {
		fmt.Printf("Next page cursor: %s\n", *allOrdersResp.Cursor)
	}
	for i, order := range allOrdersResp.Data {
		fmt.Printf("  Order %d: ID=%s, Status=%s, MarketID=%s, Strategy=%s, Amount=%s (raw: %s)\n",
			i+1, order.ID, order.Status, order.MarketID.String(), order.Strategy, order.Amount.String(), order.RawAmount)
	}

	// Example 2: Get orders by market ID
	fmt.Println("\n=== Example 2: Get orders by market ID ===")
	marketIDStr := os.Getenv("MARKET_ID")
	if marketIDStr != "" {
		marketID := types.MustMarketIDFromString(marketIDStr)
		opts := &types.GetOrdersOptions{
			MarketID: marketID,
		}
		marketOrdersResp, err := client.GetOrders(opts)
		if err != nil {
			log.Fatalf("Failed to get orders by market: %v", err)
		}
		fmt.Printf("Found %d orders for market %s\n", len(marketOrdersResp.Data), marketIDStr)
		for i, order := range marketOrdersResp.Data {
			fmt.Printf("  Order %d: ID=%s, Status=%s, MarketID=%s, Strategy=%s, Amount=%s\n",
				i+1, order.ID, order.Status, order.MarketID.String(), order.Strategy, order.Amount.String())
		}
	} else {
		fmt.Println("MARKET_ID not set, skipping market filter example")
	}

	// Example 3: Get orders by status (OPEN)
	fmt.Println("\n=== Example 3: Get orders by status (OPEN) ===")
	statusOpts := &types.GetOrdersOptions{
		Status: types.OrderStatusOpen,
	}
	openOrdersResp, err := client.GetOrders(statusOpts)
	if err != nil {
		log.Fatalf("Failed to get open orders: %v", err)
	}
	fmt.Printf("Found %d open orders\n", len(openOrdersResp.Data))
	for i, order := range openOrdersResp.Data {
		fmt.Printf("  Order %d: ID=%s, MarketID=%s, Strategy=%s, Amount=%s, AmountFilled=%s\n",
			i+1, order.ID, order.MarketID.String(), order.Strategy, order.Amount.String(), order.AmountFilled.String())
	}

	// Example 3b: Get orders by status (FILLED)
	fmt.Println("\n=== Example 3b: Get orders by status (FILLED) ===")
	filledOpts := &types.GetOrdersOptions{
		Status: types.OrderStatusFilled,
	}
	filledOrdersResp, err := client.GetOrders(filledOpts)
	if err != nil {
		log.Fatalf("Failed to get filled orders: %v", err)
	}
	fmt.Printf("Found %d filled orders\n", len(filledOrdersResp.Data))
	for i, order := range filledOrdersResp.Data {
		fmt.Printf("  Order %d: ID=%s, MarketID=%s, Strategy=%s, Amount=%s, AmountFilled=%s\n",
			i+1, order.ID, order.MarketID.String(), order.Strategy, order.Amount.String(), order.AmountFilled.String())
		fmt.Printf("    OrderData: Hash=%s, Maker=%s, Signer=%s, Taker=%s\n",
			order.OrderData.Hash.Hex(), order.OrderData.Maker.Hex(), order.OrderData.Signer.Hex(), order.OrderData.Taker.Hex())
		fmt.Printf("    OrderData: TokenID=%s, Side=%d, MakerAmount=%s, TakerAmount=%s\n",
			order.OrderData.TokenID, order.OrderData.Side, order.OrderData.MakerAmount.String(), order.OrderData.TakerAmount.String())
		fmt.Printf("    OrderData: Expiration=%s\n", order.OrderData.Expiration.Time().Format("2006-01-02 15:04:05"))
	}

	// Example 4: Get orders by market ID and status
	fmt.Println("\n=== Example 4: Get orders by market ID and status ===")
	if marketIDStr != "" {
		marketID := types.MustMarketIDFromString(marketIDStr)
		combinedOpts := &types.GetOrdersOptions{
			MarketID: marketID,
			Status:   types.OrderStatusOpen,
		}
		combinedOrdersResp, err := client.GetOrders(combinedOpts)
		if err != nil {
			log.Fatalf("Failed to get orders: %v", err)
		}
		fmt.Printf("Found %d open orders for market %s\n", len(combinedOrdersResp.Data), marketIDStr)
		for i, order := range combinedOrdersResp.Data {
			fmt.Printf("  Order %d: ID=%s, Status=%s, MarketID=%s, Strategy=%s, Amount=%s\n",
				i+1, order.ID, order.Status, order.MarketID.String(), order.Strategy, order.Amount.String())
		}
	} else {
		fmt.Println("MARKET_ID not set, skipping combined filter example")
	}
}
