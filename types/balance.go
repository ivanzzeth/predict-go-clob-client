package types

import (
	"github.com/shopspring/decimal"
)

// Balance represents collateral balance information
// Calculated fields: Total, Locked, Available
type Balance struct {
	Total     decimal.Decimal `json:"-"` // Calculated: total collateral balance from blockchain
	Locked    decimal.Decimal `json:"-"` // Calculated: locked collateral from OPEN BUY orders
	Available decimal.Decimal `json:"-"` // Calculated: available collateral = Total - Locked
}
