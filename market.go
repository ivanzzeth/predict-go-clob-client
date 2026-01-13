package predictclob

import (
	"encoding/json"
	"fmt"
	"net/url"

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

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, err
	}

	// Handle two possible response formats: array or object with markets field
	var responseArray types.APIBaseResponse[[]types.Market]
	if err := json.Unmarshal(respBody, &responseArray); err == nil && responseArray.Success {
		markets := responseArray.Data
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
	var marketResponseObj struct {
		Success bool `json:"success"`
		Data    struct {
			Markets []types.Market `json:"markets"`
		} `json:"data"`
		Message string `json:"message,omitempty"`
	}
	if err := json.Unmarshal(respBody, &marketResponseObj); err == nil && marketResponseObj.Success {
		markets := marketResponseObj.Data.Markets
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

	return nil, fmt.Errorf("failed to parse markets from response")
}

// getMarketsFromOpenCategories fetches markets from OPEN categories (same as POC)
func (c *Client) getMarketsFromOpenCategories(limit int) ([]types.Market, error) {
	// Get OPEN categories
	response, err := c.GetCategories(&types.GetCategoriesOptions{Status: types.CategoryStatusOpen})
	if err != nil {
		return nil, fmt.Errorf("failed to get OPEN categories: %w", err)
	}

	// Extract all markets from OPEN categories
	allMarkets := make([]types.Market, 0)
	seenIDs := make(map[string]bool)

	for _, category := range response.Data {
		for _, categoryMarket := range category.Markets {
			marketID := categoryMarket.ID.String()
			if marketID != "" && !seenIDs[marketID] {
				seenIDs[marketID] = true
				// Convert CategoryMarket to Market
				market := types.Market{
					ID:             categoryMarket.ID,
					Title:          categoryMarket.Title,
					Question:       categoryMarket.Question,
					Description:    categoryMarket.Description,
					Status:         categoryMarket.Status,
					FeeRateBps:     fmt.Sprintf("%d", categoryMarket.FeeRateBps),
					IsNegRisk:      categoryMarket.IsNegRisk,
					IsYieldBearing: categoryMarket.IsYieldBearing,
					CreatedAt:      categoryMarket.CreatedAt,
				}
				// Convert outcomes
				for _, outcome := range categoryMarket.Outcomes {
					market.Outcomes = append(market.Outcomes, types.Outcome{
						Name:      outcome.Name,
						IndexSet:  outcome.IndexSet,
						OnChainID: outcome.OnChainID,
					})
				}
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

	return &response.Data, nil
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
func (c *Client) GetMarketLastSale(marketID types.MarketID) (*types.Sale, error) {
	path := fmt.Sprintf(constants.EndpointMarketSale, url.QueryEscape(marketID.String()))

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, err
	}

	var response types.APIBaseResponse[types.Sale]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse sale response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	return &response.Data, nil
}
