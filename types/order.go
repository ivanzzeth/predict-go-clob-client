package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/shopspring/decimal"
)

// OrderSide represents the side of an order
type OrderSide int

const (
	OrderSideBuy  OrderSide = 0
	OrderSideSell OrderSide = 1
)

// OrderStrategy represents the order strategy/type
type OrderStrategy string

const (
	OrderStrategyLimit  OrderStrategy = "LIMIT"
	OrderStrategyMarket OrderStrategy = "MARKET"
)

// UnixTime represents a time that can be unmarshaled from unix timestamp (int) or RFC3339 string
type UnixTime time.Time

// UnmarshalJSON implements json.Unmarshaler to handle both unix timestamp (int) and RFC3339 string
func (ut *UnixTime) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as int64 (unix timestamp)
	var timestamp int64
	if err := json.Unmarshal(data, &timestamp); err == nil {
		*ut = UnixTime(time.Unix(timestamp, 0))
		return nil
	}

	// Try to unmarshal as float64 (unix timestamp)
	var timestampFloat float64
	if err := json.Unmarshal(data, &timestampFloat); err == nil {
		*ut = UnixTime(time.Unix(int64(timestampFloat), 0))
		return nil
	}

	// Try to unmarshal as RFC3339 string
	var timestampStr string
	if err := json.Unmarshal(data, &timestampStr); err == nil {
		if timestampStr == "" {
			*ut = UnixTime(time.Time{})
			return nil
		}
		t, err := time.Parse(time.RFC3339, timestampStr)
		if err == nil {
			*ut = UnixTime(t)
			return nil
		}
		// Try other formats
		if t, err := time.Parse("2006-01-02T15:04:05Z", timestampStr); err == nil {
			*ut = UnixTime(t)
			return nil
		}
	}

	return fmt.Errorf("cannot unmarshal UnixTime: %s", string(data))
}

// Time returns the underlying time.Time
func (ut UnixTime) Time() time.Time {
	return time.Time(ut)
}

// PlaceOrderInput represents input for placing an order
type PlaceOrderInput struct {
	MarketID          MarketID        // Market ID
	TokenID           TokenID         // Token ID (onChainId from outcome)
	Side              OrderSide       // BUY or SELL
	Strategy          OrderStrategy   // LIMIT or MARKET
	Amount            decimal.Decimal // Amount in USDT (for BUY) or shares (for SELL)
	Price             decimal.Decimal // Price per share (required for LIMIT orders)
	SlippageBps       int             // Slippage in basis points (default 10 for MARKET orders)
	ExpirationSeconds int64           // Order expiration in seconds from now (default: constants.DefaultOrderExpirationSeconds)
}

// PlaceOrderResult represents the result of placing an order
type PlaceOrderResult struct {
	OrderID   string `json:"orderId"`
	OrderHash string `json:"orderHash"`
	Success   bool   `json:"success"`
}

// CancelOrderInput represents input for canceling orders
type CancelOrderInput struct {
	OrderIDs []string // Order IDs to cancel
}

// CancelOrderResult represents the result of canceling orders
type CancelOrderResult struct {
	Removed []string `json:"removed"`
	Noop    []string `json:"noop"`
	Success bool     `json:"success"`
}

// GetOrdersOptions represents options for getting orders
type GetOrdersOptions struct {
	First    *int        // Pagination: number of results to return
	After    string      // Pagination: cursor for next page
	MarketID MarketID    // Filter by market ID
	Status   OrderStatus // Filter by status (OPEN, FILLED)
}

