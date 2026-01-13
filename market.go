package predictclob

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
)

// GetCategories gets all categories with optional filters and pagination
func (c *Client) GetCategories(opts *types.GetCategoriesOptions) (*types.GetCategoriesResponse, error) {
	path := constants.EndpointCategories

	params := url.Values{}
	if opts != nil {
		if opts.First != nil {
			params.Set("first", fmt.Sprintf("%d", *opts.First))
		}
		if opts.After != nil && *opts.After != "" {
			params.Set("after", *opts.After)
		}
		if opts.Status != "" {
			params.Set("status", opts.Status.String())
		}
		if opts.Sort != "" {
			params.Set("sort", opts.Sort.String())
		}
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, err
	}

	var response types.GetCategoriesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse categories response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return &response, nil
}

// GetCategory gets category information by slug
func (c *Client) GetCategory(slug string) (*types.Category, error) {
	path := fmt.Sprintf(constants.EndpointCategoryBySlug, url.QueryEscape(slug))

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, err
	}

	var response types.APIBaseResponse[types.Category]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse category response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	return &response.Data, nil
}

// GetMarkets gets markets with optional pagination
func (c *Client) GetMarkets(opts *types.GetMarketsOptions) (*types.GetMarketsResponse, error) {
	path := constants.EndpointMarkets

	params := url.Values{}
	if opts != nil {
		if opts.First != nil && *opts.First != "" {
			params.Set("first", *opts.First)
		}
		if opts.After != nil && *opts.After != "" {
			params.Set("after", *opts.After)
		}
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, err
	}

	var response types.GetMarketsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse markets response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return &response, nil
}

// GetMarket gets a specific market by ID
// useCache: if true and cacheTTL > 0, returns cached market if available and not expired
// For split/merge/redeem operations, useCache=true is recommended as they only need conditionID and isYieldBearing
func (c *Client) GetMarket(marketID types.MarketID, useCache bool) (*types.Market, error) {
	marketIDStr := marketID.String()

	// Try cache if enabled
	if useCache && c.cacheTTL > 0 {
		c.cacheMu.RLock()
		cached, exists := c.marketCache[marketIDStr]
		c.cacheMu.RUnlock()

		if exists && cached != nil {
			// Check if cache is still valid
			if time.Now().Before(cached.expiresAt) {
				return cached.market, nil
			}
			// Cache expired, remove it
			c.cacheMu.Lock()
			delete(c.marketCache, marketIDStr)
			c.cacheMu.Unlock()
		}
	}

	// Fetch from API
	path := fmt.Sprintf(constants.EndpointMarketByID, url.QueryEscape(marketIDStr))

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, err
	}

	var response types.APIBaseResponse[types.Market]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse market response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	market := &response.Data

	// Update cache if enabled
	if useCache && c.cacheTTL > 0 {
		c.cacheMu.Lock()
		c.marketCache[marketIDStr] = &cachedMarket{
			market:    market,
			expiresAt: time.Now().Add(c.cacheTTL),
		}
		c.cacheMu.Unlock()
	}

	return market, nil
}

// GetMarketStats gets market statistics
func (c *Client) GetMarketStats(marketID types.MarketID) (*types.MarketStats, error) {
	path := fmt.Sprintf(constants.EndpointMarketStats, url.QueryEscape(marketID.String()))

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, err
	}

	var response types.APIBaseResponse[types.MarketStats]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse market stats response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	return &response.Data, nil
}

// GetMarketOrderbook gets the orderbook for a market
func (c *Client) GetMarketOrderbook(marketID types.MarketID) (*types.Orderbook, error) {
	path := fmt.Sprintf(constants.EndpointMarketOrderbook, url.QueryEscape(marketID.String()))

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, err
	}

	var response types.APIBaseResponse[types.Orderbook]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse orderbook response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	return &response.Data, nil
}

// GetMarketLastSale gets the last sale information for a market
// Returns nil if no sale data is available (data field is null)
func (c *Client) GetMarketLastSale(marketID types.MarketID) (*types.MarketLastSale, error) {
	path := fmt.Sprintf(constants.EndpointMarketLastSale, url.QueryEscape(marketID.String()))

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, err
	}

	// Handle nullable data field
	var response struct {
		Success bool                  `json:"success"`
		Data    *types.MarketLastSale `json:"data"` // nullable
		Message string                `json:"message,omitempty"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse last sale response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Return nil if data is null (no sale available)
	return response.Data, nil
}
