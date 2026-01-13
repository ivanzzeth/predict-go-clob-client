package types

import (
	"encoding/json"

	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/shopspring/decimal"
)

// Position represents a user position
type Position struct {
	ID          PositionID      `json:"id"`
	Market      Market          `json:"market"`
	Outcome     CategoryMarketOutcome `json:"outcome"` // Outcome for this position
	Amount      decimal.Decimal `json:"-"`        // Human readable decimal (converted from wei) - Total amount
	ValueUsd    decimal.Decimal `json:"-"`        // Human readable decimal
	RawAmount   string          `json:"amount"`   // Raw wei amount as string
	RawValueUsd string          `json:"valueUsd"` // Raw value USD as string
	Total       decimal.Decimal `json:"-"`        // Calculated: total amount (same as Amount, for clarity)
	Locked      decimal.Decimal `json:"-"`        // Calculated: locked amount from OPEN SELL orders
	Available   decimal.Decimal `json:"-"`        // Calculated: available amount = Total - Locked
}

// UnmarshalJSON implements custom unmarshaling for Position to convert wei amounts to decimals
func (p *Position) UnmarshalJSON(data []byte) error {
	type Alias Position
	aux := &struct {
		Amount   string `json:"amount"`
		ValueUsd string `json:"valueUsd"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Save raw values and convert Amount from wei to decimal
	if aux.Amount != "" {
		p.RawAmount = aux.Amount
		amountWei, err := decimal.NewFromString(aux.Amount)
		if err == nil {
			p.Amount = amountWei.Shift(-constants.TokenDecimals)
		}
	}

	// Save raw values and convert ValueUsd (already in decimal format, no conversion needed)
	if aux.ValueUsd != "" {
		p.RawValueUsd = aux.ValueUsd
		valueUsd, err := decimal.NewFromString(aux.ValueUsd)
		if err == nil {
			p.ValueUsd = valueUsd
		}
	}

	// Initialize calculated fields (will be set by GetPositions)
	p.Total = p.Amount
	p.Locked = decimal.Zero
	p.Available = p.Amount

	return nil
}

// GetPositionsOptions represents options for getting positions
type GetPositionsOptions struct {
	MarketID MarketID
	First    int    // Limit for pagination
	After    string // Cursor for pagination
}
