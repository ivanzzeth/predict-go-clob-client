package types

// Account represents user account information
type Account struct {
	Name     string    `json:"name"`
	Address  Address   `json:"address"`
	ImageURL string    `json:"imageUrl,omitempty"`
	Referral *Referral `json:"referral,omitempty"`
}

// Referral represents referral information
type Referral struct {
	Code   string         `json:"code"`
	Status ReferralStatus `json:"status"`
}
