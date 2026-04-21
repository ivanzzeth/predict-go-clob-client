package types

// SetReferralRequest represents the request body for POST /v1/account/referral
type SetReferralRequest struct {
	Data SetReferralData `json:"data"`
}

// SetReferralData contains the referral code to set
type SetReferralData struct {
	ReferralCode string `json:"referralCode"`
}

// SetReferralResponse represents the response from POST /v1/account/referral
type SetReferralResponse struct {
	Success bool `json:"success"`
}
