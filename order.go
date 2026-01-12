package predictclob

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	predictcontracts "github.com/ivanzzeth/predict-go-contracts"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
	"github.com/ivanzzeth/predict-go-order-utils/pkg/builder"
	ordermodel "github.com/ivanzzeth/predict-go-order-utils/pkg/model"
	"github.com/ivanzzeth/predict-go-order-utils/pkg/utils"
	"github.com/shopspring/decimal"
)

// OrderSide represents the side of an order
type OrderSide int

const (
	OrderSideBuy  OrderSide = 0
	OrderSideSell OrderSide = 1
)

// OrderStrategy represents the order strategy/type
type OrderStrategy string

const (
	OrderStrategyLimit  OrderStrategy = "LIMIT"
	OrderStrategyMarket OrderStrategy = "MARKET"
)

// PlaceOrderInput represents input for placing an order
type PlaceOrderInput struct {
	MarketID          types.MarketID  // Market ID
	TokenID           types.TokenID   // Token ID (onChainId from outcome)
	Side              OrderSide       // BUY or SELL
	Strategy          OrderStrategy   // LIMIT or MARKET
	Amount            decimal.Decimal // Amount in USDT (for BUY) or shares (for SELL)
	Price             decimal.Decimal // Price per share (required for LIMIT orders)
	SlippageBps       int             // Slippage in basis points (default 10 for MARKET orders)
	ExpirationSeconds int64           // Order expiration in seconds from now (default: constants.DefaultOrderExpirationSeconds)
}

// PlaceOrderResult represents the result of placing an order
type PlaceOrderResult struct {
	OrderID   string `json:"orderId"`
	OrderHash string `json:"orderHash"`
	Success   bool   `json:"success"`
}

