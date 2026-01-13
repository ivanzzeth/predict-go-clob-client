package types

import (
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/shopspring/decimal"
)

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeCreate         ActivityType = "CREATE"
	ActivityTypeMatchSuccess   ActivityType = "MATCH_SUCCESS"
	ActivityTypeCancel         ActivityType = "CANCEL"
	ActivityTypeInvalidate     ActivityType = "INVALIDATE"
	ActivityTypeNoMarketMatch  ActivityType = "NO_MARKET_MATCH"
	ActivityTypeExpired        ActivityType = "EXPIRED"
	ActivityTypeConvert        ActivityType = "CONVERT"
	ActivityTypeMerge          ActivityType = "MERGE"
	ActivityTypeSplit          ActivityType = "SPLIT"
	ActivityTypeRedeem         ActivityType = "REDEEM"
)

// String returns the string representation of the activity type
func (a ActivityType) String() string {
	return string(a)
}

// ActivityOrderQuote represents a simplified order quote in activity
type ActivityOrderQuote struct {
	QuoteType QuoteType       `json:"quoteType"`
	Amount    decimal.Decimal `json:"-"` // Human readable decimal (converted from wei)
	Price     decimal.Decimal `json:"-"` // Human readable decimal (converted from wei)
	RawAmount string          `json:"-"` // Raw wei amount as string
	RawPrice  string          `json:"-"` // Raw wei price as string
}

// UnmarshalJSON implements custom unmarshaling for ActivityOrderQuote to convert wei amounts to decimals
func (q *ActivityOrderQuote) UnmarshalJSON(data []byte) error {
	type Alias ActivityOrderQuote
	aux := &struct {
		Amount string `json:"amount"`
		Price  string `json:"price"`
		*Alias
	}{
		Alias: (*Alias)(q),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Save raw values and convert Amount from wei to decimal
	if aux.Amount != "" {
		q.RawAmount = aux.Amount
		amountWei, err := decimal.NewFromString(aux.Amount)
		if err == nil {
			q.Amount = amountWei.Shift(-constants.TokenDecimals)
		}
	}

	// Save raw values and convert Price from wei to decimal
	if aux.Price != "" {
		q.RawPrice = aux.Price
		priceWei, err := decimal.NewFromString(aux.Price)
		if err == nil {
			q.Price = priceWei.Shift(-constants.TokenDecimals)
		}
	}

	return nil
}

// Activity represents a single account activity item
type Activity struct {
	Name            ActivityType        `json:"name"`
	CreatedAt       time.Time           `json:"createdAt"`
	TransactionHash *common.Hash        `json:"-"`                         // nullable, converted from string
	AmountFilled    decimal.Decimal     `json:"-"`                         // Human readable decimal (converted from wei)
	PriceExecuted   decimal.Decimal    `json:"-"`                         // Human readable decimal (converted from wei)
	RawAmountFilled *string             `json:"-"`                         // Raw wei amount as string (nullable)
	RawPriceExecuted *string            `json:"-"`                         // Raw wei price as string (nullable)
	Order           *ActivityOrderQuote `json:"order,omitempty"`          // nullable
	Market          MatchEventMarket    `json:"market"`
	Outcome         *MatchEventOutcome  `json:"outcome,omitempty"` // nullable
}

// UnmarshalJSON implements custom unmarshaling for Activity to convert wei amounts to decimals and transactionHash to common.Hash
func (a *Activity) UnmarshalJSON(data []byte) error {
	type Alias Activity
	aux := &struct {
		TransactionHash interface{} `json:"transactionHash"` // nullable, can be null or string
		AmountFilled    *string     `json:"amountFilled"`    // nullable
		PriceExecuted   *string     `json:"priceExecuted"`   // nullable
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle TransactionHash: convert string to common.Hash
	if aux.TransactionHash != nil {
		if hashStr, ok := aux.TransactionHash.(string); ok && hashStr != "" {
			hash := common.HexToHash(hashStr)
			a.TransactionHash = &hash
		}
		// If it's null or empty string, leave TransactionHash as nil
	}

	// Save raw values and convert AmountFilled from wei to decimal
	if aux.AmountFilled != nil && *aux.AmountFilled != "" {
		a.RawAmountFilled = aux.AmountFilled
		amountWei, err := decimal.NewFromString(*aux.AmountFilled)
		if err == nil {
			a.AmountFilled = amountWei.Shift(-constants.TokenDecimals)
		}
	}

	// Save raw values and convert PriceExecuted from wei to decimal
	if aux.PriceExecuted != nil && *aux.PriceExecuted != "" {
		a.RawPriceExecuted = aux.PriceExecuted
		priceWei, err := decimal.NewFromString(*aux.PriceExecuted)
		if err == nil {
			a.PriceExecuted = priceWei.Shift(-constants.TokenDecimals)
		}
	}

	return nil
}

// GetActivityOptions represents options for getting account activity
type GetActivityOptions struct {
	First *int   // Pagination: number of results to return
	After string // Pagination: cursor for next page
}

// ActivityResponse represents the response from GetActivity API
type ActivityResponse struct {
	Success bool       `json:"success"`
	Cursor  *string    `json:"cursor,omitempty"` // nullable
	Data    []Activity `json:"data"`
}
