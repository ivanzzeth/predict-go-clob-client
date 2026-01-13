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
	if err := c.requireJWTToken(); err != nil {
		return nil, err
	}

	respBody, err := c.doRequest("GET", constants.EndpointAccount, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	var response types.APIBaseResponse[types.Account]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse account response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	return &response.Data, nil
}

// GetPositions gets positions for the authenticated user
// Requires JWT token authentication
func (c *Client) GetPositions(opts *types.GetPositionsOptions) ([]types.Position, error) {
	if err := c.requireJWTToken(); err != nil {
		return nil, err
	}

	path := constants.EndpointPositions

	params := url.Values{}
	if opts != nil {
		if !opts.MarketID.IsZero() {
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

	var response types.APIBaseResponse[[]types.Position]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse positions response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	return response.Data, nil
}

// GetActivity gets account activity including orders, matches, conversions, merges, and splits
// Requires JWT token authentication
func (c *Client) GetActivity(opts *types.GetActivityOptions) (*types.ActivityResponse, error) {
	if err := c.requireJWTToken(); err != nil {
		return nil, err
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