// OrderDataInfo represents the nested order data
type OrderDataInfo struct {
	Hash           common.Hash     `json:"hash,omitempty"` // common.Hash implements json.Unmarshaler
	Salt           string          `json:"salt"`
	Maker          common.Address  `json:"maker"`       // common.Address implements json.Unmarshaler
	Signer         common.Address  `json:"signer"`      // common.Address implements json.Unmarshaler
	Taker          common.Address  `json:"taker"`       // common.Address implements json.Unmarshaler
	TokenID        TokenID         `json:"tokenId"`     // String-based type, auto-handled
	MakerAmount    decimal.Decimal `json:"-"`           // Human readable decimal (converted from wei)
	TakerAmount    decimal.Decimal `json:"-"`           // Human readable decimal (converted from wei)
	RawMakerAmount string          `json:"makerAmount"` // Raw wei amount as string
	RawTakerAmount string          `json:"takerAmount"` // Raw wei amount as string
	Expiration     UnixTime        `json:"expiration"`  // Unix timestamp (int) or RFC3339 string
	Nonce          string          `json:"nonce"`
	FeeRateBps     string          `json:"feeRateBps"`
	Side           OrderSide       `json:"side"` // Int-based type, auto-handled
	SignatureType  int             `json:"signatureType"`
	Signature      string          `json:"signature"`
}

// UnmarshalJSON implements custom unmarshaling for OrderDataInfo to handle wei to decimal conversion
func (o *OrderDataInfo) UnmarshalJSON(data []byte) error {
	type Alias OrderDataInfo
	aux := &struct {
		MakerAmount string `json:"makerAmount"`
		TakerAmount string `json:"takerAmount"`
		*Alias
	}{
		Alias: (*Alias)(o),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert MakerAmount from wei to decimal
	if aux.MakerAmount != "" {
		o.RawMakerAmount = aux.MakerAmount
		makerAmountWei, err := decimal.NewFromString(aux.MakerAmount)
		if err == nil {
			o.MakerAmount = makerAmountWei.Shift(-constants.TokenDecimals)
		}
	}

	// Convert TakerAmount from wei to decimal
	if aux.TakerAmount != "" {
		o.RawTakerAmount = aux.TakerAmount
		takerAmountWei, err := decimal.NewFromString(aux.TakerAmount)
		if err == nil {
			o.TakerAmount = takerAmountWei.Shift(-constants.TokenDecimals)
		}
	}

	return nil
}

// Order represents an order from the API (used in GetOrders list)
type Order struct {
	ID                string          `json:"id"`
	MarketID          MarketID        `json:"marketId"` // Handles int/string via UnmarshalJSON
	Currency          string          `json:"currency"`
	Amount            decimal.Decimal `json:"-"`            // Human readable decimal (converted from wei)
	AmountFilled      decimal.Decimal `json:"-"`            // Human readable decimal (converted from wei)
	RawAmount         string          `json:"amount"`       // Raw wei amount as string
	RawAmountFilled   string          `json:"amountFilled"` // Raw wei amount filled as string
	IsNegRisk         bool            `json:"isNegRisk"`
	IsYieldBearing    bool            `json:"isYieldBearing"`
	Strategy          OrderStrategy   `json:"strategy"` // String-based type, auto-handled
	Status            OrderStatus     `json:"status"`   // String-based type, auto-handled
	RewardEarningRate float64         `json:"rewardEarningRate"`
	OrderData         OrderDataInfo   `json:"order"`
}

// UnmarshalJSON implements custom unmarshaling for Order to handle wei to decimal conversion
func (o *Order) UnmarshalJSON(data []byte) error {
	type Alias Order
	aux := &struct {
		Amount       string `json:"amount"`
		AmountFilled string `json:"amountFilled"`
		*Alias
	}{
		Alias: (*Alias)(o),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert Amount from wei to decimal
	if aux.Amount != "" {
		o.RawAmount = aux.Amount
		amountWei, err := decimal.NewFromString(aux.Amount)
		if err == nil {
			o.Amount = amountWei.Shift(-constants.TokenDecimals)
		}
	}

	// Convert AmountFilled from wei to decimal
	if aux.AmountFilled != "" {
		o.RawAmountFilled = aux.AmountFilled
		amountFilledWei, err := decimal.NewFromString(aux.AmountFilled)
		if err == nil {
			o.AmountFilled = amountFilledWei.Shift(-constants.TokenDecimals)
		}
	}

	return nil
}

// GetOrdersResponse represents the response from GetOrders API
type GetOrdersResponse struct {
	Success bool    `json:"success"`
	Cursor  *string `json:"cursor,omitempty"` // nullable
	Data    []Order `json:"data"`
}
