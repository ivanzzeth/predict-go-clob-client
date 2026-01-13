# Predict Go CLOB Client

A Golang SDK for interacting with the Predict.fun platform API. This client provides a comprehensive interface for querying market data and managing orders on Predict.fun.

## Features

- 📊 **Market Data** - Query categories, markets, orderbooks, and statistics
- 🔐 **Authentication** - EOA wallet authentication with JWT token support
- 📝 **Order Management** - Create, cancel, and query orders
- ✅ **Well Tested** - Comprehensive unit tests for all core functionality
- 🎯 **Type Safe** - Strongly typed API with clear interfaces

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
    fmt.Printf("Found %d categories\n", len(categories))

    // Get markets
    markets, err := client.GetMarkets(nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d markets\n", len(markets))
}
```

### Authenticated Client

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
    // Create client with API key
    client, err := predictclob.NewClient(
        predictclob.WithAPIHost(constants.DefaultAPIHost),
        predictclob.WithAPIKey(os.Getenv("PREDICT_API_KEY")),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Authenticate with private key
    token, address, err := client.Authenticate(os.Getenv("WALLET_PRIVATE_KEY"))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Authenticated as: %s\n", address.Hex())

    // Get account info
    account, err := client.GetAccount()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Account: %s\n", account.Name)

    // Get positions
    positions, err := client.GetPositions(nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d positions\n", len(positions))
}
```

## API Reference

See [docs/API.md](docs/API.md) for detailed API documentation.

## License

MIT
