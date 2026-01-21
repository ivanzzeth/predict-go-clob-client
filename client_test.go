package predictclob

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/ethsig/eip712"
	"github.com/ivanzzeth/predict-go-clob-client/errs"
	"github.com/ivanzzeth/predict-go-clob-client/types"
)

// mockSigner implements Signer interface for testing
type mockSigner struct {
	address   common.Address
	signature []byte
}

func (m *mockSigner) GetAddress() common.Address {
	return m.address
}

func (m *mockSigner) PersonalSign(message string) ([]byte, error) {
	return m.signature, nil
}

func (m *mockSigner) SignTypedData(typedData eip712.TypedData) ([]byte, error) {
	return m.signature, nil
}

func TestDoRequest_JWTRefreshOn401(t *testing.T) {
	var requestCount atomic.Int32
	var authMessageCount atomic.Int32
	var authCount atomic.Int32

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/message":
			authMessageCount.Add(1)
			resp := types.APIBaseResponse[types.AuthMessageResponse]{
				Success: true,
				Data: types.AuthMessageResponse{
					Message: "test message",
				},
			}
			json.NewEncoder(w).Encode(resp)

		case "/v1/auth":
			authCount.Add(1)
			resp := types.APIBaseResponse[types.JWTTokenResponse]{
				Success: true,
				Data: types.JWTTokenResponse{
					Token: "new-jwt-token",
				},
			}
			json.NewEncoder(w).Encode(resp)

		case "/v1/test":
			count := requestCount.Add(1)
			if count == 1 {
				// First request: return 401 Invalid JWT
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"success":false,"message":"Invalid JWT","trace":"test"}`))
				return
			}
			// Second request: success
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true,"data":{}}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create mock signer with valid 65-byte signature
	sig := make([]byte, 65)
	for i := range sig {
		sig[i] = byte(i)
	}
	mockSign := &mockSigner{
		address:   common.HexToAddress("0x1234567890123456789012345678901234567890"),
		signature: sig,
	}

	// Create client with mock signer
	client := &Client{
		host:      server.URL,
		apiKey:    "test-api-key",
		jwtToken:  "old-jwt-token",
		reqClient: CreateReqClientWithProxy(nil, "", 30*time.Second),
		signer:    mockSign,
	}

	// Make request - should get 401, refresh, and retry
	_, err := client.doRequest("GET", "/v1/test", nil, true)
	if err != nil {
		t.Fatalf("Expected success after JWT refresh, got error: %v", err)
	}

	// Verify request was made twice (initial + retry)
	if requestCount.Load() != 2 {
		t.Errorf("Expected 2 requests to /v1/test, got %d", requestCount.Load())
	}

	// Verify authentication was called once
	if authMessageCount.Load() != 1 {
		t.Errorf("Expected 1 auth message request, got %d", authMessageCount.Load())
	}
	if authCount.Load() != 1 {
		t.Errorf("Expected 1 auth request, got %d", authCount.Load())
	}

	// Verify token was updated
	if client.jwtToken != "new-jwt-token" {
		t.Errorf("Expected token to be 'new-jwt-token', got '%s'", client.jwtToken)
	}
}

func TestDoRequest_NoRefreshWithoutSigner(t *testing.T) {
	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success":false,"message":"Invalid JWT","trace":"test"}`))
	}))
	defer server.Close()

	// Create client WITHOUT signer
	client := &Client{
		host:      server.URL,
		apiKey:    "test-api-key",
		jwtToken:  "old-jwt-token",
		reqClient: CreateReqClientWithProxy(nil, "", 30*time.Second),
		signer:    nil, // No signer
	}

	// Make request - should fail without retry
	_, err := client.doRequest("GET", "/v1/test", nil, true)
	if err == nil {
		t.Fatal("Expected error when JWT expired without signer")
	}

	// Verify only one request was made (no retry)
	if requestCount.Load() != 1 {
		t.Errorf("Expected 1 request (no retry without signer), got %d", requestCount.Load())
	}

	// Verify error contains ErrJWTTokenExpired
	if !containsError(err, errs.ErrJWTTokenExpired) {
		t.Errorf("Expected ErrJWTTokenExpired in error, got: %v", err)
	}
}

func TestDoRequest_NoInfiniteRetryLoop(t *testing.T) {
	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/message":
			resp := types.APIBaseResponse[types.AuthMessageResponse]{
				Success: true,
				Data:    types.AuthMessageResponse{Message: "test"},
			}
			json.NewEncoder(w).Encode(resp)

		case "/v1/auth":
			resp := types.APIBaseResponse[types.JWTTokenResponse]{
				Success: true,
				Data:    types.JWTTokenResponse{Token: "still-invalid-token"},
			}
			json.NewEncoder(w).Encode(resp)

		case "/v1/test":
			requestCount.Add(1)
			// Always return 401 - simulating consistently invalid token
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"success":false,"message":"Invalid JWT","trace":"test"}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	sig := make([]byte, 65)
	mockSign := &mockSigner{
		address:   common.HexToAddress("0x1234567890123456789012345678901234567890"),
		signature: sig,
	}

	client := &Client{
		host:      server.URL,
		apiKey:    "test-api-key",
		jwtToken:  "old-jwt-token",
		reqClient: CreateReqClientWithProxy(nil, "", 30*time.Second),
		signer:    mockSign,
	}

	// Make request - should try once, refresh, retry once, then fail
	_, err := client.doRequest("GET", "/v1/test", nil, true)
	if err == nil {
		t.Fatal("Expected error after retry with still-invalid token")
	}

	// Verify only 2 requests were made (initial + one retry, no infinite loop)
	if requestCount.Load() != 2 {
		t.Errorf("Expected exactly 2 requests (no infinite loop), got %d", requestCount.Load())
	}
}

func TestRefreshJWTToken_RequiresSigner(t *testing.T) {
	client := &Client{
		signer: nil,
	}

	err := client.refreshJWTToken()
	if err == nil {
		t.Fatal("Expected error when refreshing without signer")
	}

	if !containsError(err, errs.ErrSignerRequiredForRefresh) {
		t.Errorf("Expected ErrSignerRequiredForRefresh, got: %v", err)
	}
}

// containsError checks if err wraps or equals target
func containsError(err, target error) bool {
	if err == nil {
		return target == nil
	}
	return err.Error() == target.Error() ||
		(len(err.Error()) > len(target.Error()) &&
		 err.Error()[:len(target.Error())] == target.Error()) ||
		containsErrorInChain(err, target)
}

func containsErrorInChain(err, target error) bool {
	for err != nil {
		if err.Error() == target.Error() {
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = unwrapper.Unwrap()
	}
	return false
}
