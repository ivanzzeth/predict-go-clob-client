package main

import (
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethsig"
	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/types"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

func main() {
	// Load .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get required environment variables
	privateKeyHex := os.Getenv("WALLET_PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("WALLET_PRIVATE_KEY environment variable is required")
	}

	marketID := os.Getenv("MARKET_ID")
	if marketID == "" {
		log.Fatal("MARKET_ID environment variable is required")
	}

	tokenID := os.Getenv("TOKEN_ID")
	if tokenID == "" {
		log.Fatal("TOKEN_ID environment variable is required")
	}

	amountStr := os.Getenv("AMOUNT")
	if amountStr == "" {
		log.Fatal("AMOUNT environment variable is required")
	}
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		log.Fatalf("Invalid AMOUNT: %v", err)
	}

	// Strategy (default LIMIT)
	strategyStr := os.Getenv("STRATEGY")
	if strategyStr == "" {
		strategyStr = "LIMIT"
	}
	var strategy types.OrderStrategy
	switch strategyStr {
	case "MARKET":
		strategy = types.OrderStrategyMarket
	case "LIMIT":
		strategy = types.OrderStrategyLimit
	default:
		log.Fatalf("Invalid STRATEGY: %s (expected LIMIT or MARKET)", strategyStr)
	}

	// Price: required for LIMIT; optional for MARKET (can be 0)
	price := decimal.Zero
	priceStr := os.Getenv("PRICE")
	if strategy == types.OrderStrategyLimit {
		if priceStr == "" {
			log.Fatal("PRICE environment variable is required for LIMIT orders")
		}
		var err error
		price, err = decimal.NewFromString(priceStr)
		if err != nil {
			log.Fatalf("Invalid PRICE: %v", err)
		}
	} else {
		// MARKET: allow empty -> 0
		if priceStr != "" {
			var err error
			price, err = decimal.NewFromString(priceStr)
			if err != nil {
				log.Fatalf("Invalid PRICE: %v", err)
			}
		}
	}

	// Get order side (default to BUY)
	sideStr := os.Getenv("SIDE")
	var side types.OrderSide
	if sideStr == "SELL" {
		side = types.OrderSideSell
	} else {
		side = types.OrderSideBuy
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// Create EOA signer
	eoaSigner := ethsig.NewEthPrivateKeySigner(privateKey)
	address := eoaSigner.GetAddress()
	log.Printf("EOA Address: %s", address.Hex())

	// Get API key from environment
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create client (auto-authenticates if Signer, APIKey are set and JWTToken is not)
	client, err := predictclob.NewClient(
		predictclob.WithChainID(big.NewInt(56)),
		predictclob.WithAPIKey(apiKey),
		predictclob.WithEOATradingSigner(eoaSigner),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	log.Printf("Client created and authenticated successfully\n")

	// Place limit order
	log.Println("=== Placing order ===")
	fmt.Printf("Market ID: %s\n", marketID)
	fmt.Printf("Token ID: %s\n", tokenID)
	sideDisplayStr := "BUY"
	if side == types.OrderSideSell {
		sideDisplayStr = "SELL"
	}
	fmt.Printf("Side: %s\n", sideDisplayStr)
	fmt.Printf("Amount: %s shares\n", amount.String())
	fmt.Printf("Price: %s\n", price.String())
	fmt.Printf("Strategy: %s\n", string(strategy))

	// SlippageBps: used for MARKET orders (basis points)
	slippageBps := 0
	if strategy == types.OrderStrategyMarket {
		slippageStr := os.Getenv("SLIPPAGE_BPS")
		if slippageStr == "" {
			slippageStr = "1000" // default 10%
		}
		slippageDec, err := decimal.NewFromString(slippageStr)
		if err != nil {
			log.Fatalf("Invalid SLIPPAGE_BPS: %v", err)
		}
		slippageBps64 := slippageDec.IntPart()
		if slippageBps64 < 0 || slippageBps64 > 10000 {
			log.Fatalf("Invalid SLIPPAGE_BPS: %d (expected 0..10000)", slippageBps64)
		}
		slippageBps = int(slippageBps64)
		fmt.Printf("SlippageBps: %d\n", slippageBps)
	}

	result, err := client.PlaceOrder(&types.PlaceOrderInput{
		MarketID:    types.MustMarketIDFromString(marketID),
		TokenID:     types.TokenID(tokenID),
		Side:        side,
		Strategy:    strategy,
		Amount:      amount,
		Price:       price,
		SlippageBps: slippageBps,
	})
	if err != nil {
		log.Fatalf("Failed to place order: %v", err)
	}

	fmt.Printf("\nOrder placed successfully!\n")
	fmt.Printf("Order ID: %s\n", result.OrderID)
	fmt.Printf("Order Hash: %s\n", result.OrderHash)
}
