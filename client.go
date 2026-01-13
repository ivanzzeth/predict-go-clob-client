package predictclob

import (
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/imroc/req/v3"
	"github.com/ivanzzeth/ethclient"
	"github.com/ivanzzeth/ethsig"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/errs"
	"github.com/ivanzzeth/predict-go-clob-client/types"
	predictcontracts "github.com/ivanzzeth/predict-go-contracts"
	"github.com/ivanzzeth/predict-go-contracts/signer"
	"github.com/ivanzzeth/predict-go-order-utils/pkg/builder"
)

// Signer interface for order signing operations
// Extends builder.Signer with PersonalSigner interface for authentication
type Signer interface {
	builder.Signer
	ethsig.PersonalSigner
}

// EOATradingSigner interface for chain operations (re-exported from signer package)
type EOATradingSigner = signer.EOATradingSigner

// cachedMarket represents a cached market with expiration time
type cachedMarket struct {
	market    *types.Market
	expiresAt time.Time
}

// Client is the main client for interacting with Predict.fun API
type Client struct {
	host      string
	apiKey    string
	jwtToken  string
	reqClient *req.Client

	// Order signing
	chainID *big.Int
	signer  Signer         // For order signing (CLOB API)
	funder  common.Address // Maker address (usually same as signer for EOA)

	// Chain operations
	rpcURL            string
	eoaTradingSigner  EOATradingSigner // For chain operations (enable trading, split, merge, redeem)
	ethClient         *ethclient.Client
	contractInterface *predictcontracts.ContractInterface

	// Cache
	marketCache map[string]*cachedMarket // key: marketID string
	cacheTTL    time.Duration            // Cache TTL, 0 means no caching
	cacheMu     sync.RWMutex             // Mutex for cache access
}

// ClientConfig represents configuration for creating a client
type ClientConfig struct {
	APIHost    string
	APIKey     string
	JWTToken   string
	UserAgent  string
	Transport  *http.Transport
	APITimeout time.Duration

	// Order signing config
	ChainID *big.Int
	Signer  Signer         // For order signing (CLOB API)
	Funder  common.Address // Maker address

	// Chain operations config
	RPCURL           string
	EOATradingSigner EOATradingSigner // For chain operations

	// Cache config
	CacheTTL time.Duration // Cache TTL for market data, 0 means no caching (default: 0)
}

// ClientOption is a function that configures a ClientConfig
type ClientOption func(*ClientConfig)

// WithAPIHost sets the API host URL
func WithAPIHost(host string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.APIHost = host
	}
}

// WithAPIKey sets the API key
func WithAPIKey(apiKey string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.APIKey = apiKey
	}
}

// WithJWTToken sets the JWT token
func WithJWTToken(token string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.JWTToken = token
	}
}

// WithHttpClientOptions sets HTTP client options
func WithHttpClientOptions(transport *http.Transport, apiTimeout time.Duration) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.Transport = transport
		cfg.APITimeout = apiTimeout
	}
}

// WithUserAgent sets the user agent
func WithUserAgent(userAgent string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.UserAgent = userAgent
	}
}

// WithChainID sets the chain ID for chain operations
func WithChainID(chainID *big.Int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.ChainID = chainID
	}
}

// WithRPCURL sets the RPC URL for chain operations
func WithRPCURL(rpcURL string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.RPCURL = rpcURL
	}
}

// WithSigner sets the signer and funder address for order signing (CLOB API)
// signer: implements builder.Signer interface for signing orders
// funder: the maker address that will be used in orders (typically same as signer.GetAddress() for EOA)
func WithSigner(signer Signer, funder common.Address) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.Signer = signer
		cfg.Funder = funder
	}
}

// WithEOATradingSigner sets the EOA trading signer for chain operations (enable trading, split, merge, redeem)
// signer: implements EOATradingSigner interface for signing transactions
// funder: the maker address (typically same as signer.GetAddress() for EOA)
func WithEOATradingSigner(signer EOATradingSigner) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.Signer = signer
		cfg.EOATradingSigner = signer
		cfg.Funder = signer.GetAddress()
	}
}

// WithCacheTTL sets the cache TTL for market data
// ttl: time duration for cache expiration, 0 means no caching
// For split/merge/redeem operations, caching is useful as they only need conditionID and isYieldBearing
func WithCacheTTL(ttl time.Duration) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.CacheTTL = ttl
	}
}

// NewClient creates a new Predict API client instance
func NewClient(options ...ClientOption) (*Client, error) {
	defaultConfig := &ClientConfig{
		APIHost:    constants.DefaultAPIHost,
		APITimeout: 30 * time.Second,
		ChainID:    big.NewInt(56), // Default to BNB Chain mainnet
	}

	for _, option := range options {
		option(defaultConfig)
	}

	// Remove trailing slash from API host (same as Python POC)
	apiHost := strings.TrimSuffix(defaultConfig.APIHost, "/")

	reqClient := CreateReqClientWithProxy(defaultConfig.Transport, defaultConfig.UserAgent, defaultConfig.APITimeout)

	// If funder is not set but signer is, use signer's address as funder
	funder := defaultConfig.Funder
	if funder == (common.Address{}) && defaultConfig.Signer != nil {
		funder = defaultConfig.Signer.GetAddress()
	}
	// If funder is still not set but EOATradingSigner is, use its address
	if funder == (common.Address{}) && defaultConfig.EOATradingSigner != nil {
		funder = defaultConfig.EOATradingSigner.GetAddress()
	}

	client := &Client{
		host:             apiHost,
		apiKey:           defaultConfig.APIKey,
		jwtToken:         defaultConfig.JWTToken,
		reqClient:        reqClient,
		chainID:          defaultConfig.ChainID,
		signer:           defaultConfig.Signer,
		funder:           funder,
		rpcURL:           defaultConfig.RPCURL,
		eoaTradingSigner: defaultConfig.EOATradingSigner,
		marketCache:      make(map[string]*cachedMarket),
		cacheTTL:         defaultConfig.CacheTTL,
	}

	// Initialize contract interface if RPC URL and EOATradingSigner are provided
	if defaultConfig.RPCURL != "" && defaultConfig.EOATradingSigner != nil {
		if err := client.initContractInterface(); err != nil {
			return nil, fmt.Errorf("failed to initialize contract interface: %w", err)
		}
	}

	// Auto-authenticate if Signer and APIKey are provided, but JWTToken is not set
	if defaultConfig.Signer != nil && defaultConfig.APIKey != "" && defaultConfig.JWTToken == "" {
		_, _, err := client.Authenticate()
		if err != nil {
			return nil, fmt.Errorf("failed to auto-authenticate during client initialization: %w", err)
		}
	}

	return client, nil
}

