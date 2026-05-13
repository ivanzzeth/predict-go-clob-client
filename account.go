package predictclob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
	predictcontracts "github.com/ivanzzeth/predict-go-contracts"
	"github.com/shopspring/decimal"
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

// SetReferral sets a referral code for the authenticated user.
// The referral code must be exactly 5 characters long.
// Returns error if the user already has a referral code set.
// Requires JWT token authentication.
func (c *Client) SetReferral(referralCode string) error {
	if err := c.requireJWTToken(); err != nil {
		return err
	}

	if len(referralCode) != 5 {
		return fmt.Errorf("referral code must be exactly 5 characters, got %d", len(referralCode))
	}

	reqBody := types.SetReferralRequest{
		Data: types.SetReferralData{
			ReferralCode: referralCode,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal referral request: %w", err)
	}

	respBody, err := c.doRequest("POST", constants.EndpointAccountReferral, bodyBytes, true)
	if err != nil {
		return fmt.Errorf("failed to set referral: %w", err)
	}

	var response types.SetReferralResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return fmt.Errorf("failed to parse referral response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("API returned success=false for set referral")
	}

	return nil
}

// GetPositions gets ALL positions for the authenticated user (auto-paginates)
// Requires JWT token authentication
// Calculates Locked and Available amounts based on OPEN SELL orders
func (c *Client) GetPositions(opts *types.GetPositionsOptions) ([]types.Position, error) {
	if err := c.requireJWTToken(); err != nil {
		return nil, err
	}

	// Determine page size: default to 20 if not specified
	first := 20
	var after string
	var marketID types.MarketID
	if opts != nil {
		marketID = opts.MarketID
		if opts.First > 0 {
			first = opts.First
		}
		after = opts.After
	}

	// Auto-paginate: collect all positions across pages
	var allPositions []types.Position

	for {
		path := constants.EndpointPositions
		params := url.Values{}
		params.Set("first", fmt.Sprintf("%d", first))
		if !marketID.IsZero() {
			params.Set("marketId", marketID.String())
		}
		if after != "" {
			params.Set("after", after)
		}
		path += "?" + params.Encode()

		respBody, err := c.doRequest("GET", path, nil, true)
		if err != nil {
			return nil, fmt.Errorf("failed to get positions: %w", err)
		}

		var response types.GetPositionsResponse
		if err := json.Unmarshal(respBody, &response); err != nil {
			return nil, fmt.Errorf("failed to parse positions response: %w", err)
		}

		if !response.Success {
			return nil, fmt.Errorf("API returned success=false")
		}

		allPositions = append(allPositions, response.Data...)

		// Check if there are more pages
		if response.Cursor == nil || *response.Cursor == "" {
			break
		}
		after = *response.Cursor
	}

	// Get all OPEN orders to calculate locked amounts
	openOrdersOpts := &types.GetOrdersOptions{
		Status: types.OrderStatusOpen,
	}
	openOrdersResp, err := c.GetOrders(openOrdersOpts)
	if err != nil {
		// If we can't get orders, return positions with Locked=0 and Available=Total
		for i := range allPositions {
			allPositions[i].Total = allPositions[i].Amount
			allPositions[i].Locked = decimal.Zero
			allPositions[i].Available = allPositions[i].Amount
		}
		return allPositions, nil
	}

	// Build a map of locked amounts by (marketID, tokenID)
	// For SELL orders, the locked amount is the makerAmount (shares being sold)
	lockedMap := make(map[string]decimal.Decimal) // key: "marketID:tokenID"
	for _, order := range openOrdersResp.Data {
		if order.Status == types.OrderStatusOpen && order.OrderData.Side == types.OrderSideSell {
			key := fmt.Sprintf("%s:%s", order.MarketID.String(), string(order.OrderData.TokenID))
			lockedMap[key] = lockedMap[key].Add(order.OrderData.MakerAmount)
		}
	}

	// Calculate Locked and Available for each position
	for i := range allPositions {
		allPositions[i].Total = allPositions[i].Amount

		// Find locked amount for this position's market and outcome
		key := fmt.Sprintf("%s:%s", allPositions[i].Market.ID.String(), string(allPositions[i].Outcome.OnChainID))
		locked := lockedMap[key]

		allPositions[i].Locked = locked
		allPositions[i].Available = allPositions[i].Total.Sub(locked)

		// Ensure Available is not negative
		if allPositions[i].Available.IsNegative() {
			allPositions[i].Available = decimal.Zero
		}
	}

	return allPositions, nil
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

// GetBalance gets collateral balance for the authenticated user
// Calculates Locked and Available amounts based on OPEN BUY orders
// Requires contractInterface to be initialized (via WithEOATradingSigner or WithRPCURL)
func (c *Client) GetBalance(ctx context.Context) (*types.Balance, error) {
	if c.contractInterface == nil {
		return nil, fmt.Errorf("contractInterface is required for getting balance. Use WithEOATradingSigner or WithRPCURL option when creating client")
	}

	// Get signer address
	var address common.Address
	if c.eoaTradingSigner != nil {
		address = c.eoaTradingSigner.GetAddress()
	} else if c.signer != nil {
		address = c.signer.GetAddress()
	} else {
		return nil, fmt.Errorf("no signer available to get address")
	}

	// Get total balance from blockchain
	balanceInfo, err := c.contractInterface.CheckBalanceAndAllowance(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get collateral balance: %w", err)
	}

	// Convert balance from wei to decimal (collateral has 18 decimals)
	totalBalance := decimal.NewFromBigInt(balanceInfo.Balance, 0).Shift(-predictcontracts.COLLATERAL_TOKEN_DECIMALS)

	// Get all OPEN orders to calculate locked amounts
	openOrdersOpts := &types.GetOrdersOptions{
		Status: types.OrderStatusOpen,
	}
	openOrdersResp, err := c.GetOrders(openOrdersOpts)
	if err != nil {
		// If we can't get orders, return balance with Locked=0 and Available=Total
		return &types.Balance{
			Total:     totalBalance,
			Locked:    decimal.Zero,
			Available: totalBalance,
		}, nil
	}

	// Calculate locked collateral from OPEN BUY orders
	// For BUY orders, makerAmount is the collateral being paid
	// Locked amount = makerAmount * (unfilled ratio) = makerAmount * (Amount - AmountFilled) / Amount
	lockedCollateral := decimal.Zero
	for _, order := range openOrdersResp.Data {
		if order.Status == types.OrderStatusOpen && order.OrderData.Side == types.OrderSideBuy {
			// Calculate unfilled ratio
			if order.Amount.GreaterThan(decimal.Zero) {
				unfilledRatio := order.Amount.Sub(order.AmountFilled).Div(order.Amount)
				// Locked collateral = makerAmount * unfilledRatio
				// Note: For BUY orders, the full makerAmount is locked, not just the unfilled portion
				// The unfilled ratio applies to the shares, but the collateral is locked upfront
				lockedCollateral = lockedCollateral.Add(order.OrderData.MakerAmount.Mul(unfilledRatio))
			} else {
				// If Amount is 0 or negative, use full makerAmount as locked
				lockedCollateral = lockedCollateral.Add(order.OrderData.MakerAmount)
			}
		}
	}

	// Calculate available balance
	availableBalance := totalBalance.Sub(lockedCollateral)
	if availableBalance.IsNegative() {
		availableBalance = decimal.Zero
	}

	return &types.Balance{
		Total:     totalBalance,
		Locked:    lockedCollateral,
		Available: availableBalance,
	}, nil
}
