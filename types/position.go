package types

// Position represents a user position
type Position struct {
	ID       PositionID `json:"id"`
	Market   Market     `json:"market"`
	Amount   string     `json:"amount"`
	ValueUsd string     `json:"valueUsd,omitempty"`
}

// GetPositionsOptions represents options for getting positions
type GetPositionsOptions struct {
	MarketID MarketID
	First    int    // Limit for pagination
	After    string // Cursor for pagination
}
