package predictclob

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
)

// GetAccount gets the connected account information
// Requires JWT token authentication
func (c *Client) GetAccount() (*types.Account, error) {
	if c.jwtToken == "" {
		return nil, fmt.Errorf("JWT token is required for account operations")
	}

	respBody, err := c.doRequest("GET", constants.EndpointAccount, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Parse data as Account
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var account types.Account
	if err := json.Unmarshal(dataBytes, &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}

	return &account, nil
}

// GetPositions gets positions for the authenticated user
// Requires JWT token authentication
func (c *Client) GetPositions(opts *types.GetPositionsOptions) ([]types.Position, error) {
	if c.jwtToken == "" {
		return nil, fmt.Errorf("JWT token is required for position operations")
	}

	path := constants.EndpointPositions

	params := url.Values{}
	if opts != nil {
		if opts.MarketID != "" {
			params.Set("marketId", opts.MarketID.String())
		}
		if opts.First > 0 {
			params.Set("first", fmt.Sprintf("%d", opts.First))
		}
		if opts.After != "" {
			params.Set("after", opts.After)
		}
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Parse data as array of positions
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var positions []types.Position
	if err := json.Unmarshal(dataBytes, &positions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal positions: %w", err)
	}

	return positions, nil
}

// GetActivity gets account activity including orders, matches, conversions, merges, and splits
// Requires JWT token authentication
func (c *Client) GetActivity(opts *types.GetActivityOptions) (*types.ActivityResponse, error) {
	if c.jwtToken == "" {
		return nil, fmt.Errorf("JWT token is required for getting activity")
	}

	path := constants.EndpointAccountActivity
	params := url.Values{}

	if opts != nil {
		if opts.First != nil {
			params.Set("first", fmt.Sprintf("%d", *opts.First))
		}
		if opts.After != "" {
			params.Set("after", opts.After)
		}
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	var response types.ActivityResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return &response, nil
}
