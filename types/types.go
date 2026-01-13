package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
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
type MarketStats struct {
	Volume       string `json:"volume"`
	OpenInterest string `json:"openInterest"`
	BidPrice     string `json:"bidPrice"`
	AskPrice     string `json:"askPrice"`
	LastPrice    string `json:"lastPrice"`
	TraderCount  int    `json:"traderCount,omitempty"`
}

// OrderbookLevel represents a single level in the orderbook
type OrderbookLevel struct {
	Price  string `json:"price"`  // Price per share (as string)
	Amount string `json:"amount"` // Order amount (as string)
}

// Orderbook represents the orderbook for a market.
//
// Sorting rules (automatically applied during JSON unmarshaling):
//   - Bids: sorted in descending order by price (highest price first, best bid at index 0)
//     Example: [0.60, 0.55, 0.50, ...] where 0.60 is the best bid
//   - Asks: sorted in ascending order by price (lowest price first, best ask at index 0)
//     Example: [0.40, 0.45, 0.50, ...] where 0.40 is the best ask
//
// This ensures that:
//   - Bids[0] is always the best bid (highest buy price)
//   - Asks[0] is always the best ask (lowest sell price)
//   - The spread can be calculated as Asks[0].Price - Bids[0].Price
type Orderbook struct {
	Bids []OrderbookLevel `json:"bids"` // Buy orders, sorted descending by price
	Asks []OrderbookLevel `json:"asks"` // Sell orders, sorted ascending by price
}

// UnmarshalJSON implements custom unmarshaling for Orderbook to handle array format
func (o *Orderbook) UnmarshalJSON(data []byte) error {
	var raw struct {
		Bids interface{} `json:"bids"`
		Asks interface{} `json:"asks"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Convert bids
	if bidsArray, ok := raw.Bids.([]interface{}); ok {
		o.Bids = make([]OrderbookLevel, 0, len(bidsArray))
		for _, item := range bidsArray {
			if arr, ok := item.([]interface{}); ok && len(arr) >= 2 {
				// Format: [price, amount]
				o.Bids = append(o.Bids, OrderbookLevel{
					Price:  fmt.Sprintf("%v", arr[0]),
					Amount: fmt.Sprintf("%v", arr[1]),
				})
			} else if m, ok := item.(map[string]interface{}); ok {
				// Format: {"price": "...", "amount": "..."}
				price, _ := m["price"].(string)
				amount, _ := m["amount"].(string)
				if price == "" {
					if priceVal, ok := m["pricePerShare"].(string); ok {
						price = priceVal
					}
				}
				if amount == "" {
					if amountVal, ok := m["makerAmount"].(string); ok {
						amount = amountVal
					}
				}
				o.Bids = append(o.Bids, OrderbookLevel{
					Price:  price,
					Amount: amount,
				})
			}
		}
	}

	// Convert asks
	if asksArray, ok := raw.Asks.([]interface{}); ok {
		o.Asks = make([]OrderbookLevel, 0, len(asksArray))
		for _, item := range asksArray {
			if arr, ok := item.([]interface{}); ok && len(arr) >= 2 {
				// Format: [price, amount]
				o.Asks = append(o.Asks, OrderbookLevel{
					Price:  fmt.Sprintf("%v", arr[0]),
					Amount: fmt.Sprintf("%v", arr[1]),
				})
			} else if m, ok := item.(map[string]interface{}); ok {
				// Format: {"price": "...", "amount": "..."}
				price, _ := m["price"].(string)
				amount, _ := m["amount"].(string)
				if price == "" {
					if priceVal, ok := m["pricePerShare"].(string); ok {
						price = priceVal
					}
				}
				if amount == "" {
					if amountVal, ok := m["makerAmount"].(string); ok {
						amount = amountVal
					}
				}
				o.Asks = append(o.Asks, OrderbookLevel{
					Price:  price,
					Amount: amount,
				})
			}
		}
	}

	// Sort bids in descending order (highest price first, best bid at index 0)
	sort.Slice(o.Bids, func(i, j int) bool {
		priceI, errI := strconv.ParseFloat(o.Bids[i].Price, 64)
		priceJ, errJ := strconv.ParseFloat(o.Bids[j].Price, 64)
		if errI != nil || errJ != nil {
			// If parsing fails, keep original order
			return false
		}
		return priceI > priceJ
	})

	// Sort asks in ascending order (lowest price first, best ask at index 0)
	sort.Slice(o.Asks, func(i, j int) bool {
		priceI, errI := strconv.ParseFloat(o.Asks[i].Price, 64)
		priceJ, errJ := strconv.ParseFloat(o.Asks[j].Price, 64)
		if errI != nil || errJ != nil {
			// If parsing fails, keep original order
			return false
		}
		return priceI < priceJ
	})

	return nil
}

// Sale represents a trade/sale
type Sale struct {
	TransactionHash common.Hash    `json:"transactionHash"`
	Price           string         `json:"price"`
	Amount          string         `json:"amount"`
	Seller          common.Address `json:"seller"`
	Buyer           common.Address `json:"buyer"`
	Timestamp       time.Time      `json:"timestamp"`
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
