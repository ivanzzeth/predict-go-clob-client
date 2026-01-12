package predictclob

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
)

// GetCategories gets all categories with optional status filter
func (c *Client) GetCategories(opts *types.GetCategoriesOptions) ([]types.Category, error) {
	path := constants.EndpointCategories
	
	params := url.Values{}
	if opts != nil && opts.Status != "" {
		params.Set("status", opts.Status.String())
	}
	
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Parse data as array of categories
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var categories []types.Category
	if err := json.Unmarshal(dataBytes, &categories); err != nil {
		return nil, fmt.Errorf("failed to unmarshal categories: %w", err)
	}

	return categories, nil
}

// GetCategory gets category information by slug
func (c *Client) GetCategory(slug string) (*types.Category, error) {
	path := fmt.Sprintf(constants.EndpointCategoryBySlug, url.QueryEscape(slug))

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Parse data as category
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var category types.Category
	if err := json.Unmarshal(dataBytes, &category); err != nil {
		return nil, fmt.Errorf("failed to unmarshal category: %w", err)
	}

	return &category, nil
}

// GetMarkets gets markets with optional filters
// If status is "OPEN", it fetches markets from OPEN categories (same as POC)
func (c *Client) GetMarkets(opts *types.GetMarketsOptions) ([]types.Market, error) {
	// If status is OPEN, fetch from OPEN categories instead (same as POC)
	if opts != nil && opts.Status == types.MarketStatusOpen {
		return c.getMarketsFromOpenCategories(opts.Limit)
	}

	path := constants.EndpointMarkets

	params := url.Values{}
	if opts != nil {
		if opts.CategoryID != "" {
			params.Set("categoryId", opts.CategoryID.String())
		}
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
		// Note: Don't pass status to API if it's not OPEN
		// The API doesn't support status filter directly
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Parse data - could be array or object with markets field
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var markets []types.Market
	
	// Try to parse as array first
	if err := json.Unmarshal(dataBytes, &markets); err == nil {
		// Filter by status if provided (for non-OPEN status)
		if opts != nil && opts.Status != "" && opts.Status != types.MarketStatusOpen {
			filtered := make([]types.Market, 0)
			for _, m := range markets {
				if m.Status == opts.Status {
					filtered = append(filtered, m)
				}
			}
			markets = filtered
		}
		return markets, nil
	}

	// Try to parse as object with markets field
	var marketResponse struct {
		Markets []types.Market `json:"markets"`
	}
	if err := json.Unmarshal(dataBytes, &marketResponse); err == nil {
		// Filter by status if provided (for non-OPEN status)
		if opts != nil && opts.Status != "" && opts.Status != types.MarketStatusOpen {
			filtered := make([]types.Market, 0)
			for _, m := range marketResponse.Markets {
				if m.Status == opts.Status {
					filtered = append(filtered, m)
				}
			}
			marketResponse.Markets = filtered
		}
		return marketResponse.Markets, nil
	}

	return nil, fmt.Errorf("failed to parse markets from response")
}

// getMarketsFromOpenCategories fetches markets from OPEN categories (same as POC)
func (c *Client) getMarketsFromOpenCategories(limit int) ([]types.Market, error) {
	// Get OPEN categories
	categories, err := c.GetCategories(&types.GetCategoriesOptions{Status: types.CategoryStatusOpen})
	if err != nil {
		return nil, fmt.Errorf("failed to get OPEN categories: %w", err)
	}

	// Extract all markets from OPEN categories
	allMarkets := make([]types.Market, 0)
	seenIDs := make(map[string]bool)
	
	for _, category := range categories {
		for _, market := range category.Markets {
			if market.ID != "" && !seenIDs[string(market.ID)] {
				seenIDs[string(market.ID)] = true
				allMarkets = append(allMarkets, market)
			}
		}
	}

	// Apply limit if specified
	if limit > 0 && limit < len(allMarkets) {
		allMarkets = allMarkets[:limit]
	}

	return allMarkets, nil
}

// GetMarket gets a specific market by ID
func (c *Client) GetMarket(marketID types.MarketID) (*types.Market, error) {
	path := fmt.Sprintf(constants.EndpointMarketByID, url.QueryEscape(marketID.String()))

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Parse data as market
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var market types.Market
	if err := json.Unmarshal(dataBytes, &market); err != nil {
		return nil, fmt.Errorf("failed to unmarshal market: %w", err)
	}

	return &market, nil
}

// GetMarketStats gets market statistics
func (c *Client) GetMarketStats(marketID types.MarketID) (*types.MarketStats, error) {
	path := fmt.Sprintf(constants.EndpointMarketStats, url.QueryEscape(marketID.String()))

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Parse data as market stats
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var stats types.MarketStats
	if err := json.Unmarshal(dataBytes, &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal market stats: %w", err)
	}

	return &stats, nil
}

// GetMarketOrderbook gets the orderbook for a market
func (c *Client) GetMarketOrderbook(marketID types.MarketID) (*types.Orderbook, error) {
	path := fmt.Sprintf(constants.EndpointMarketOrderbook, url.QueryEscape(marketID.String()))

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Parse data as orderbook
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var orderbook types.Orderbook
	if err := json.Unmarshal(dataBytes, &orderbook); err != nil {
		return nil, fmt.Errorf("failed to unmarshal orderbook: %w", err)
	}

	return &orderbook, nil
}

// GetMarketLastSale gets the last sale information for a market
func (c *Client) GetMarketLastSale(marketID types.MarketID) (*types.Sale, error) {
	path := fmt.Sprintf(constants.EndpointMarketSale, url.QueryEscape(marketID.String()))

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Parse data as sale
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var sale types.Sale
	if err := json.Unmarshal(dataBytes, &sale); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sale: %w", err)
	}

	return &sale, nil
}
