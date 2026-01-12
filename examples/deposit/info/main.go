package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
)

func main() {
	// Load .env file from project root directory
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get private key from environment
	privateKey := os.Getenv("WALLET_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("WALLET_PRIVATE_KEY environment variable is required")
	}

	// Get chain ID from environment (default: BNB Mainnet)
	chainID := int64(constants.ChainIDBNBMainnet)
	if chainIDStr := os.Getenv("CHAIN_ID"); chainIDStr != "" {
		if id, err := strconv.ParseInt(chainIDStr, 10, 64); err == nil {
			chainID = id
		}
	}

	// Get RPC URL from environment
	rpcURL := os.Getenv("BNB_RPC_URL")
	if rpcURL == "" {
		if chainID == constants.ChainIDBNBTestnet {
			rpcURL = "https://data-seed-prebsc-1-s1.binance.org:8545"
		} else {
			rpcURL = "https://bsc-dataseed.binance.org"
		}
	}

	// Call API
	depositInfo, address, err := predictclob.GetDepositInfoFromPrivateKey(chainID, privateKey)
	if err != nil {
		log.Fatalf("Error getting deposit info: %v", err)
	}

	// Print deposit info using %+v to show all fields
	fmt.Printf("Deposit Info:\n%+v\n\n", depositInfo)

	// Get USDT balance
	balance, err := predictclob.GetUSDTBalance(rpcURL, chainID, address.Hex())
	if err != nil {
		log.Fatalf("Error getting USDT balance: %v", err)
	}

	// Print balance using %+v to show all fields
	fmt.Printf("USDT Balance:\n%+v\n", balance)
}