// PlaceOrder places a limit or market order on the Predict exchange
func (c *Client) PlaceOrder(input *PlaceOrderInput) (*PlaceOrderResult, error) {
	if c.signer == nil {
		return nil, fmt.Errorf("signer is required for placing orders. Use WithSigner option when creating client")
	}
	if c.jwtToken == "" {
		return nil, fmt.Errorf("JWT token is required for placing orders. Use WithJWTToken option or authenticate first")
	}

	// Get market info to get feeRateBps and other parameters
	market, err := c.GetMarket(input.MarketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get market: %w", err)
	}

	// Validate price for LIMIT orders
	if input.Strategy == OrderStrategyLimit && input.Price.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("price is required and must be positive for LIMIT orders")
	}

	// Get token ID from input or market outcomes
	tokenID := string(input.TokenID)
	if tokenID == "" {
		// Use outcome based on side
		if len(market.Outcomes) == 0 {
			return nil, fmt.Errorf("market has no outcomes")
		}

		// For BUY, use first outcome (indexSet=1, usually "Yes"/"Up")
		// For SELL, use second outcome if available (indexSet=2, usually "No"/"Down")
		if input.Side == OrderSideBuy {
			for _, outcome := range market.Outcomes {
				if outcome.IndexSet == 1 {
					tokenID = string(outcome.OnChainID)
					break
				}
			}
			if tokenID == "" {
				tokenID = string(market.Outcomes[0].OnChainID)
			}
		} else {
			for _, outcome := range market.Outcomes {
				if outcome.IndexSet == 2 {
					tokenID = string(outcome.OnChainID)
					break
				}
			}
			if tokenID == "" && len(market.Outcomes) > 1 {
				tokenID = string(market.Outcomes[1].OnChainID)
			} else if tokenID == "" {
				tokenID = string(market.Outcomes[0].OnChainID)
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

	if input.Strategy == OrderStrategyMarket {
		// For MARKET orders, use 0 price and get amounts from orderbook
		pricePerShare = decimal.Zero
		makerAmount = input.Amount.Shift(constants.TokenDecimals) // Convert to wei
		takerAmount = decimal.Zero
	} else {
		// For LIMIT orders, calculate based on side
		// Amount is always in shares (quantity of outcome tokens)
		quantityWei := input.Amount.Shift(constants.TokenDecimals) // Convert to wei
		priceWei := input.Price.Shift(constants.TokenDecimals)     // Convert price to wei

		if input.Side == OrderSideBuy {
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
	if input.Side == OrderSideBuy {
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
		FeeRateBps:    market.FeeRateBps,
		Side:          orderSide,
		SignatureType: predictcontracts.SignatureTypeEOA,
		Nonce:         "0",
		Expiration:    fmt.Sprintf("%d", expiration),
	}

	// Build and sign order
	signedOrder, err := orderBuilder.BuildSignedOrder(c.signer, orderData, exchangeAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to build signed order: %w", err)
	}

	// Prepare API request
	signatureHex := fmt.Sprintf("0x%x", signedOrder.Signature)

	// Set slippage for MARKET orders
	slippageBps := "0"
	if input.Strategy == OrderStrategyMarket {
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
	respBody, err := c.doRequest("POST", constants.EndpointOrders, bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	// Parse response
	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Extract result
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var resultData struct {
		OrderID   string `json:"orderId"`
		OrderHash string `json:"orderHash"`
		Hash      string `json:"hash"`
	}
	if err := json.Unmarshal(dataBytes, &resultData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	orderHashStr := resultData.OrderHash
	if orderHashStr == "" {
		orderHashStr = resultData.Hash
	}

	return &PlaceOrderResult{
		OrderID:   resultData.OrderID,
		OrderHash: orderHashStr,
		Success:   true,
	}, nil
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

// CancelOrderInput represents input for canceling orders
type CancelOrderInput struct {
	OrderIDs []string // Order IDs to cancel
}

// CancelOrderResult represents the result of canceling orders
type CancelOrderResult struct {
	Removed []string `json:"removed"`
	Noop    []string `json:"noop"`
	Success bool     `json:"success"`
}

// CancelOrder cancels a single order by ID
func (c *Client) CancelOrder(orderID string) (*CancelOrderResult, error) {
	return c.CancelOrders(&CancelOrderInput{OrderIDs: []string{orderID}})
}

// CancelOrders cancels multiple orders by their IDs
func (c *Client) CancelOrders(input *CancelOrderInput) (*CancelOrderResult, error) {
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

	respBody, err := c.doRequest("POST", constants.EndpointOrdersRemove, bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel orders: %w", err)
	}

	var response struct {
		Success bool     `json:"success"`
		Removed []string `json:"removed"`
		Noop    []string `json:"noop"`
		Message string   `json:"message"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &CancelOrderResult{
		Removed: response.Removed,
		Noop:    response.Noop,
		Success: response.Success,
	}, nil
}

// GetOrdersOptions represents options for getting orders
type GetOrdersOptions struct {
	MarketID types.MarketID
	Status   string // "OPEN", "FILLED", "CANCELLED"
}

// Order represents an order from the API
type Order struct {
	ID          string        `json:"id"`
	Hash        string        `json:"hash"`
	MarketID    string        `json:"marketId"`
	Side        int           `json:"side"`
	Strategy    string        `json:"strategy"`
	Status      string        `json:"status"`
	Price       string        `json:"price"`
	Amount      string        `json:"amount"`
	MakerAmount string        `json:"makerAmount"`
	TakerAmount string        `json:"takerAmount"`
	CreatedAt   time.Time     `json:"createdAt"`
	OrderData   OrderDataInfo `json:"order"`
}

// OrderDataInfo represents the nested order data
type OrderDataInfo struct {
	Salt          string `json:"salt"`
	Maker         string `json:"maker"`
	Signer        string `json:"signer"`
	Taker         string `json:"taker"`
	TokenID       string `json:"tokenId"`
	MakerAmount   string `json:"makerAmount"`
	TakerAmount   string `json:"takerAmount"`
	Expiration    string `json:"expiration"`
	Nonce         string `json:"nonce"`
	FeeRateBps    string `json:"feeRateBps"`
	Side          int    `json:"side"`
	SignatureType int    `json:"signatureType"`
	Signature     string `json:"signature"`
	Hash          string `json:"hash"`
}

// GetOrders retrieves orders for the authenticated user
func (c *Client) GetOrders(opts *GetOrdersOptions) ([]Order, error) {
	if c.jwtToken == "" {
		return nil, fmt.Errorf("JWT token is required for getting orders")
	}

	path := constants.EndpointOrders
	params := ""

	if opts != nil {
		if opts.MarketID != "" {
			if params != "" {
				params += "&"
			}
			params += "marketId=" + string(opts.MarketID)
		}
		if opts.Status != "" {
			if params != "" {
				params += "&"
			}
			params += "status=" + opts.Status
		}
	}

	if params != "" {
		path += "?" + params
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	// Parse data
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// Try parsing as array first
	var orders []Order
	if err := json.Unmarshal(dataBytes, &orders); err == nil {
		return orders, nil
	}

	// Try parsing as object with orders field
	var ordersResp struct {
		Orders []Order `json:"orders"`
	}
	if err := json.Unmarshal(dataBytes, &ordersResp); err == nil {
		return ordersResp.Orders, nil
	}

	return nil, fmt.Errorf("failed to parse orders from response")
}

// GetOrder retrieves a specific order by ID or hash
func (c *Client) GetOrder(orderIDOrHash string) (*Order, error) {
	if c.jwtToken == "" {
		return nil, fmt.Errorf("JWT token is required for getting order")
	}

	path := constants.EndpointOrders + "/" + orderIDOrHash

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	var response types.APIBaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false: %s", response.Message)
	}

	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var order Order
	if err := json.Unmarshal(dataBytes, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}
