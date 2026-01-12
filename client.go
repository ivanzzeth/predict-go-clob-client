package predictclob

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
)

// Client is the main client for interacting with Predict.fun API
type Client struct {
	host      string
	apiKey    string
	jwtToken  string
	reqClient *req.Client
}

// ClientConfig represents configuration for creating a client
type ClientConfig struct {
	APIHost    string
	APIKey     string
	JWTToken   string
	UserAgent  string
	Transport  *http.Transport
	APITimeout time.Duration
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

// NewClient creates a new Predict API client instance
func NewClient(options ...ClientOption) (*Client, error) {
	defaultConfig := &ClientConfig{
		APIHost:    constants.DefaultAPIHost,
		APITimeout: 30 * time.Second,
	}

	for _, option := range options {
		option(defaultConfig)
	}

	// Remove trailing slash from API host (same as Python POC)
	apiHost := strings.TrimSuffix(defaultConfig.APIHost, "/")

	reqClient := CreateReqClientWithProxy(defaultConfig.Transport, defaultConfig.UserAgent, defaultConfig.APITimeout)

	return &Client{
		host:      apiHost,
		apiKey:    defaultConfig.APIKey,
		jwtToken:  defaultConfig.JWTToken,
		reqClient: reqClient,
	}, nil
}

// NewReadOnlyClient creates a read-only client without authentication
func NewReadOnlyClient(apiHost string) *Client {
	// Remove trailing slash from API host (same as Python POC)
	apiHost = strings.TrimSuffix(apiHost, "/")

	reqClient := CreateReqClientWithProxy(nil, "", 30*time.Second)
	return &Client{
		host:      apiHost,
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

// doRequest sends an HTTP request without authentication
func (c *Client) doRequest(method, path string, body []byte) ([]byte, error) {
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, resp.String())
	}

	return resp.Bytes(), nil
}

// GetOk checks if the API is healthy
// Note: This endpoint may not exist on all API versions
func (c *Client) GetOk() (bool, error) {
	// Try to check API health by calling a simple endpoint
	_, err := c.doRequest("GET", constants.EndpointCategories, nil)
	return err == nil, err
}
