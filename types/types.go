package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/shopspring/decimal"
)

// APIBaseResponse represents the base response structure from Predict API
// Uses generics to avoid interface{} and double parsing
type APIBaseResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data"`
	Message string `json:"message,omitempty"`
}

// Category is defined in category.go

// Market represents a prediction market (matches API response format)
// This is the same structure as CategoryMarket but used in GetMarkets response
type Market struct {
	ID                     MarketID                  `json:"id"`          // IntegerOrString (handles int/string via UnmarshalJSON)
	ImageURL               string                    `json:"imageUrl"`    // required
	Title                  string                    `json:"title"`       // required
	Question               string                    `json:"question"`    // required
	Description            string                    `json:"description"` // required
	Status                 MarketStatus              `json:"status"`      // enum
	IsNegRisk              bool                      `json:"isNegRisk"`
	IsYieldBearing         bool                      `json:"isYieldBearing"`
	FeeRateBps             FeeRateBps                `json:"feeRateBps"`              // Handles int/string via UnmarshalJSON
	Resolution             *CategoryMarketResolution `json:"resolution,omitempty"`    // nullable
	OracleQuestionID       common.Hash               `json:"oracleQuestionId"`        // common.Hash implements json.Unmarshaler
	ConditionID            common.Hash               `json:"conditionId"`             // required, common.Hash implements json.Unmarshaler
	ResolverAddress        common.Address            `json:"resolverAddress"`         // common.Address implements json.Unmarshaler
	Outcomes               []CategoryMarketOutcome   `json:"outcomes"`                // required
	QuestionIndex          *int                      `json:"questionIndex,omitempty"` // nullable
	SpreadThreshold        float64                   `json:"spreadThreshold"`
	ShareThreshold         float64                   `json:"shareThreshold"`
	PolymarketConditionIDs []common.Hash             `json:"polymarketConditionIds"`       // []common.Hash, each implements json.Unmarshaler
	KalshiMarketTicker     *string                   `json:"kalshiMarketTicker,omitempty"` // nullable
	CategorySlug           string                    `json:"categorySlug"`                 // required
	CreatedAt              time.Time                 `json:"createdAt"`                    // date string
	DecimalPrecision       int                       `json:"decimalPrecision"`             // enum: 2 or 3
}

// Market doesn't need custom UnmarshalJSON
// MarketID (IntegerOrString) handles int/string via its own UnmarshalJSON
// MarketStatus is string-based, auto-handled
// ResolverAddress (common.Address) implements json.Unmarshaler, auto-handled
// FeeRateBps handles int/string via its own UnmarshalJSON
// CategoryMarketOutcome and CategoryMarketResolution are defined in category.go

// Outcome represents a market outcome
type Outcome struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	IndexSet  int     `json:"indexSet"`
	OnChainID TokenID `json:"onChainId"`
}

// MarketStats represents market statistics
// All USD amounts are stored as decimal.Decimal for precise calculations
type MarketStats struct {
	TotalLiquidityUsd decimal.Decimal `json:"totalLiquidityUsd"` // Total liquidity in USD (human readable decimal)
	VolumeTotalUsd    decimal.Decimal `json:"volumeTotalUsd"`    // Total volume in USD (human readable decimal)
	Volume24hUsd      decimal.Decimal `json:"volume24hUsd"`      // 24-hour volume in USD (human readable decimal)
}

