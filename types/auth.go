package types

// AuthMessageResponse represents the response from GET /v1/auth/message
type AuthMessageResponse struct {
	Message string `json:"message"`
}

// AuthRequest represents the request body for POST /v1/auth
type AuthRequest struct {
	Signer    string `json:"signer"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

// JWTTokenResponse represents the response from POST /v1/auth
type JWTTokenResponse struct {
	Token string `json:"token"`
}
