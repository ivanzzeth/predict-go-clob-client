package predictclob

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/errs"
	"github.com/ivanzzeth/predict-go-clob-client/types"
)

// GetAuthMessage gets the authentication message from Predict.fun API that needs to be signed.
//
// This is the first step in the authentication flow. The returned message must be signed
// using the user's private key (EIP-191 personal sign), and the signature should be sent
// to GetJWTTokenFromSignature to obtain a JWT token.
//
// This endpoint requires an API key to be set in the client.
//
// Returns:
//   - *AuthMessageResponse: The response containing the message to sign
//   - error: Returns error if:
//   - API key is missing (validated before request)
//   - HTTP request fails (network error, timeout, etc.)
//   - Response parsing fails (invalid JSON)
//   - API returns success=false (check response.Message for details)
func (c *Client) GetAuthMessage() (*types.AuthMessageResponse, error) {
	respBody, err := c.doRequest("GET", constants.EndpointAuthMessage, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth message: %w", err)
	}

	var response types.APIBaseResponse[types.AuthMessageResponse]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse auth message response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	return &response.Data, nil
}

// SignMessage signs a message using a PersonalSigner interface.
//
// This function implements EIP-191 personal sign, which is the standard way to sign
// messages in Ethereum. The message is prefixed with "\x19Ethereum Signed Message:\n"
// before signing.
//
// Parameters:
//   - message: The message to sign (plain text string)
//   - signer: The signer that implements PersonalSigner and AddressGetter interfaces
//
// Returns:
//   - string: The signature as a hex string with 0x prefix (65 bytes, r + s + v)
//   - common.Address: The Ethereum address from the signer
//   - error: Returns error if:
//   - Message is empty
//   - Signer is nil
//   - Signer does not implement required interfaces
//   - Message signing fails (crypto operation error)
func SignMessage(message string, signer Signer) (string, common.Address, error) {
	// Validate input
	if message == "" {
		return "", common.Address{}, errs.ErrEmptyMessage
	}
	if signer == nil {
		return "", common.Address{}, fmt.Errorf("signer cannot be nil")
	}

	// Get address from signer
	address := signer.GetAddress()
	if address == (common.Address{}) {
		return "", common.Address{}, fmt.Errorf("signer returned zero address")
	}

	// Sign message using personal sign (EIP-191)
	signature, err := signer.PersonalSign(message)
	if err != nil {
		return "", common.Address{}, fmt.Errorf("failed to sign message: %w", err)
	}

	// Convert signature to hex string with 0x prefix
	signatureHex := "0x" + hex.EncodeToString(signature)

	return signatureHex, address, nil
}

// GetJWTTokenFromSignature exchanges a signed message for a JWT token from Predict.fun API.
//
// This is the second step in the authentication flow. After signing the message obtained
// from GetAuthMessage, call this method with the signature to receive a JWT token.
// The JWT token should be set in the client using SetJWTToken for subsequent authenticated requests.
//
// This endpoint requires an API key to be set in the client.
// The signerAddress will be automatically converted to checksum format (EIP-55).
//
// Parameters:
//   - signerAddress: The Ethereum address of the signer (will be converted to checksum format)
//   - message: The original message that was signed (must match the message from GetAuthMessage)
//   - signature: The signature of the message (hex string with 0x prefix, 65 bytes: r + s + v)
//
// Returns:
//   - *JWTTokenResponse: The response containing the JWT token
//   - error: Returns error if:
//   - API key is missing (validated before request)
//   - Invalid address format (cannot be converted to common.Address)
//   - Request marshaling fails (JSON encoding error)
//   - HTTP request fails (network error, timeout, etc.)
//   - Response parsing fails (invalid JSON)
//   - API returns success=false (authentication failed, check response.Message)
func (c *Client) GetJWTTokenFromSignature(signerAddress string, message string, signature string) (*types.JWTTokenResponse, error) {
	// Validate inputs
	if signerAddress == "" {
		return nil, errs.NewInvalidAddressError("address cannot be empty")
	}
	if message == "" {
		return nil, errs.ErrEmptyMessage
	}
	if signature == "" {
		return nil, errs.NewInvalidSignatureError("signature cannot be empty")
	}

	// Validate signature format (should be 0x + 130 hex chars = 65 bytes)
	if len(signature) < 2 || signature[0:2] != "0x" {
		return nil, errs.NewInvalidSignatureError("signature must start with 0x prefix")
	}
	if len(signature) != 132 { // 0x + 130 hex chars
		return nil, errs.NewInvalidSignatureLengthError(len(signature), 132)
	}

	// Ensure address is checksum format
	address := common.HexToAddress(signerAddress)
	if address == (common.Address{}) {
		return nil, errs.NewInvalidAddressError(signerAddress)
	}

	// Build request body
	authReq := types.AuthRequest{
		Signer:    address, // Use common.Address directly
		Message:   message,
		Signature: signature,
	}

	bodyBytes, err := json.Marshal(authReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	respBody, err := c.doRequest("POST", constants.EndpointAuth, bodyBytes, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT token: %w", err)
	}

	var response types.APIBaseResponse[types.JWTTokenResponse]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JWT token response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	return &response.Data, nil
}

// Authenticate performs the complete authentication flow in a single call.
//
// This is a convenience method that combines the following steps:
// 1. Get auth message from the API (GetAuthMessage)
// 2. Sign the message with the client's signer (SignMessage)
// 3. Exchange the signature for a JWT token (GetJWTTokenFromSignature)
// 4. Automatically set the JWT token in the client (SetJWTToken)
//
// This method uses EIP-191 personal sign for message signing.
// The signer must be set in the client using WithSigner option when creating the client.
//
// Returns:
//   - string: The JWT token (also automatically set in the client)
//   - common.Address: The Ethereum address from the client's signer
//   - error: Returns error if any step in the authentication flow fails:
//   - API key is missing (validated before request)
//   - Client signer is nil (signer must be set using WithSigner option)
//   - Failed to get auth message (see GetAuthMessage for details)
//   - Failed to sign message (see SignMessage for details)
//   - Failed to get JWT token (see GetJWTTokenFromSignature for details)
//
// Example:
//
//	client, err := predictclob.NewClient(
//	    predictclob.WithSigner(signer, address),
//	    predictclob.WithAPIKey(apiKey),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	token, address, err := client.Authenticate()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Authenticated as %s\n", address.Hex())
func (c *Client) Authenticate() (string, common.Address, error) {
	if c.signer == nil {
		return "", common.Address{}, fmt.Errorf("signer is not set in client, use WithSigner option when creating client")
	}

	// Step 1: Get auth message
	authMessage, err := c.GetAuthMessage()
	if err != nil {
		return "", common.Address{}, fmt.Errorf("failed to get auth message: %w", err)
	}

	// Step 2: Sign message
	signature, address, err := SignMessage(authMessage.Message, c.signer)
	if err != nil {
		return "", common.Address{}, fmt.Errorf("failed to sign message: %w", err)
	}

	// Step 3: Get JWT token
	jwtToken, err := c.GetJWTTokenFromSignature(address.Hex(), authMessage.Message, signature)
	if err != nil {
		return "", common.Address{}, fmt.Errorf("failed to get JWT token: %w", err)
	}

	// Step 4: Set JWT token in client
	c.SetJWTToken(jwtToken.Token)

	return jwtToken.Token, address, nil
}
