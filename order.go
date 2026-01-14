package predictclob

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
	predictcontracts "github.com/ivanzzeth/predict-go-contracts"
	"github.com/ivanzzeth/predict-go-order-utils/pkg/builder"
	ordermodel "github.com/ivanzzeth/predict-go-order-utils/pkg/model"
	"github.com/ivanzzeth/predict-go-order-utils/pkg/utils"
	"github.com/shopspring/decimal"
)

// PlaceOrder places a limit or market order on the Predict exchange
func (c *Client) PlaceOrder(input *types.PlaceOrderInput) (*types.PlaceOrderResult, error) {
	if c.signer == nil {
		return nil, fmt.Errorf("signer is required for placing orders. Use WithSigner option when creating client")
	}
	if c.jwtToken == "" {
		return nil, fmt.Errorf("JWT token is required for placing orders. Use WithJWTToken option or authenticate first")
	}

	// Get market info to get feeRateBps and other parameters
	market, err := c.GetMarket(input.MarketID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get market: %w", err)
	}

	// Validate price for LIMIT orders
	if input.Strategy == types.OrderStrategyLimit && input.Price.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("price is required and must be positive for LIMIT orders")
	}

	// Get token ID from input or market outcomes
	// Also determine if this is outcome 0 (IndexSet=1) or outcome 1 (IndexSet=2)
	tokenID := string(input.TokenID)
	isOutcome1 := false // Track if we're trading outcome 1 (IndexSet=2)
	if tokenID == "" {
		// Use outcome based on side
		if len(market.Outcomes) == 0 {
			return nil, fmt.Errorf("market has no outcomes")
		}

		// For BUY, use first outcome (indexSet=1, usually "Yes"/"Up")
		// For SELL, use second outcome if available (indexSet=2, usually "No"/"Down")
		if input.Side == types.OrderSideBuy {
			for _, outcome := range market.Outcomes {
				if outcome.IndexSet == 1 {
					tokenID = string(outcome.OnChainID)
					isOutcome1 = false // IndexSet=1 is outcome 0
					break
				}
			}
			if tokenID == "" {
				tokenID = string(market.Outcomes[0].OnChainID)
				// Check if first outcome is IndexSet=2 (outcome 1)
				if market.Outcomes[0].IndexSet == 2 {
					isOutcome1 = true
				}
			}
		} else {
			for _, outcome := range market.Outcomes {
				if outcome.IndexSet == 2 {
					tokenID = string(outcome.OnChainID)
					isOutcome1 = true // IndexSet=2 is outcome 1
					break
				}
			}
			if tokenID == "" && len(market.Outcomes) > 1 {
				tokenID = string(market.Outcomes[1].OnChainID)
				if market.Outcomes[1].IndexSet == 2 {
					isOutcome1 = true
				}
			} else if tokenID == "" {
				tokenID = string(market.Outcomes[0].OnChainID)
				if market.Outcomes[0].IndexSet == 2 {
					isOutcome1 = true
				}
			}
		}
	} else {
		// TokenID is provided, need to find which outcome it corresponds to
		for _, outcome := range market.Outcomes {
			if string(outcome.OnChainID) == tokenID {
				isOutcome1 = (outcome.IndexSet == 2) // IndexSet=2 means outcome 1
				break
			}
		}
	}

	if tokenID == "" {
		return nil, fmt.Errorf("could not determine tokenId from market")
	}

	// Get exchange address based on market type
	exchangeAddr, err := c.getExchangeAddress(market.IsNegRisk, market.IsYieldBearing)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange address: %w", err)
	}

	// Create order builder
	orderBuilder := builder.NewExchangeOrderBuilderImpl(c.chainID, nil)

	// Calculate maker and taker amounts
	var makerAmount, takerAmount decimal.Decimal
	pricePerShare := input.Price

	// If price is zero (or MARKET order), get price from orderbook
	if pricePerShare.IsZero() || input.Strategy == types.OrderStrategyMarket {
		// Query orderbook to get best price
		orderbook, err := c.GetMarketOrderbook(input.MarketID)
		if err != nil {
			return nil, fmt.Errorf("failed to get orderbook for market order: %w", err)
		}

		// For binary markets, orderbook returns prices for outcome 0
		// If we're trading outcome 1, we need to convert prices: price_1 = 1 - price_0
		// Also swap bids and asks: bid_1 = 1 - ask_0, ask_1 = 1 - bid_0
		var bestBid, bestAsk decimal.Decimal
		if isOutcome1 {
			// Outcome 1: convert complementary prices and swap bids/asks
			if len(orderbook.Bids) == 0 || len(orderbook.Asks) == 0 {
				return nil, fmt.Errorf("orderbook missing bids or asks for outcome 1 conversion")
			}
			// For outcome 1: bid = 1 - outcome 0's ask, ask = 1 - outcome 0's bid
			one := decimal.NewFromInt(1)
			bestBid = one.Sub(orderbook.BestAsk) // Outcome 1 bid = 1 - outcome 0 ask
			bestAsk = one.Sub(orderbook.BestBid) // Outcome 1 ask = 1 - outcome 0 bid
		} else {
			// Outcome 0: use prices as-is
			if len(orderbook.Bids) == 0 || len(orderbook.Asks) == 0 {
				return nil, fmt.Errorf("orderbook missing bids or asks")
			}
			bestBid = orderbook.BestBid
			bestAsk = orderbook.BestAsk
		}

		// Get best price based on side
		// BUY: use bestAsk (lowest sell price)
		// SELL: use bestBid (highest buy price)
		if input.Side == types.OrderSideBuy {
			pricePerShare = bestAsk
		} else {
			pricePerShare = bestBid
		}
	}

	if input.Strategy == types.OrderStrategyMarket {
		// For MARKET orders, calculate amounts with slippage
		// Get slippage multiplier (slippageBps is in basis points, e.g., 1000 = 10%)
		slippageMultiplier := decimal.NewFromInt(1)
		if input.SlippageBps > 0 {
			// Convert basis points to decimal multiplier
			// 1000 bps = 10% = 0.1, so multiplier = 1 + 0.1 = 1.1 for BUY (price goes up)
			// For SELL: multiplier = 1 - 0.1 = 0.9 (price goes down)
			slippageDecimal := decimal.NewFromInt(int64(input.SlippageBps)).Div(decimal.NewFromInt(10000))
			if input.Side == types.OrderSideBuy {
				slippageMultiplier = decimal.NewFromInt(1).Add(slippageDecimal) // BUY: price goes up
			} else {
				slippageMultiplier = decimal.NewFromInt(1).Sub(slippageDecimal) // SELL: price goes down
			}
		}

		// Apply slippage to price
		priceWithSlippage := pricePerShare.Mul(slippageMultiplier)
		quantityWei := input.Amount.Shift(constants.TokenDecimals) // Convert to wei

		// Debug logging for MARKET orders
		fmt.Fprintf(os.Stderr, "[DEBUG] MARKET order calculation: Side=%v, pricePerShare=%s, slippageBps=%d, slippageMultiplier=%s, priceWithSlippage=%s, input.Amount=%s\n",
			input.Side, pricePerShare.String(), input.SlippageBps, slippageMultiplier.String(), priceWithSlippage.String(), input.Amount.String())

		if input.Side == types.OrderSideBuy {
			// BUY: makerAmount is quoteToken (USDT), takerAmount is baseToken (shares)
			// makerAmount = quantity * priceWithSlippage (USDT to pay)
			// takerAmount = quantity (shares to receive)
			makerAmount = input.Amount.Mul(priceWithSlippage).Shift(constants.TokenDecimals)
			takerAmount = quantityWei
		} else {
			// SELL: makerAmount is baseToken (shares), takerAmount is quoteToken (USDT)
			// makerAmount = quantity (shares to give)
			// takerAmount = quantity * priceWithSlippage (USDT to receive)
			makerAmount = quantityWei
			takerAmount = input.Amount.Mul(priceWithSlippage).Shift(constants.TokenDecimals)
			fmt.Fprintf(os.Stderr, "[DEBUG] MARKET SELL: makerAmount=%s, takerAmount=%s (input.Amount=%s * priceWithSlippage=%s)\n",
				makerAmount.String(), takerAmount.String(), input.Amount.String(), priceWithSlippage.String())
		}
		// Update pricePerShare to include slippage for API
		pricePerShare = priceWithSlippage
	} else {
		// For LIMIT orders, calculate based on side
		// Amount is always in shares (quantity of outcome tokens)
		quantityWei := input.Amount.Shift(constants.TokenDecimals) // Convert to wei
		priceWei := input.Price.Shift(constants.TokenDecimals)     // Convert price to wei

		if input.Side == types.OrderSideBuy {
			// BUY: makerAmount is what you pay (USDT), takerAmount is what you get (shares)
			// makerAmount = price * quantity
			// takerAmount = quantity
			makerAmount = input.Amount.Mul(input.Price).Shift(constants.TokenDecimals)
			takerAmount = quantityWei
		} else {
			// SELL: makerAmount is what you give (shares), takerAmount is what you get (USDT)
			// makerAmount = quantity
			// takerAmount = price * quantity
			makerAmount = quantityWei
			takerAmount = input.Amount.Mul(input.Price).Shift(constants.TokenDecimals)
		}
		_ = priceWei // pricePerShare is already set
	}

	// Convert side to order-utils format
	var orderSide ordermodel.Side
	if input.Side == types.OrderSideBuy {
		orderSide = ordermodel.BUY
	} else {
		orderSide = ordermodel.SELL
	}

	// Build order data
	signerAddr := c.signer.GetAddress()

	// Expiration: use input or default
	expirationSeconds := input.ExpirationSeconds
	if expirationSeconds <= 0 {
		expirationSeconds = constants.DefaultOrderExpirationSeconds
	}
	expiration := time.Now().Unix() + expirationSeconds

	orderData := &ordermodel.OrderData{
		Maker:         signerAddr.Hex(),
		Signer:        signerAddr.Hex(),
		Taker:         common.HexToAddress("0x0").Hex(),
		TokenId:       tokenID,
		MakerAmount:   makerAmount.BigInt().String(),
		TakerAmount:   takerAmount.BigInt().String(),
		FeeRateBps:    fmt.Sprintf("%d", market.FeeRateBps),
		Side:          orderSide,
		SignatureType: predictcontracts.SignatureTypeEOA,
		Nonce:         "0",
		Expiration:    fmt.Sprintf("%d", expiration),
	}

	// Print full OrderData for debugging - use fmt.Fprintf to stderr to ensure immediate output
	fmt.Fprintf(os.Stderr, "[DEBUG] OrderData: Maker=%s, Signer=%s, TokenId=%s, MakerAmount=%s, TakerAmount=%s, Side=%v, FeeRateBps=%s, Expiration=%s\n",
		orderData.Maker, orderData.Signer, orderData.TokenId, orderData.MakerAmount, orderData.TakerAmount, orderData.Side, orderData.FeeRateBps, orderData.Expiration)

	// Build and sign order
	signedOrder, err := orderBuilder.BuildSignedOrder(c.signer, orderData, exchangeAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to build signed order: %w", err)
	}

	// Prepare API request
	signatureHex := fmt.Sprintf("0x%x", signedOrder.Signature)

	// Set slippage for MARKET orders
	slippageBps := "0"
	if input.Strategy == types.OrderStrategyMarket {
		if input.SlippageBps > 0 {
			slippageBps = fmt.Sprintf("%d", input.SlippageBps)
		} else {
			slippageBps = "10" // Default 0.1%
		}
	}

	// Convert pricePerShare to wei
	pricePerShareWei := pricePerShare.Shift(constants.TokenDecimals).BigInt().String()

	requestBody := map[string]interface{}{
		"data": map[string]interface{}{
			"order": map[string]interface{}{
				"salt":          signedOrder.Order.Salt.String(),
				"maker":         signedOrder.Order.Maker.Hex(),
				"signer":        signedOrder.Order.Signer.Hex(),
				"taker":         signedOrder.Order.Taker.Hex(),
				"tokenId":       signedOrder.Order.TokenId.String(),
				"makerAmount":   signedOrder.Order.MakerAmount.String(),
				"takerAmount":   signedOrder.Order.TakerAmount.String(),
				"expiration":    signedOrder.Order.Expiration.String(),
				"nonce":         signedOrder.Order.Nonce.String(),
				"feeRateBps":    signedOrder.Order.FeeRateBps.String(),
				"side":          signedOrder.Order.Side.Int64(),
				"signatureType": signedOrder.Order.SignatureType.Int64(),
				"signature":     signatureHex,
			},
			"pricePerShare": pricePerShareWei,
			"strategy":      string(input.Strategy),
			"slippageBps":   slippageBps,
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Send request
	respBody, err := c.doRequest("POST", constants.EndpointOrders, bodyBytes, true)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	// Parse response
	var response types.APIBaseResponse[types.PlaceOrderResult]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse place order response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	result := response.Data
	// Handle both orderHash and hash fields
	if result.OrderHash == "" {
		result.OrderHash = result.Hash
	}
	result.Success = true

	return &result, nil
}

// getExchangeAddress returns the exchange contract address based on market type
func (c *Client) getExchangeAddress(isNegRisk, isYieldBearing bool) (common.Address, error) {
	var exchangeType ordermodel.VerifyingContract
	if isYieldBearing {
		if isNegRisk {
			exchangeType = ordermodel.YieldBearingNegRiskCTFExchange
		} else {
			exchangeType = ordermodel.YieldBearingCTFExchange
		}
	} else {
		if isNegRisk {
			exchangeType = ordermodel.NegRiskCTFExchange
		} else {
			exchangeType = ordermodel.CTFExchange
		}
	}

	return utils.GetVerifyingContractAddress(c.chainID, exchangeType)
}

// CancelOrder cancels a single order by ID
func (c *Client) CancelOrder(orderID string) (*types.CancelOrderResult, error) {
	return c.CancelOrders(&types.CancelOrderInput{OrderIDs: []string{orderID}})
}

// CancelOrders cancels multiple orders by their IDs
func (c *Client) CancelOrders(input *types.CancelOrderInput) (*types.CancelOrderResult, error) {
	if c.jwtToken == "" {
		return nil, fmt.Errorf("JWT token is required for canceling orders")
	}
	if len(input.OrderIDs) == 0 {
		return nil, fmt.Errorf("at least one order ID is required")
	}

	requestBody := map[string]interface{}{
		"data": map[string]interface{}{
			"ids": input.OrderIDs,
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	respBody, err := c.doRequest("POST", constants.EndpointOrdersRemove, bodyBytes, true)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel orders: %w", err)
	}

	var response types.CancelOrderResult
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetOrders retrieves orders for the authenticated user
func (c *Client) GetOrders(opts *types.GetOrdersOptions) (*types.GetOrdersResponse, error) {
	if c.jwtToken == "" {
		return nil, fmt.Errorf("JWT token is required for getting orders")
	}

	path := constants.EndpointOrders
	params := url.Values{}

	if opts != nil {
		if opts.First != nil {
			params.Set("first", fmt.Sprintf("%d", *opts.First))
		}
		if opts.After != "" {
			params.Set("after", opts.After)
		}
		if !opts.MarketID.IsZero() {
			params.Set("marketId", opts.MarketID.String())
		}
		if opts.Status != "" {
			params.Set("status", opts.Status.String())
		}
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	var response types.GetOrdersResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return &response, nil
}

// GetOrderByHash retrieves a specific order by its hash
func (c *Client) GetOrderByHash(orderHash string) (*types.Order, error) {
	if c.jwtToken == "" {
		return nil, fmt.Errorf("JWT token is required for getting order")
	}

	path := constants.EndpointOrders + "/" + url.QueryEscape(orderHash)

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	var response types.APIBaseResponse[types.Order]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse order response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	return &response.Data, nil
}

// GetOrderMatches retrieves order match events with optional filtering
func (c *Client) GetOrderMatches(opts *types.GetOrderMatchesOptions) (*types.MatchEventResponse, error) {
	if c.jwtToken == "" {
		return nil, fmt.Errorf("JWT token is required for getting matches")
	}

	path := constants.EndpointOrdersMatches
	params := url.Values{}

	if opts != nil {
		if opts.First != nil {
			params.Set("first", fmt.Sprintf("%d", *opts.First))
		}
		if opts.After != "" {
			params.Set("after", opts.After)
		}
		if opts.CategoryID != "" {
			params.Set("categoryId", opts.CategoryID.String())
		}
		if !opts.MarketID.IsZero() {
			params.Set("marketId", opts.MarketID.String())
		}
		if opts.MinValueUsdtWei != "" {
			params.Set("minValueUsdtWei", opts.MinValueUsdtWei)
		}
		if opts.SignerAddress != "" {
			params.Set("signerAddress", opts.SignerAddress)
		}
		if opts.IsSignerMaker != nil {
			if *opts.IsSignerMaker {
				params.Set("isSignerMaker", "true")
			} else {
				params.Set("isSignerMaker", "false")
			}
		}
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	var response types.MatchEventResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return &response, nil
}