// UnmarshalJSON implements custom unmarshaling for MarketStats
// decimal.Decimal supports unmarshaling from both string and number formats
func (m *MarketStats) UnmarshalJSON(data []byte) error {
	type Alias MarketStats
	aux := &struct {
		TotalLiquidityUsd interface{} `json:"totalLiquidityUsd"`
		VolumeTotalUsd    interface{} `json:"volumeTotalUsd"`
		Volume24hUsd      interface{} `json:"volume24hUsd"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert TotalLiquidityUsd (can be number or string)
	if aux.TotalLiquidityUsd != nil {
		var totalLiquidityStr string
		switch v := aux.TotalLiquidityUsd.(type) {
		case string:
			totalLiquidityStr = v
		case float64:
			totalLiquidityStr = fmt.Sprintf("%v", v)
		case int:
			totalLiquidityStr = fmt.Sprintf("%d", v)
		case int64:
			totalLiquidityStr = fmt.Sprintf("%d", v)
		default:
			totalLiquidityStr = fmt.Sprintf("%v", v)
		}
		if totalLiquidityStr != "" {
			if totalLiquidity, err := decimal.NewFromString(totalLiquidityStr); err == nil {
				m.TotalLiquidityUsd = totalLiquidity
			}
		}
	}

	// Convert VolumeTotalUsd (can be number or string)
	if aux.VolumeTotalUsd != nil {
		var volumeTotalStr string
		switch v := aux.VolumeTotalUsd.(type) {
		case string:
			volumeTotalStr = v
		case float64:
			volumeTotalStr = fmt.Sprintf("%v", v)
		case int:
			volumeTotalStr = fmt.Sprintf("%d", v)
		case int64:
			volumeTotalStr = fmt.Sprintf("%d", v)
		default:
			volumeTotalStr = fmt.Sprintf("%v", v)
		}
		if volumeTotalStr != "" {
			if volumeTotal, err := decimal.NewFromString(volumeTotalStr); err == nil {
				m.VolumeTotalUsd = volumeTotal
			}
		}
	}

	// Convert Volume24hUsd (can be number or string)
	if aux.Volume24hUsd != nil {
		var volume24hStr string
		switch v := aux.Volume24hUsd.(type) {
		case string:
			volume24hStr = v
		case float64:
			volume24hStr = fmt.Sprintf("%v", v)
		case int:
			volume24hStr = fmt.Sprintf("%d", v)
		case int64:
			volume24hStr = fmt.Sprintf("%d", v)
		default:
			volume24hStr = fmt.Sprintf("%v", v)
		}
		if volume24hStr != "" {
			if volume24h, err := decimal.NewFromString(volume24hStr); err == nil {
				m.Volume24hUsd = volume24h
			}
		}
	}

	return nil
}

// OrderbookLevel represents a single level in the orderbook
type OrderbookLevel struct {
	Price     decimal.Decimal `json:"-"`      // Human readable decimal (converted from wei)
	Amount    decimal.Decimal `json:"-"`      // Human readable decimal (converted from wei)
	RawPrice  string          `json:"price"`  // Raw wei price as string
	RawAmount string          `json:"amount"` // Raw wei amount as string
}

// LastOrderSettled represents the last settled order information
type LastOrderSettled struct {
	ID       string        `json:"id"`       // Order ID
	Price    string        `json:"price"`    // Price as string (already in decimal format, not wei)
	Kind     OrderStrategy `json:"kind"`     // Order strategy (MARKET or LIMIT)
	MarketID MarketID      `json:"marketId"` // Market identifier
	Side     OrderSide     `json:"side"`     // Order side (Bid or Ask)
	Outcome  MarketOutcome `json:"outcome"`  // Market outcome (Yes or No)
}

// Orderbook represents the orderbook for a market.
//
// Sorting rules (automatically applied during JSON unmarshaling - standard orderbook practice):
//   - Bids: sorted in descending order by price (highest price first, best bid at index 0)
//     Example: [0.60, 0.55, 0.50, ...] where 0.60 is the best bid (Bids[0])
//   - Asks: sorted in ascending order by price (lowest price first, best ask at index 0)
//     Example: [0.40, 0.45, 0.50, ...] where 0.40 is the best ask (Asks[0])
//
// This ensures that (standard orderbook practice):
//   - Bids[0] is always the best bid (highest buy price, bestBid)
//   - Asks[0] is always the best ask (lowest sell price, bestAsk)
//   - The spread can be calculated as Asks[0].Price - Bids[0].Price
type Orderbook struct {
	MarketID          MarketID          `json:"marketId"`                   // Market identifier
	UpdateTimestampMs int64             `json:"updateTimestampMs"`          // Update timestamp in milliseconds
	LastOrderSettled  *LastOrderSettled `json:"lastOrderSettled,omitempty"` // Last settled order (optional)
	Bids              []OrderbookLevel  `json:"bids"`                       // Buy orders, sorted descending by price
	Asks              []OrderbookLevel  `json:"asks"`                       // Sell orders, sorted ascending by price
	BestBid           decimal.Decimal   `json:"-"`                          // Calculated: best bid price (highest buy price, Bids[0].Price if exists)
	BestAsk           decimal.Decimal   `json:"-"`                          // Calculated: best ask price (lowest sell price, Asks[0].Price if exists)
	Spread            decimal.Decimal   `json:"-"`                          // Calculated: spread = BestAsk - BestBid (only valid when both exist)
}

// UnmarshalJSON implements custom unmarshaling for OrderbookLevel to convert wei to decimal
func (ol *OrderbookLevel) UnmarshalJSON(data []byte) error {
	// Handle array format: [price, amount]
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err == nil && len(arr) >= 2 {
		// Extract price and amount as numbers (can be float64 or string)
		var priceStr, amountStr string

		switch v := arr[0].(type) {
		case float64:
			priceStr = fmt.Sprintf("%.0f", v)
		case string:
			priceStr = v
		default:
			priceStr = fmt.Sprintf("%v", v)
		}

		switch v := arr[1].(type) {
		case float64:
			amountStr = fmt.Sprintf("%.0f", v)
		case string:
			amountStr = v
		default:
			amountStr = fmt.Sprintf("%v", v)
		}

		ol.RawPrice = priceStr
		ol.RawAmount = amountStr

		// Convert price from wei to decimal
		if priceStr != "" {
			priceWei, err := decimal.NewFromString(priceStr)
			if err == nil {
				ol.Price = priceWei.Shift(-constants.TokenDecimals)
			}
		}

		// Convert amount from wei to decimal
		if amountStr != "" {
			amountWei, err := decimal.NewFromString(amountStr)
			if err == nil {
				ol.Amount = amountWei.Shift(-constants.TokenDecimals)
			}
		}

		return nil
	}

	// Handle object format: {"price": "...", "amount": "..."}
	type Alias OrderbookLevel
	aux := &struct {
		Price  interface{} `json:"price"`
		Amount interface{} `json:"amount"`
		*Alias
	}{
		Alias: (*Alias)(ol),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Extract price
	var priceStr string
	switch v := aux.Price.(type) {
	case float64:
		priceStr = fmt.Sprintf("%.0f", v)
	case string:
		priceStr = v
	default:
		priceStr = fmt.Sprintf("%v", v)
	}

	// Extract amount
	var amountStr string
	switch v := aux.Amount.(type) {
	case float64:
		amountStr = fmt.Sprintf("%.0f", v)
	case string:
		amountStr = v
	default:
		amountStr = fmt.Sprintf("%v", v)
	}

	ol.RawPrice = priceStr
	ol.RawAmount = amountStr

	// Convert price from wei to decimal
	if priceStr != "" {
		priceWei, err := decimal.NewFromString(priceStr)
		if err == nil {
			ol.Price = priceWei.Shift(-constants.TokenDecimals)
		}
	}

	// Convert amount from wei to decimal
	if amountStr != "" {
		amountWei, err := decimal.NewFromString(amountStr)
		if err == nil {
			ol.Amount = amountWei.Shift(-constants.TokenDecimals)
		}
	}

	return nil
}

// UnmarshalJSON implements custom unmarshaling for Orderbook to handle array format
func (o *Orderbook) UnmarshalJSON(data []byte) error {
	aux := &struct {
		MarketID          interface{}   `json:"marketId"`
		UpdateTimestampMs interface{}   `json:"updateTimestampMs"`
		LastOrderSettled  interface{}   `json:"lastOrderSettled,omitempty"`
		Bids              []interface{} `json:"bids"`
		Asks              []interface{} `json:"asks"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert MarketID
	if aux.MarketID != nil {
		var marketIDStr string
		switch v := aux.MarketID.(type) {
		case float64:
			marketIDStr = fmt.Sprintf("%.0f", v)
		case string:
			marketIDStr = v
		default:
			marketIDStr = fmt.Sprintf("%v", v)
		}
		if marketIDStr != "" {
			if id, err := NewMarketIDFromString(marketIDStr); err == nil {
				o.MarketID = id
			}
		}
	}

	// Convert UpdateTimestampMs
	if aux.UpdateTimestampMs != nil {
		switch v := aux.UpdateTimestampMs.(type) {
		case float64:
			o.UpdateTimestampMs = int64(v)
		case int64:
			o.UpdateTimestampMs = v
		case int:
			o.UpdateTimestampMs = int64(v)
		}
	}

	// Convert LastOrderSettled
	if aux.LastOrderSettled != nil {
		lastOrderData, err := json.Marshal(aux.LastOrderSettled)
		if err == nil {
			var lastOrder LastOrderSettled
			if err := json.Unmarshal(lastOrderData, &lastOrder); err == nil {
				o.LastOrderSettled = &lastOrder
			}
		}
	}

	// Convert bids
	o.Bids = make([]OrderbookLevel, 0, len(aux.Bids))
	for _, item := range aux.Bids {
		itemData, err := json.Marshal(item)
		if err != nil {
			continue
		}
		var level OrderbookLevel
		if err := json.Unmarshal(itemData, &level); err == nil {
			o.Bids = append(o.Bids, level)
		}
	}

	// Convert asks
	o.Asks = make([]OrderbookLevel, 0, len(aux.Asks))
	for _, item := range aux.Asks {
		itemData, err := json.Marshal(item)
		if err != nil {
			continue
		}
		var level OrderbookLevel
		if err := json.Unmarshal(itemData, &level); err == nil {
			o.Asks = append(o.Asks, level)
		}
	}

	// Sort bids in descending order (highest price first, best bid at index 0)
	// Standard practice: bestBid = Bids[0] (highest buy price)
	sort.Slice(o.Bids, func(i, j int) bool {
		return o.Bids[i].Price.GreaterThan(o.Bids[j].Price)
	})

	// Sort asks in ascending order (lowest price first, best ask at index 0)
	// Standard practice: bestAsk = Asks[0] (lowest sell price)
	sort.Slice(o.Asks, func(i, j int) bool {
		return o.Asks[i].Price.LessThan(o.Asks[j].Price)
	})

	// Calculate BestBid, BestAsk, and Spread (standard orderbook practice)
	// BestBid is always at Bids[0] after descending sort
	if len(o.Bids) > 0 {
		o.BestBid = o.Bids[0].Price
	}
	// BestAsk is always at Asks[0] after ascending sort
	if len(o.Asks) > 0 {
		o.BestAsk = o.Asks[0].Price
	}
	// Spread = BestAsk - BestBid (only valid when both exist)
	if len(o.Bids) > 0 && len(o.Asks) > 0 {
		o.Spread = o.BestAsk.Sub(o.BestBid)
	}

	return nil
}

// Sale represents a trade/sale (deprecated, use MarketLastSale for last-sale endpoint)
type Sale struct {
	TransactionHash common.Hash    `json:"transactionHash"`
	Price           string         `json:"price"`
	Amount          string         `json:"amount"`
	Seller          common.Address `json:"seller"`
	Buyer           common.Address `json:"buyer"`
	Timestamp       time.Time      `json:"timestamp"`
}

// MarketOutcome represents the outcome of a market (Yes or No)
type MarketOutcome string

const (
	MarketOutcomeYes MarketOutcome = "Yes"
	MarketOutcomeNo  MarketOutcome = "No"
)

// String returns the string representation of the outcome
func (o MarketOutcome) String() string {
	return string(o)
}

// MarketLastSale represents the last sale information for a market
// This matches the response from /v1/markets/{id}/last-sale endpoint
type MarketLastSale struct {
	QuoteType          QuoteType       `json:"quoteType"`       // Ask or Bid
	Outcome            MarketOutcome   `json:"outcome"`         // Yes or No
	PriceInCurrency    decimal.Decimal `json:"-"`               // Human readable decimal (converted from wei)
	RawPriceInCurrency string          `json:"priceInCurrency"` // Raw wei price as string
	Strategy           OrderStrategy   `json:"strategy"`        // MARKET or LIMIT
}

// UnmarshalJSON implements custom unmarshaling for MarketLastSale to convert wei to decimal
func (m *MarketLastSale) UnmarshalJSON(data []byte) error {
	type Alias MarketLastSale
	aux := &struct {
		PriceInCurrency string `json:"priceInCurrency"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Save raw value and convert PriceInCurrency from wei to decimal
	if aux.PriceInCurrency != "" {
		m.RawPriceInCurrency = aux.PriceInCurrency
		priceWei, err := decimal.NewFromString(aux.PriceInCurrency)
		if err == nil {
			m.PriceInCurrency = priceWei.Shift(-constants.TokenDecimals)
		}
	}

	return nil
}

// UnmarshalJSON implements custom unmarshaling for Sale to handle Address and TransactionHash conversion
func (s *Sale) UnmarshalJSON(data []byte) error {
	type Alias Sale
	aux := &struct {
		TransactionHash interface{} `json:"transactionHash"`
		Seller          interface{} `json:"seller"`
		Buyer           interface{} `json:"buyer"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle TransactionHash: convert string to common.Hash
	if aux.TransactionHash != nil {
		if hashStr, ok := aux.TransactionHash.(string); ok && hashStr != "" {
			s.TransactionHash = common.HexToHash(hashStr)
		} else {
			return fmt.Errorf("invalid transactionHash type: %T", aux.TransactionHash)
		}
	}

	// Handle Seller: convert string to common.Address
	if aux.Seller != nil {
		if sellerStr, ok := aux.Seller.(string); ok {
			s.Seller = common.HexToAddress(sellerStr)
		} else {
			return fmt.Errorf("invalid seller address type: %T", aux.Seller)
		}
	}

	// Handle Buyer: convert string to common.Address
	if aux.Buyer != nil {
		if buyerStr, ok := aux.Buyer.(string); ok {
			s.Buyer = common.HexToAddress(buyerStr)
		} else {
			return fmt.Errorf("invalid buyer address type: %T", aux.Buyer)
		}
	}

	return nil
}

// GetCategoriesOptions is defined in category.go

// GetMarketsOptions represents options for getting markets
type GetMarketsOptions struct {
	First *string `json:"first,omitempty"` // string to be decoded into a number
	After *string `json:"after,omitempty"` // pagination cursor
}

// GetMarketsResponse represents the response from GetMarkets API
type GetMarketsResponse struct {
	Success bool     `json:"success"`
	Cursor  *string  `json:"cursor,omitempty"` // nullable
	Data    []Market `json:"data"`
}

// UnixTimestamp is a time.Time that unmarshals from Unix timestamp (seconds)
type UnixTimestamp struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler for Unix timestamp
func (ut *UnixTimestamp) UnmarshalJSON(data []byte) error {
	var timestamp int64
	if err := json.Unmarshal(data, &timestamp); err == nil {
		ut.Time = time.Unix(timestamp, 0)
		return nil
	}

	var timestampStr string
	if err := json.Unmarshal(data, &timestampStr); err == nil {
		if timestampStr == "" {
			ut.Time = time.Time{}
			return nil
		}
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return err
		}
		ut.Time = timestamp
		return nil
	}

	return json.Unmarshal(data, &ut.Time)
}

// MarshalJSON implements json.Marshaler for Unix timestamp
func (ut UnixTimestamp) MarshalJSON() ([]byte, error) {
	if ut.Time.IsZero() {
		return []byte("0"), nil
	}
	return []byte(fmt.Sprintf("%d", ut.Time.Unix())), nil
}
