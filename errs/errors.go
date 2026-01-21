package errs

import "fmt"

// SDK errors
var (
	// ErrAPIKeyRequired is returned when an API key is required but not set
	ErrAPIKeyRequired = fmt.Errorf("API key is required for this endpoint. Use WithAPIKey option when creating client or SetAPIKey method")

	// ErrJWTTokenRequired is returned when a JWT token is required but not set
	ErrJWTTokenRequired = fmt.Errorf("JWT token is required for this operation. Use Authenticate method or SetJWTToken method")

	// ErrJWTTokenExpired is returned when the JWT token has expired (401 Invalid JWT)
	ErrJWTTokenExpired = fmt.Errorf("JWT token has expired")

	// ErrJWTRefreshFailed is returned when auto-refresh of JWT token fails
	ErrJWTRefreshFailed = fmt.Errorf("failed to refresh JWT token")

	// ErrSignerRequiredForRefresh is returned when JWT refresh is needed but no signer is available
	ErrSignerRequiredForRefresh = fmt.Errorf("signer is required for JWT token refresh")

	// ErrInvalidAddress is returned when an address format is invalid
	ErrInvalidAddress = fmt.Errorf("invalid address format")

	// ErrInvalidSignature is returned when a signature format is invalid
	ErrInvalidSignature = fmt.Errorf("invalid signature format")

	// ErrEmptyMessage is returned when a message is empty
	ErrEmptyMessage = fmt.Errorf("message cannot be empty")

	// ErrEmptyPrivateKey is returned when a private key is empty
	ErrEmptyPrivateKey = fmt.Errorf("private key cannot be empty")
)

// NewInvalidAddressError creates a new error for invalid address format
func NewInvalidAddressError(address string) error {
	return fmt.Errorf("%w: %s", ErrInvalidAddress, address)
}

// NewInvalidSignatureError creates a new error for invalid signature format
func NewInvalidSignatureError(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidSignature, reason)
}

// NewInvalidPrivateKeyLengthError creates a new error for invalid private key length
func NewInvalidPrivateKeyLengthError(got, expected int) error {
	return fmt.Errorf("invalid private key length: expected %d hex characters (%d bytes), got %d", expected, expected/2, got)
}

// NewInvalidSignatureLengthError creates a new error for invalid signature length
func NewInvalidSignatureLengthError(got, expected int) error {
	return fmt.Errorf("invalid signature length: expected %d characters (0x + %d hex), got %d", expected, expected-2, got)
}
