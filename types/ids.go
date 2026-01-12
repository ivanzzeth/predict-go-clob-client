package types

// MarketID represents a market identifier
type MarketID string

// String returns the string representation of the ID
func (id MarketID) String() string {
	return string(id)
}

// CategoryID represents a category identifier
type CategoryID string

// String returns the string representation of the ID
func (id CategoryID) String() string {
	return string(id)
}

// PositionID represents a position identifier
type PositionID string

// String returns the string representation of the ID
func (id PositionID) String() string {
	return string(id)
}

// TokenID represents a token identifier
type TokenID string

// String returns the string representation of the ID
func (id TokenID) String() string {
	return string(id)
}

// Address represents an Ethereum address
type Address string

// String returns the string representation of the address
func (a Address) String() string {
	return string(a)
}

// TransactionHash represents a transaction hash
type TransactionHash string

// String returns the string representation of the hash
func (h TransactionHash) String() string {
	return string(h)
}
