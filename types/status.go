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
	MarketStatusRegistered MarketStatus = "REGISTERED"
	MarketStatusOpen       MarketStatus = "OPEN"
	MarketStatusResolved   MarketStatus = "RESOLVED"
	MarketStatusCancelled  MarketStatus = "CANCELLED"
)

// String returns the string representation of the status
func (s MarketStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s MarketStatus) IsValid() bool {
	return s == MarketStatusRegistered || s == MarketStatusOpen || 
		   s == MarketStatusResolved || s == MarketStatusCancelled
}

// ReferralStatus represents the status of a referral
type ReferralStatus string

const (
	ReferralStatusActive   ReferralStatus = "ACTIVE"
	ReferralStatusInactive ReferralStatus = "INACTIVE"
)

// String returns the string representation of the status
func (s ReferralStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s ReferralStatus) IsValid() bool {
	return s == ReferralStatusActive || s == ReferralStatusInactive
}
