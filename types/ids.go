package types

import (
	"encoding/json"
	"fmt"
	"math/big"
)

// IntegerOrString represents a value that can be unmarshaled from int or string
// It's based on big.Int to handle large integers
type IntegerOrString big.Int

// UnmarshalJSON implements json.Unmarshaler to handle both int and string
func (ios *IntegerOrString) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		bigInt := new(big.Int)
		if _, ok := bigInt.SetString(str, 10); !ok {
			return fmt.Errorf("cannot parse IntegerOrString from string: %s", str)
		}
		*ios = IntegerOrString(*bigInt)
		return nil
	}

	// Try to unmarshal as int
	var num int64
	if err := json.Unmarshal(data, &num); err == nil {
		*ios = IntegerOrString(*big.NewInt(num))
		return nil
	}

	// Try to unmarshal as float64
	var numFloat float64
	if err := json.Unmarshal(data, &numFloat); err == nil {
		*ios = IntegerOrString(*big.NewInt(int64(numFloat)))
		return nil
	}

	return fmt.Errorf("cannot unmarshal IntegerOrString: %s", string(data))
}

// String returns the string representation
func (ios *IntegerOrString) String() string {
	return (*big.Int)(ios).String()
}

// Int64 returns the int64 value (may overflow for large numbers)
func (ios *IntegerOrString) Int64() int64 {
	return (*big.Int)(ios).Int64()
}

// BigInt returns the underlying *big.Int
func (ios *IntegerOrString) BigInt() *big.Int {
	return (*big.Int)(ios)
}

// MarketID represents a market identifier
type MarketID string

// String returns the string representation of the ID
func (id MarketID) String() string {
	return string(id)
}

// UnmarshalJSON implements json.Unmarshaler to handle both int and string
func (id *MarketID) UnmarshalJSON(data []byte) error {
	var ios IntegerOrString
	if err := json.Unmarshal(data, &ios); err != nil {
		return err
	}
	*id = MarketID(ios.String())
	return nil
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
