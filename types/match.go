package types

import (
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/shopspring/decimal"
)

// OutcomeStatus represents the status of an outcome
type OutcomeStatus string

const (
	OutcomeStatusWon  OutcomeStatus = "WON"
	OutcomeStatusLost OutcomeStatus = "LOST"
)

// String returns the string representation of the status
func (s OutcomeStatus) String() string {
	return string(s)
}

// FeeRateBps represents fee rate in basis points, can be unmarshaled from int or string
type FeeRateBps int

// UnmarshalJSON implements json.Unmarshaler to handle both int and string
func (f *FeeRateBps) UnmarshalJSON(data []byte) error {
	var ios IntegerOrString
	if err := json.Unmarshal(data, &ios); err != nil {
		return err
	}
	*f = FeeRateBps(ios.Int64())
	return nil
}

// MatchEventOutcome represents an outcome in a match event
type MatchEventOutcome struct {
	Name      string        `json:"name"`
	IndexSet  int           `json:"indexSet"`
	OnChainID TokenID       `json:"onChainId"`
	Status    OutcomeStatus `json:"status,omitempty"` // nullable
}

// MatchEventResolution represents the resolution of a market
type MatchEventResolution struct {
	Name      string        `json:"name"`
	IndexSet  int           `json:"indexSet"`
	OnChainID TokenID       `json:"onChainId"`
	Status    OutcomeStatus `json:"status,omitempty"` // nullable
}

// MatchEventMarket represents market information in a match event
type MatchEventMarket struct {
	ID                     MarketID              `json:"id"`
	ImageURL               string                `json:"imageUrl"`
	Title                  string                `json:"title"`
	Question               string                `json:"question"`
	Description            string                `json:"description"`
	Status                 MarketStatus          `json:"status"`
	IsNegRisk              bool                  `json:"isNegRisk"`
	IsYieldBearing         bool                  `json:"isYieldBearing"`
	FeeRateBps             FeeRateBps            `json:"feeRateBps"`           // Handles int/string via UnmarshalJSON
	Resolution             *MatchEventResolution `json:"resolution,omitempty"` // nullable
	OracleQuestionID       common.Hash          `json:"oracleQuestionId"`    // common.Hash implements json.Unmarshaler
	ConditionID            common.Hash          `json:"conditionId"`          // common.Hash implements json.Unmarshaler
	ResolverAddress        common.Address       `json:"resolverAddress"`     // common.Address implements json.Unmarshaler
	Outcomes               []MatchEventOutcome  `json:"outcomes"`
	QuestionIndex          *int                 `json:"questionIndex,omitempty"` // nullable
	SpreadThreshold        float64              `json:"spreadThreshold"`
	ShareThreshold         float64              `json:"shareThreshold"`
	PolymarketConditionIDs []common.Hash        `json:"polymarketConditionIds"` // []common.Hash, each implements json.Unmarshaler
	KalshiMarketTicker     *string               `json:"kalshiMarketTicker,omitempty"` // nullable
	CategorySlug           string                `json:"categorySlug"`
	CreatedAt              time.Time             `json:"createdAt"`
	DecimalPrecision       int                   `json:"decimalPrecision"`
}

// MatchEventMarket doesn't need custom UnmarshalJSON
// MarketID handles int/string via its own UnmarshalJSON
// MarketStatus is string-based, auto-handled
// ResolverAddress implements json.Unmarshaler, auto-handled
// FeeRateBps handles int/string via its own UnmarshalJSON

// QuoteType represents the type of quote (Ask or Bid)
type QuoteType string

const (
	QuoteTypeAsk QuoteType = "Ask"
	QuoteTypeBid QuoteType = "Bid"
)

// String returns the string representation of the quote type
func (q QuoteType) String() string {
	return string(q)
}

// MatchEventQuote represents a quote in a match event (for taker or makers)
type MatchEventQuote struct {
	QuoteType QuoteType         `json:"quoteType"`
	Amount    decimal.Decimal   `json:"-"`      // Human readable decimal (converted from wei)
	Price     decimal.Decimal   `json:"-"`      // Human readable decimal (converted from wei)
	RawAmount string            `json:"amount"` // Raw wei amount as string
	RawPrice  string            `json:"price"`  // Raw wei price as string
	Outcome   MatchEventOutcome `json:"outcome"`
	Signer    string            `json:"signer"`
}

// UnmarshalJSON implements custom unmarshaling for MatchEventQuote to convert wei amounts to decimals
func (q *MatchEventQuote) UnmarshalJSON(data []byte) error {
	type Alias MatchEventQuote
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

// MatchEvent represents a single order match event
type MatchEvent struct {
	Market           MatchEventMarket  `json:"market"`
	Taker            MatchEventQuote   `json:"taker"`
	AmountFilled     decimal.Decimal   `json:"-"`             // Human readable decimal (converted from wei)
	PriceExecuted    decimal.Decimal   `json:"-"`             // Human readable decimal (converted from wei)
	RawAmountFilled  string            `json:"amountFilled"`  // Raw wei amount as string
	RawPriceExecuted string            `json:"priceExecuted"` // Raw wei price as string
	Makers           []MatchEventQuote `json:"makers"`
	TransactionHash  common.Hash       `json:"transactionHash"` // common.Hash implements json.Unmarshaler
	ExecutedAt       time.Time         `json:"executedAt"`
}

// UnmarshalJSON implements custom unmarshaling for MatchEvent to convert wei amounts to decimals
func (m *MatchEvent) UnmarshalJSON(data []byte) error {
	type Alias MatchEvent
	aux := &struct {
		AmountFilled  string `json:"amountFilled"`
		PriceExecuted string `json:"priceExecuted"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Save raw values and convert AmountFilled from wei to decimal
	if aux.AmountFilled != "" {
		m.RawAmountFilled = aux.AmountFilled
		amountWei, err := decimal.NewFromString(aux.AmountFilled)
		if err == nil {
			m.AmountFilled = amountWei.Shift(-constants.TokenDecimals)
		}
	}

	// Save raw values and convert PriceExecuted from wei to decimal
	if aux.PriceExecuted != "" {
		m.RawPriceExecuted = aux.PriceExecuted
		priceWei, err := decimal.NewFromString(aux.PriceExecuted)
		if err == nil {
			m.PriceExecuted = priceWei.Shift(-constants.TokenDecimals)
		}
	}

	return nil
}

// GetOrderMatchesOptions represents options for getting order match events
type GetOrderMatchesOptions struct {
	First           *int   // Pagination: number of results to return
	After           string // Pagination: cursor for next page
	CategoryID      CategoryID
	MarketID        MarketID
	MinValueUsdtWei string // Minimum value in USDT wei
	SignerAddress   string // Filter by signer address
	IsSignerMaker   *bool  // Filter by whether signer is maker (true) or taker (false), nil for both
}

// MatchEventResponse represents the response from GetOrderMatches API
type MatchEventResponse struct {
	Success bool         `json:"success"`
	Cursor  *string      `json:"cursor,omitempty"` // nullable
	Data    []MatchEvent `json:"data"`
}
