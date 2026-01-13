package types

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// Account represents user account information
type Account struct {
	Name     string    `json:"name"`
	Address  common.Address `json:"address"`
	ImageURL *string   `json:"imageUrl,omitempty"` // nullable
	Referral *Referral `json:"referral,omitempty"`
}

// Referral represents referral information
type Referral struct {
	Code   *string        `json:"code"` // nullable
	Status ReferralStatus `json:"status"`
}

// UnmarshalJSON implements custom unmarshaling for Account to handle null values and Address conversion
func (a *Account) UnmarshalJSON(data []byte) error {
	type Alias Account
	aux := &struct {
		Address  interface{} `json:"address"`
		ImageURL interface{} `json:"imageUrl,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle Address: convert string to common.Address
	if aux.Address != nil {
		if addrStr, ok := aux.Address.(string); ok {
			a.Address = common.HexToAddress(addrStr)
		} else {
			return fmt.Errorf("invalid address type: %T", aux.Address)
		}
	}

	// Handle ImageURL: can be null, string, or missing
	if aux.ImageURL != nil {
		if imgStr, ok := aux.ImageURL.(string); ok && imgStr != "" {
			a.ImageURL = &imgStr
		}
		// If it's null or empty string, leave ImageURL as nil
	}

	return nil
}

// UnmarshalJSON implements custom unmarshaling for Referral to handle null values
func (r *Referral) UnmarshalJSON(data []byte) error {
	type Alias Referral
	aux := &struct {
		Code   interface{} `json:"code"`
		Status interface{} `json:"status"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle Code: can be null or string
	if aux.Code != nil {
		if codeStr, ok := aux.Code.(string); ok && codeStr != "" {
			r.Code = &codeStr
		}
		// If it's null or empty string, leave Code as nil
	}

	// Handle Status: convert to ReferralStatus
	if aux.Status != nil {
		if statusStr, ok := aux.Status.(string); ok {
			r.Status = ReferralStatus(statusStr)
		} else {
			r.Status = ReferralStatus(fmt.Sprintf("%v", aux.Status))
		}
	}

	return nil
}
