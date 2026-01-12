package predictclob

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethsig"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
)

// GetAuthMessage gets the message to sign from Predict.fun API
func (c *Client) GetAuthMessage() (*types.AuthMessageResponse, error) {
	respBody, err := c.doRequest("GET", constants.EndpointAuthMessage, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth message: %w", err)
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Parse data as AuthMessageResponse
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var authMessage types.AuthMessageResponse
	if err := json.Unmarshal(dataBytes, &authMessage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth message: %w", err)
	}

	return &authMessage, nil
}

// SignMessage signs a message using EOA private key (EIP-191 personal sign)
// Returns the signature as a hex string with 0x prefix
func SignMessage(message string, privateKeyHex string) (string, common.Address, error) {
	// Remove 0x prefix if present
	if len(privateKeyHex) >= 2 && privateKeyHex[0:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", common.Address{}, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Get address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", common.Address{}, fmt.Errorf("failed to get public key")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Create signer
	signer := ethsig.NewEthPrivateKeySigner(privateKey)

	// Sign message using personal sign (EIP-191)
	signature, err := signer.PersonalSign(message)
	if err != nil {
		return "", common.Address{}, fmt.Errorf("failed to sign message: %w", err)
	}

	// Convert signature to hex string with 0x prefix
	signatureHex := "0x" + hex.EncodeToString(signature)

	return signatureHex, address, nil
}

// GetJWTTokenFromSignature gets JWT token from Predict.fun API using signed message
func (c *Client) GetJWTTokenFromSignature(signerAddress string, message string, signature string) (*types.JWTTokenResponse, error) {
	// Ensure address is checksum format
	address := common.HexToAddress(signerAddress)
	signerAddress = address.Hex()

	// Build request body
	authReq := types.AuthRequest{
		Signer:    signerAddress,
		Message:   message,
		Signature: signature,
	}

	bodyBytes, err := json.Marshal(authReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	respBody, err := c.doRequest("POST", constants.EndpointAuth, bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT token: %w", err)
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Parse data as JWTTokenResponse
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var jwtToken types.JWTTokenResponse
	if err := json.Unmarshal(dataBytes, &jwtToken); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT token: %w", err)
	}

	return &jwtToken, nil
}

// Authenticate performs the complete authentication flow:
// 1. Get auth message
// 2. Sign message with private key
// 3. Get JWT token
// 4. Set JWT token in client
func (c *Client) Authenticate(privateKeyHex string) (string, common.Address, error) {
	// Step 1: Get auth message
	authMessage, err := c.GetAuthMessage()
	if err != nil {
		return "", common.Address{}, fmt.Errorf("failed to get auth message: %w", err)
	}

	// Step 2: Sign message
	signature, address, err := SignMessage(authMessage.Message, privateKeyHex)
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
