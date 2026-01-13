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

// Market represents a prediction market
type Market struct {
	ID             MarketID     `json:"id"`
	Title          string       `json:"title"`
	Question       string       `json:"question"`
	Description    string       `json:"description,omitempty"`
	Status         MarketStatus `json:"status"`
	FeeRateBps     string       `json:"feeRateBps"`
	IsNegRisk      bool         `json:"isNegRisk"`
	IsYieldBearing bool         `json:"isYieldBearing"`
	TokenID        TokenID      `json:"tokenId,omitempty"`
	OutcomeTokenID TokenID      `json:"outcomeTokenId,omitempty"`
	Outcomes       []Outcome    `json:"outcomes,omitempty"`
	CreatedAt      time.Time    `json:"createdAt,omitempty"`
	UpdatedAt      time.Time    `json:"updatedAt,omitempty"`
}

// UnmarshalJSON implements custom unmarshaling for Market to handle ID and feeRateBps as number or string
func (m *Market) UnmarshalJSON(data []byte) error {
	type Alias Market
	aux := &struct {
		ID             interface{} `json:"id"`
		Status         interface{} `json:"status"`
		TokenID        interface{} `json:"tokenId,omitempty"`
		OutcomeTokenID interface{} `json:"outcomeTokenId,omitempty"`
		FeeRateBps     interface{} `json:"feeRateBps"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert ID to MarketID
	switch v := aux.ID.(type) {
	case float64:
		m.ID = MarketID(fmt.Sprintf("%.0f", v))
	case string:
		m.ID = MarketID(v)
	default:
		m.ID = MarketID(fmt.Sprintf("%v", v))
	}

	// Convert Status to MarketStatus
	if aux.Status != nil {
		if statusStr, ok := aux.Status.(string); ok {
			m.Status = MarketStatus(statusStr)
		} else {
			m.Status = MarketStatus(fmt.Sprintf("%v", aux.Status))
		}
	}

	// Convert TokenID
	if aux.TokenID != nil {
		if tokenIDStr, ok := aux.TokenID.(string); ok {
			m.TokenID = TokenID(tokenIDStr)
		} else {
			m.TokenID = TokenID(fmt.Sprintf("%v", aux.TokenID))
		}
	}

	// Convert OutcomeTokenID
	if aux.OutcomeTokenID != nil {
		if tokenIDStr, ok := aux.OutcomeTokenID.(string); ok {
			m.OutcomeTokenID = TokenID(tokenIDStr)
		} else {
			m.OutcomeTokenID = TokenID(fmt.Sprintf("%v", aux.OutcomeTokenID))
		}
	}

	// Convert feeRateBps to string
	switch v := aux.FeeRateBps.(type) {
	case float64:
		m.FeeRateBps = fmt.Sprintf("%.0f", v)
	case string:
		m.FeeRateBps = v
	default:
		m.FeeRateBps = fmt.Sprintf("%v", v)
	}

	return nil
}

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
	CategoryID CategoryID
	Limit      int
	Offset     int
	Status     MarketStatus // "OPEN" or "RESOLVED"
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
