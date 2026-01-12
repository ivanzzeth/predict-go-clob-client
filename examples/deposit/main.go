package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
)

func main() {
	// Load .env file from project root
	_ = godotenv.Load(".env")

	// Get configuration from environment
	privateKey := os.Getenv("WALLET_PRIVATE_KEY")
	rpcURL := os.Getenv("BNB_RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://bsc-dataseed.binance.org" // Default BNB Mainnet RPC
	}

	chainID := int64(constants.ChainIDBNBMainnet)
	if os.Getenv("CHAIN_ID") == "97" {
		chainID = constants.ChainIDBNBTestnet
		rpcURL = "https://data-seed-prebsc-1-s1.binance.org:8545" // Default BNB Testnet RPC
	}

	if privateKey == "" {
		log.Fatal("WALLET_PRIVATE_KEY environment variable is required")
	}

	fmt.Println("=== Predict.fun Deposit Module ===\n")

	// Get deposit info from private key
	fmt.Println("1. Getting deposit information...")
	depositInfo, address, err := predictclob.GetDepositInfoFromPrivateKey(chainID, privateKey)
	if err != nil {
		log.Fatalf("Failed to get deposit info: %v", err)
	}

	fmt.Printf("   ✓ Wallet Address: %s\n", depositInfo.WalletAddress)
	fmt.Printf("   ✓ USDT Contract: %s\n", depositInfo.USDTAddress)
	fmt.Printf("   ✓ Chain: %s (Chain ID: %d)\n\n", depositInfo.ChainName, depositInfo.ChainID)

	// Get USDT balance
	fmt.Println("2. Checking USDT balance...")
	balance, err := predictclob.GetUSDTBalance(rpcURL, chainID, address.Hex())
	if err != nil {
		log.Fatalf("Failed to get USDT balance: %v", err)
	}

	fmt.Printf("   ✓ Balance: %s USDT\n", balance.BalanceUSDT)
	fmt.Printf("   ✓ Balance (wei): %s\n\n", balance.BalanceWei.String())

	// Show deposit instructions
	fmt.Println("3. Deposit Instructions:")
	fmt.Println("   To deposit USDT:")
	fmt.Printf("   1. Transfer USDT to: %s\n", depositInfo.WalletAddress)
	fmt.Printf("   2. Make sure you're on %s (Chain ID: %d)\n", depositInfo.ChainName, depositInfo.ChainID)
	fmt.Printf("   3. USDT Contract Address: %s\n", depositInfo.USDTAddress)
	fmt.Println("   4. Wait for on-chain confirmation")
	fmt.Println("   5. Check balance again using this tool")
	fmt.Println("\n   Note: Deposit is a standard on-chain ERC-20 transfer.")
	fmt.Println("         You can use MetaMask, exchange, or any wallet to send USDT.")
}
