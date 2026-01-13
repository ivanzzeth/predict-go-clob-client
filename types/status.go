package types

// CategoryStatus represents the status of a category
type CategoryStatus string

const (
	CategoryStatusOpen     CategoryStatus = "OPEN"
	CategoryStatusResolved CategoryStatus = "RESOLVED"
)

// String returns the string representation of the status
func (s CategoryStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s CategoryStatus) IsValid() bool {
	return s == CategoryStatusOpen || s == CategoryStatusResolved
}

// MarketStatus represents the status of a market
type MarketStatus string

const (
	MarketStatusRegistered    MarketStatus = "REGISTERED"
	MarketStatusPriceProposed MarketStatus = "PRICE_PROPOSED"
	MarketStatusPriceDisputed MarketStatus = "PRICE_DISPUTED"
	MarketStatusPaused        MarketStatus = "PAUSED"
	MarketStatusUnpaused      MarketStatus = "UNPAUSED"
	MarketStatusResolved      MarketStatus = "RESOLVED"
	// Legacy status values (kept for backward compatibility)
	MarketStatusOpen      MarketStatus = "OPEN"
	MarketStatusCancelled MarketStatus = "CANCELLED"
)

// String returns the string representation of the status
func (s MarketStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s MarketStatus) IsValid() bool {
	return s == MarketStatusRegistered || s == MarketStatusPriceProposed ||
		s == MarketStatusPriceDisputed || s == MarketStatusPaused ||
		s == MarketStatusUnpaused || s == MarketStatusResolved ||
		s == MarketStatusOpen || s == MarketStatusCancelled
}

// ReferralStatus represents the status of a referral
type ReferralStatus string

const (
	ReferralStatusLocked   ReferralStatus = "LOCKED"
	ReferralStatusUnlocked ReferralStatus = "UNLOCKED"
)

// String returns the string representation of the status
func (s ReferralStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s ReferralStatus) IsValid() bool {
	return s == ReferralStatusLocked || s == ReferralStatusUnlocked
}

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusOpen        OrderStatus = "OPEN"
	OrderStatusFilled      OrderStatus = "FILLED"
	OrderStatusExpired     OrderStatus = "EXPIRED"
	OrderStatusCancelled   OrderStatus = "CANCELLED"
	OrderStatusInvalidated OrderStatus = "INVALIDATED"
)

// String returns the string representation of the status
func (s OrderStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s OrderStatus) IsValid() bool {
	return s == OrderStatusOpen || s == OrderStatusFilled ||
		s == OrderStatusExpired || s == OrderStatusCancelled ||
		s == OrderStatusInvalidated
}
