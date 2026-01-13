# Get Market Statistics

## Endpoint

```
GET /v1/markets/{id}/stats
```

## Description

Get statistics about a specific market.

## Authentication

Requires API key authentication via `x-api-key` header.

## Parameters

### Path Parameters

| Parameter | Type   | Required | Description                    |
|-----------|--------|----------|--------------------------------|
| id        | string | Yes      | Market ID (string to be decoded into a number) |

## Response

### Success Response (200 OK)

```json
{
  "success": true,
  "data": {
    "totalLiquidityUsd": 1234567.89,
    "volumeTotalUsd": 9876543.21,
    "volume24hUsd": 123456.78
  }
}
```

### Response Fields

| Field            | Type   | Description                    |
|------------------|--------|--------------------------------|
| success          | boolean| Indicates if the request was successful |
| data              | object | Market statistics data         |
| data.totalLiquidityUsd | number | Total liquidity in USD |
| data.volumeTotalUsd | number | Total volume in USD |
| data.volume24hUsd | number | 24-hour volume in USD |

## Example Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/ivanzzeth/predict-go-clob-client"
    "github.com/ivanzzeth/predict-go-clob-client/constants"
    "github.com/ivanzzeth/predict-go-clob-client/types"
)

func main() {
    // Create read-only client with API key
    client := predictclob.NewReadOnlyClient(constants.DefaultAPIHost, "your-api-key")

    // Get market ID
    marketID := types.MustMarketIDFromString("123")

    // Get market statistics
    stats, err := client.GetMarketStats(marketID)
    if err != nil {
        log.Fatalf("Error getting market stats: %v", err)
    }

    // Access statistics
    fmt.Printf("Total Liquidity: %s USD\n", stats.TotalLiquidityUsd.String())
    fmt.Printf("Total Volume: %s USD\n", stats.VolumeTotalUsd.String())
    fmt.Printf("24h Volume: %s USD\n", stats.Volume24hUsd.String())
}
```

## SDK Implementation

The SDK provides the `GetMarketStats` method on the `Client` struct:

```go
func (c *Client) GetMarketStats(marketID types.MarketID) (*types.MarketStats, error)
```

### Return Type

The method returns a `*types.MarketStats` struct with the following fields:

- `TotalLiquidityUsd` (decimal.Decimal): Total liquidity in USD (human-readable decimal)
- `VolumeTotalUsd` (decimal.Decimal): Total volume in USD (human-readable decimal)
- `Volume24hUsd` (decimal.Decimal): 24-hour volume in USD (human-readable decimal)
- `RawTotalLiquidityUsd` (string): Raw total liquidity USD as string
- `RawVolumeTotalUsd` (string): Raw total volume USD as string
- `RawVolume24hUsd` (string): Raw 24-hour volume USD as string

All USD amounts are stored as `decimal.Decimal` for precise financial calculations. The raw string values are preserved for reference.

## Error Handling

The method may return the following errors:

- `ErrAPIKeyRequired`: If API key is not set
- API errors: If the API returns a non-200 status code
- Parsing errors: If the response cannot be parsed

## See Also

- [Get Market](./get-market.md)
- [Get Market Orderbook](./get-market-orderbook.md)
- [Get Market Last Sale](./get-market-last-sale.md)
- [Examples](../../../examples/market/stats/main.go)
