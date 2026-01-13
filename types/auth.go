package types

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
)

// AuthMessageResponse represents the response from GET /v1/auth/message
type AuthMessageResponse struct {
	Message string `json:"message"`
}

// AuthRequest represents the request body for POST /v1/auth
type AuthRequest struct {
	Signer    common.Address `json:"-"`        // Use strong type, serialize to string
	Message   string         `json:"message"`
	Signature string         `json:"signature"`
}

// MarshalJSON implements json.Marshaler to serialize Signer as hex string
func (r AuthRequest) MarshalJSON() ([]byte, error) {
	type Alias AuthRequest
	return json.Marshal(&struct {
		Signer string `json:"signer"`
		*Alias
	}{
		Signer: r.Signer.Hex(),
		Alias:  (*Alias)(&r),
	})
}

// JWTTokenResponse represents the response from POST /v1/auth
type JWTTokenResponse struct {
	Token string `json:"token"`
}
