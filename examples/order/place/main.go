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

	priceStr := os.Getenv("PRICE")
	if priceStr == "" {
		log.Fatal("PRICE environment variable is required")
	}
	price, err := decimal.NewFromString(priceStr)
	if err != nil {
		log.Fatalf("Invalid PRICE: %v", err)
	}

	// Get order side (default to BUY)
	sideStr := os.Getenv("SIDE")
	var side predictclob.OrderSide
	if sideStr == "SELL" {
		side = predictclob.OrderSideSell
	} else {
		side = predictclob.OrderSideBuy
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Printf("EOA Address: %s", address.Hex())

	// Create EOA signer
	eoaSigner := ethsig.NewEthPrivateKeySigner(privateKey)

	// Get API key from environment
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create client
	client, err := predictclob.NewClient(
		predictclob.WithChainID(big.NewInt(56)),
		predictclob.WithAPIKey(apiKey),
		predictclob.WithEOATradingSigner(eoaSigner, address),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Authenticate
	jwtToken, _, err := client.Authenticate(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}
	client.SetJWTToken(jwtToken)
	log.Printf("Authenticated successfully\n")

	// Place limit order
	log.Println("=== Placing order ===")
	fmt.Printf("Market ID: %s\n", marketID)
	fmt.Printf("Token ID: %s\n", tokenID)
	fmt.Printf("Side: %s\n", side)
	fmt.Printf("Amount: %s shares\n", amount.String())
	fmt.Printf("Price: %s\n", price.String())

	result, err := client.PlaceOrder(&predictclob.PlaceOrderInput{
		MarketID: types.MarketID(marketID),
		TokenID:  types.TokenID(tokenID),
		Side:     side,
		Strategy: predictclob.OrderStrategyLimit,
		Amount:   amount,
		Price:    price,
	})
	if err != nil {
		log.Fatalf("Failed to place order: %v", err)
	}

	fmt.Printf("\nOrder placed successfully!\n")
	fmt.Printf("Order ID: %s\n", result.OrderID)
	fmt.Printf("Order Hash: %s\n", result.OrderHash)
}
