# Predict Go CLOB Client

A Golang SDK for interacting with the Predict.fun platform API. This client provides a comprehensive interface for querying market data, managing orders, and performing chain operations on Predict.fun.

## Features

- 📊 **Market Data** - Query categories, markets, orderbooks, and statistics
- 🔐 **Authentication** - EOA wallet authentication with JWT token support
- 📝 **Order Management** - Create, cancel, and query orders
- ⛓️ **Chain Operations** - Split, merge, and redeem positions with automatic neg-risk detection
- 💰 **Balance & Positions** - Query user balances and positions
- 🎯 **Type Safe** - Strongly typed API with clear interfaces
- ⚡ **Caching** - Built-in market data caching with TTL support

## Installation

```bash
go get github.com/ivanzzeth/predict-go-clob-client
```

## Quick Start

### Read-Only Client (API Key Required)

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/ivanzzeth/predict-go-clob-client"
    "github.com/ivanzzeth/predict-go-clob-client/constants"
)

func main() {
    // Create a read-only client with API key
    client := predictclob.NewReadOnlyClient(constants.DefaultAPIHost, os.Getenv("PREDICT_API_KEY"))

    // Get categories
    categories, err := client.GetCategories(nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d categories\n", len(categories.Data))

    // Get markets
    markets, err := client.GetMarkets(nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d markets\n", len(markets.Data))
}
```

### Authenticated Client with Chain Operations

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/big"
    "os"

    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ivanzzeth/ethsig"
    "github.com/ivanzzeth/predict-go-clob-client"
    "github.com/ivanzzeth/predict-go-clob-client/constants"
    "github.com/ivanzzeth/predict-go-clob-client/types"
    "github.com/shopspring/decimal"
)

func main() {
    // Parse private key
    privateKey, _ := crypto.HexToECDSA(os.Getenv("WALLET_PRIVATE_KEY"))
    signer := ethsig.NewEthPrivateKeySigner(privateKey)

    // Create client with signer and RPC URL for chain operations
    client, err := predictclob.NewClient(
        predictclob.WithAPIHost(constants.DefaultAPIHost),
        predictclob.WithAPIKey(os.Getenv("PREDICT_API_KEY")),
        predictclob.WithEOATradingSigner(signer),
        predictclob.WithChainID(big.NewInt(56)), // BNB Mainnet
        predictclob.WithRPCURL(os.Getenv("RPC_URL")),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Enable trading (approve tokens)
    ctx := context.Background()
    txHashes, err := client.EnableTrading(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Enabled trading: %d transactions\n", len(txHashes))

    // Split collateral into outcome tokens
    marketID := types.MustMarketIDFromString("1910")
    amount := decimal.NewFromFloat(0.1)
    txHash, err := client.Split(ctx, marketID, amount)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Split transaction: %s\n", txHash.Hex())
}
```

## Examples

This SDK includes comprehensive examples for all major features. All examples use environment variables for configuration and load them from `.env` file in the project root.

### Authentication

- **[Authenticate](./examples/auth/authenticate/main.go)** - Authenticate with EOA wallet and get JWT token

### Market Operations

- **[List Markets](./examples/market/list/main.go)** - List all markets with pagination
- **[Get Market Detail](./examples/market/detail/main.go)** - Get detailed information about a specific market
- **[Get Categories](./examples/market/categories/main.go)** - List all market categories
- **[Get Category](./examples/market/category/main.go)** - Get detailed information about a specific category
- **[Find Open Markets](./examples/market/find-open/main.go)** - Find OPEN markets by querying OPEN categories
- **[Get Orderbook](./examples/market/orderbook/main.go)** - Get orderbook for a market
- **[Get Market Stats](./examples/market/stats/main.go)** - Get market statistics
- **[Get Last Sale](./examples/market/last-sale/main.go)** - Get last sale information for a market

### Order Operations

- **[Place Order](./examples/order/place/main.go)** - Place a new order (BUY or SELL)
- **[Cancel Orders](./examples/order/cancel/main.go)** - Cancel one or more orders
- **[List Orders](./examples/order/list/main.go)** - List user's orders with filtering
- **[Get Order by Hash](./examples/order/get-by-hash/main.go)** - Get order details by transaction hash
- **[Get Matches](./examples/order/matches/main.go)** - Get order matches/trades

### User Operations

- **[Get Account](./examples/user/account/main.go)** - Get user account information
- **[Get Positions](./examples/user/positions/main.go)** - Get user's positions (holdings)
- **[Get Balance](./examples/user/balance/main.go)** - Get user's collateral balance
- **[Get Activity](./examples/user/activity/main.go)** - Get user's trading activity

### Chain Operations

- **[Enable Trading](./examples/chain/enable-trading/main.go)** - Approve tokens for trading (required before split/merge/redeem)
- **[Split](./examples/chain/split/main.go)** - Split collateral into outcome tokens
- **[Merge](./examples/chain/merge/main.go)** - Merge outcome tokens back into collateral
- **[Redeem](./examples/chain/redeem/main.go)** - Redeem positions for resolved markets

## Environment Variables

All examples use environment variables loaded from `.env` file. Required variables:

```bash
# API Configuration
PREDICT_API_KEY=your_api_key_here

# Wallet Configuration (for authenticated operations)
WALLET_PRIVATE_KEY=your_private_key_hex_here

# RPC Configuration (for chain operations)
RPC_URL=https://bnb-mainnet.g.alchemy.com/v2/your_key

# Market/Order Configuration (example-specific)
MARKET_ID=1910
AMOUNT=0.1
AMOUNTS=100000000000000000,100000000000000000  # For neg-risk redeem (comma-separated wei amounts)
```

## Running Examples

```bash
# Set environment variables
export PREDICT_API_KEY=your_api_key
export WALLET_PRIVATE_KEY=your_private_key
export RPC_URL=your_rpc_url

# Or use .env file
cp .env.example .env
# Edit .env with your credentials

# Run examples
go run ./examples/market/list/main.go
go run ./examples/chain/split/main.go MARKET_ID=1910 AMOUNT=0.1
go run ./examples/chain/redeem/main.go MARKET_ID=1757
```

## Key Features

### Automatic Neg-Risk Detection

The SDK automatically detects neg-risk markets and uses the appropriate contract methods:

```go
// Automatically uses SplitNegRisk if market.IsNegRisk is true
txHash, err := client.Split(ctx, marketID, amount)

// Automatically uses MergeNegRisk if market.IsNegRisk is true
txHash, err := client.Merge(ctx, marketID, amount)

// Automatically uses RedeemNegRisk if market.IsNegRisk is true
txHash, err := client.Redeem(ctx, marketID, amounts)
```

### Market Data Caching

Enable caching to reduce API calls for chain operations:

```go
client, err := predictclob.NewClient(
    predictclob.WithCacheTTL(5 * time.Minute), // Cache for 5 minutes
    // ... other options
)

// Use cache for chain operations (only needs conditionID, not real-time data)
market, err := client.GetMarket(marketID, true) // useCache=true
```

### Type Safety

All financial amounts use `decimal.Decimal` for precision:

```go
amount := decimal.NewFromFloat(0.1)
txHash, err := client.Split(ctx, marketID, amount)
```

## API Reference

See [docs/API.md](docs/API.md) for detailed API documentation.

## License

MIT