// initContractInterface initializes the contract interface for chain operations
func (c *Client) initContractInterface() error {
	if c.rpcURL == "" {
		return fmt.Errorf("RPC URL is required for chain operations")
	}
	if c.eoaTradingSigner == nil {
		return fmt.Errorf("EOATradingSigner is required for chain operations")
	}

	// Create eth client
	ethClient, err := ethclient.Dial(c.rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC: %w", err)
	}
	c.ethClient = ethClient

	// Get contract config
	config := predictcontracts.GetContractConfig(c.chainID)

	// Create contract interface with EOA signer
	contractInterface, err := predictcontracts.NewContractInterface(
		ethClient,
		predictcontracts.WithContractConfig(config),
		predictcontracts.WithEOASigner(c.eoaTradingSigner),
	)
	if err != nil {
		return fmt.Errorf("failed to create contract interface: %w", err)
	}

	c.contractInterface = contractInterface
	return nil
}

// NewReadOnlyClient creates a read-only client with API key (no authentication required)
func NewReadOnlyClient(apiHost string, apiKey string) *Client {
	// Remove trailing slash from API host (same as Python POC)
	apiHost = strings.TrimSuffix(apiHost, "/")

	reqClient := CreateReqClientWithProxy(nil, "", 30*time.Second)
	return &Client{
		host:      apiHost,
		apiKey:    apiKey,
		reqClient: reqClient,
	}
}

// SetAPIKey sets the API key for the client
func (c *Client) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

// SetJWTToken sets the JWT token for the client
func (c *Client) SetJWTToken(token string) {
	c.jwtToken = token
}

// GetAPIHost returns the API host URL
func (c *Client) GetAPIHost() string {
	return c.host
}

// GetJWTToken returns the JWT token if set
func (c *Client) GetJWTToken() string {
	return c.jwtToken
}

// GetChainID returns the chain ID
func (c *Client) GetChainID() *big.Int {
	return c.chainID
}

// GetSignerAddress returns the signer's address
func (c *Client) GetSignerAddress() common.Address {
	if c.signer == nil {
		return common.Address{}
	}
	return c.signer.GetAddress()
}

// GetFunderAddress returns the funder/maker address
func (c *Client) GetFunderAddress() common.Address {
	return c.funder
}

// GetEOATradingSigner returns the EOA trading signer
func (c *Client) GetEOATradingSigner() EOATradingSigner {
	return c.eoaTradingSigner
}

// GetContractInterface returns the contract interface
func (c *Client) GetContractInterface() *predictcontracts.ContractInterface {
	return c.contractInterface
}

// GetEthClient returns the Ethereum client
func (c *Client) GetEthClient() *ethclient.Client {
	return c.ethClient
}

// requireAPIKey validates that API key is set in the client
// Returns error if API key is missing
func (c *Client) requireAPIKey() error {
	if c.apiKey == "" {
		return errs.ErrAPIKeyRequired
	}
	return nil
}

// requireJWTToken validates that JWT token is set in the client
// Returns error if JWT token is missing
func (c *Client) requireJWTToken() error {
	if c.jwtToken == "" {
		return errs.ErrJWTTokenRequired
	}
	return nil
}

// doRequest sends an HTTP request
// requireAPIKey: if true, validates that API key is set before making the request
func (c *Client) doRequest(method, path string, body []byte, requireAPIKey bool) ([]byte, error) {
	// Validate API key if required
	if requireAPIKey {
		if err := c.requireAPIKey(); err != nil {
			return nil, err
		}
	}

	url := c.host + path

	headers := make(map[string]string)
	if c.apiKey != "" {
		headers["x-api-key"] = c.apiKey
	}
	if c.jwtToken != "" {
		headers["Authorization"] = "Bearer " + c.jwtToken
	}

	resp, err := DoReqClientRequest(c.reqClient, method, url, body, headers, true)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, resp.String())
	}

	// Print raw response for debugging during development
	respBytes := resp.Bytes()
	log.Printf("[DEBUG] %s %s - Response (status %d): %s", method, path, resp.StatusCode, string(respBytes))

	return respBytes, nil
}

// GetOk checks if the API is healthy
// Note: This endpoint may not exist on all API versions
func (c *Client) GetOk() (bool, error) {
	// Try to check API health by calling a simple endpoint
	_, err := c.doRequest("GET", constants.EndpointCategories, nil, true)
	return err == nil, err
}
