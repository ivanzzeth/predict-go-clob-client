package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethsig"
	"github.com/joho/godotenv"
	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
	predictcontracts "github.com/ivanzzeth/predict-go-contracts"
	conditional_tokens "github.com/ivanzzeth/predict-go-contracts/contracts/conditional-tokens"
	yb_conditional_tokens "github.com/ivanzzeth/predict-go-contracts/contracts/yb-conditional-tokens"
	"github.com/shopspring/decimal"
)

func main() {
	// Load .env file from project root directory
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get required environment variables
	privateKeyHex := os.Getenv("WALLET_PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("WALLET_PRIVATE_KEY environment variable is required")
	}

	marketIDStr := os.Getenv("MARKET_ID")
	if marketIDStr == "" {
		log.Fatal("MARKET_ID environment variable is required")
	}

	// Get API key from environment
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Get RPC URL from environment (required for chain operations)
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://bsc-dataseed1.binance.org"
		log.Printf("Using default RPC URL: %s", rpcURL)
	}

	// Parse private key and create signer
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}
	signer := ethsig.NewEthPrivateKeySigner(privateKey)
	address := signer.GetAddress()
	log.Printf("EOA Address: %s", address.Hex())

	// Create client with signer and RPC URL
	client, err := predictclob.NewClient(
		predictclob.WithAPIHost(constants.DefaultAPIHost),
		predictclob.WithAPIKey(apiKey),
		predictclob.WithEOATradingSigner(signer),
		predictclob.WithChainID(big.NewInt(56)),
		predictclob.WithRPCURL(rpcURL),
		predictclob.WithCacheTTL(5*time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	log.Printf("Client created successfully\n")

	// Parse market ID
	marketID := types.MustMarketIDFromString(marketIDStr)

	// Get market info
	market, err := client.GetMarket(marketID, true)
	if err != nil {
		log.Fatalf("Failed to get market: %v", err)
	}

	fmt.Printf("=== Market Information ===\n")
	fmt.Printf("Market ID: %s\n", marketID.String())
	fmt.Printf("Title: %s\n", market.Title)
	fmt.Printf("Condition ID: %s\n", market.ConditionID.Hex())
	fmt.Printf("Is Neg Risk: %v\n", market.IsNegRisk)
	fmt.Printf("Is Yield Bearing: %v\n", market.IsYieldBearing)
	fmt.Printf("Status: %s\n", market.Status.String())
	fmt.Printf("Outcomes: %d\n", len(market.Outcomes))
	for i, outcome := range market.Outcomes {
		fmt.Printf("  Outcome %d: %s (IndexSet: %d, OnChainID: %s)\n",
			i, outcome.Name, outcome.IndexSet, string(outcome.OnChainID))
	}
	fmt.Printf("\n")

	// Get contract interface
	contractInterface := client.GetContractInterface()
	if contractInterface == nil {
		log.Fatal("Contract interface not initialized")
	}

	// Get the appropriate CTF contract based on market type
	config := contractInterface.GetConfig()
	var ctfAddress common.Address
	if market.IsYieldBearing {
		if market.IsNegRisk {
			ctfAddress = config.YieldBearingNegRiskConditionalTokens
		} else {
			ctfAddress = config.YieldBearingConditionalTokens
		}
	} else {
		if market.IsNegRisk {
			ctfAddress = config.NegRiskConditionalTokens
		} else {
			ctfAddress = config.ConditionalTokens
		}
	}

	fmt.Printf("=== On-Chain Position Balances ===\n")
	fmt.Printf("CTF Contract: %s\n", ctfAddress.Hex())
	fmt.Printf("Account: %s\n", address.Hex())
	fmt.Printf("\n")

	ctx := context.Background()

	// Query balance for each outcome
	for i, outcome := range market.Outcomes {
		tokenID := new(big.Int)
		tokenID.SetString(string(outcome.OnChainID), 10)

		var balance *big.Int

		if market.IsYieldBearing {
			// Use YB CTF contract
			ybCtf, err := yb_conditional_tokens.NewYBConditionalTokens(ctfAddress, contractInterface.GetClient())
			if err != nil {
				log.Fatalf("Failed to create YB CTF contract: %v", err)
			}
			balance, err = ybCtf.BalanceOf(&bind.CallOpts{Context: ctx}, address, tokenID)
			if err != nil {
				log.Fatalf("Failed to get balance for outcome %d: %v", i, err)
			}
		} else {
			// Use regular CTF contract
			ctf, err := conditional_tokens.NewConditionalTokens(ctfAddress, contractInterface.GetClient())
			if err != nil {
				log.Fatalf("Failed to create CTF contract: %v", err)
			}
			balance, err = ctf.BalanceOf(&bind.CallOpts{Context: ctx}, address, tokenID)
			if err != nil {
				log.Fatalf("Failed to get balance for outcome %d: %v", i, err)
			}
		}

		// Convert to decimal (18 decimals)
		balanceDecimal := decimal.NewFromBigInt(balance, 0).Shift(-predictcontracts.COLLATERAL_TOKEN_DECIMALS)

		fmt.Printf("Outcome %d (%s):\n", i, outcome.Name)
		fmt.Printf("  Token ID: %s\n", tokenID.String())
		fmt.Printf("  Balance (wei): %s\n", balance.String())
		fmt.Printf("  Balance (tokens): %s\n", balanceDecimal.String())
		fmt.Printf("\n")
	}
}
