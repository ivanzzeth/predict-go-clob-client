package predictclob

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/predict-go-clob-client/types"
	"github.com/shopspring/decimal"
)

// EnableTrading enables trading by approving necessary tokens on both YB and NYB contracts
// Returns transaction hashes for approval transactions
func (c *Client) EnableTrading(ctx context.Context) ([]common.Hash, error) {
	if c.contractInterface == nil {
		return nil, fmt.Errorf("contract interface is not initialized. Please provide RPC URL and signer when creating the client")
	}

	txHashes, err := c.contractInterface.EnableTrading(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to enable trading: %w", err)
	}

	return txHashes, nil
}

// Split splits collateral into outcome tokens for a market
// amount is in human-readable decimal format (e.g., "1.5" means 1.5 USDT)
// isYieldBearing specifies whether to use Yield Bearing contracts
func (c *Client) Split(ctx context.Context, conditionID [32]byte, amount decimal.Decimal, isYieldBearing bool) (common.Hash, error) {
	if c.contractInterface == nil {
		return common.Hash{}, fmt.Errorf("contract interface is not initialized. Please provide RPC URL and signer when creating the client")
	}

	// Convert amount to wei (18 decimals for USDT on BNB)
	amountWei := amount.Shift(18).BigInt()

	// Get the appropriate contract address based on isYieldBearing
	var ctfAddr common.Address
	if isYieldBearing {
		ctfAddr = c.contractInterface.GetConfig().YieldBearingConditionalTokens
	} else {
		ctfAddr = c.contractInterface.GetConfig().ConditionalTokens
	}

	txHash, err := c.contractInterface.Split(ctx, ctfAddr, conditionID, amountWei)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to split: %w", err)
	}

	return txHash, nil
}

// SplitByMarketID splits collateral into outcome tokens using market ID
// amount is in human-readable decimal format (e.g., "1.5" means 1.5 USDT)
func (c *Client) SplitByMarketID(ctx context.Context, marketID types.MarketID, amount decimal.Decimal) (common.Hash, error) {
	// Get market details to retrieve conditionId and isYieldBearing
	market, err := c.GetMarket(marketID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get market: %w", err)
	}

	// Get condition ID from market outcomes
	if len(market.Outcomes) == 0 {
		return common.Hash{}, fmt.Errorf("market has no outcomes")
	}

	// Extract conditionId from the first outcome's onChainId
	// The conditionId is derived from the tokenId
	conditionID, err := c.getConditionIDFromMarket(market)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get conditionId: %w", err)
	}

	return c.Split(ctx, conditionID, amount, market.IsYieldBearing)
}

// Merge merges outcome tokens back into collateral
// amount is in human-readable decimal format (e.g., "1.5" means 1.5 tokens)
// isYieldBearing specifies whether to use Yield Bearing contracts
func (c *Client) Merge(ctx context.Context, conditionID [32]byte, amount decimal.Decimal, isYieldBearing bool) (common.Hash, error) {
	if c.contractInterface == nil {
		return common.Hash{}, fmt.Errorf("contract interface is not initialized. Please provide RPC URL and signer when creating the client")
	}

	// Convert amount to wei (18 decimals)
	amountWei := amount.Shift(18).BigInt()

	// Get the appropriate contract address based on isYieldBearing
	var ctfAddr common.Address
	if isYieldBearing {
		ctfAddr = c.contractInterface.GetConfig().YieldBearingConditionalTokens
	} else {
		ctfAddr = c.contractInterface.GetConfig().ConditionalTokens
	}

	txHash, err := c.contractInterface.Merge(ctx, ctfAddr, conditionID, amountWei)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to merge: %w", err)
	}

	return txHash, nil
}

// MergeByMarketID merges outcome tokens back into collateral using market ID
// amount is in human-readable decimal format (e.g., "1.5" means 1.5 tokens)
func (c *Client) MergeByMarketID(ctx context.Context, marketID types.MarketID, amount decimal.Decimal) (common.Hash, error) {
	// Get market details
	market, err := c.GetMarket(marketID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get market: %w", err)
	}

	// Get condition ID from market
	conditionID, err := c.getConditionIDFromMarket(market)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get conditionId: %w", err)
	}

	return c.Merge(ctx, conditionID, amount, market.IsYieldBearing)
}

// Redeem redeems positions for a resolved market
// isYieldBearing specifies whether to use Yield Bearing contracts
func (c *Client) Redeem(ctx context.Context, conditionID [32]byte, isYieldBearing bool) (common.Hash, error) {
	if c.contractInterface == nil {
		return common.Hash{}, fmt.Errorf("contract interface is not initialized. Please provide RPC URL and signer when creating the client")
	}

	// Get the appropriate contract address based on isYieldBearing
	var ctfAddr common.Address
	if isYieldBearing {
		ctfAddr = c.contractInterface.GetConfig().YieldBearingConditionalTokens
	} else {
		ctfAddr = c.contractInterface.GetConfig().ConditionalTokens
	}

	txHash, err := c.contractInterface.Redeem(ctx, ctfAddr, conditionID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to redeem: %w", err)
	}

	return txHash, nil
}

// RedeemByMarketID redeems positions for a resolved market using market ID
func (c *Client) RedeemByMarketID(ctx context.Context, marketID types.MarketID) (common.Hash, error) {
	// Get market details
	market, err := c.GetMarket(marketID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get market: %w", err)
	}

	// Get condition ID from market
	conditionID, err := c.getConditionIDFromMarket(market)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get conditionId: %w", err)
	}

	return c.Redeem(ctx, conditionID, market.IsYieldBearing)
}

// getConditionIDFromMarket extracts the conditionId from market data
// The conditionId is typically embedded in the tokenId
func (c *Client) getConditionIDFromMarket(market *types.Market) ([32]byte, error) {
	if len(market.Outcomes) == 0 {
		return [32]byte{}, fmt.Errorf("market has no outcomes")
	}

	// Get the first outcome's onChainId (tokenId)
	tokenIDStr := string(market.Outcomes[0].OnChainID)
	if tokenIDStr == "" {
		return [32]byte{}, fmt.Errorf("outcome onChainId is empty")
	}

	// Parse tokenId as big.Int
	tokenID, ok := new(big.Int).SetString(tokenIDStr, 10)
	if !ok {
		return [32]byte{}, fmt.Errorf("failed to parse tokenId: %s", tokenIDStr)
	}

	// Extract conditionId from tokenId
	// In Gnosis CTF, tokenId = conditionId ^ (1 << indexSet)
	// For indexSet = 1, tokenId = conditionId ^ 2
	// So conditionId = tokenId ^ 2 (for indexSet = 1)
	indexSet := big.NewInt(int64(market.Outcomes[0].IndexSet))
	mask := new(big.Int).Lsh(big.NewInt(1), uint(indexSet.Uint64()))
	conditionIDBig := new(big.Int).Xor(tokenID, mask)

	// Convert to [32]byte
	conditionIDBytes := conditionIDBig.Bytes()
	var conditionID [32]byte
	copy(conditionID[32-len(conditionIDBytes):], conditionIDBytes)

	return conditionID, nil
}

// SplitNegRisk splits collateral into outcome tokens for a neg-risk market
// amount is in human-readable decimal format (e.g., "1.5" means 1.5 USDT)
// isYieldBearing specifies whether to use Yield Bearing contracts
func (c *Client) SplitNegRisk(ctx context.Context, conditionID [32]byte, amount decimal.Decimal, isYieldBearing bool) (common.Hash, error) {
	if c.contractInterface == nil {
		return common.Hash{}, fmt.Errorf("contract interface is not initialized. Please provide RPC URL and signer when creating the client")
	}

	// Convert amount to wei (18 decimals for USDT on BNB)
	amountWei := amount.Shift(18).BigInt()

	// Get the appropriate contract address based on isYieldBearing
	var negRiskAdapterAddr common.Address
	if isYieldBearing {
		negRiskAdapterAddr = c.contractInterface.GetConfig().YieldBearingNegRiskAdapter
	} else {
		negRiskAdapterAddr = c.contractInterface.GetConfig().NegRiskAdapter
	}

	txHash, err := c.contractInterface.SplitNegRisk(ctx, negRiskAdapterAddr, conditionID, amountWei)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to split neg-risk: %w", err)
	}

	return txHash, nil
}

// MergeNegRisk merges outcome tokens back into collateral for a neg-risk market
// amount is in human-readable decimal format (e.g., "1.5" means 1.5 tokens)
// isYieldBearing specifies whether to use Yield Bearing contracts
func (c *Client) MergeNegRisk(ctx context.Context, conditionID [32]byte, amount decimal.Decimal, isYieldBearing bool) (common.Hash, error) {
	if c.contractInterface == nil {
		return common.Hash{}, fmt.Errorf("contract interface is not initialized. Please provide RPC URL and signer when creating the client")
	}

	// Convert amount to wei (18 decimals)
	amountWei := amount.Shift(18).BigInt()

	// Get the appropriate contract address based on isYieldBearing
	var negRiskAdapterAddr common.Address
	if isYieldBearing {
		negRiskAdapterAddr = c.contractInterface.GetConfig().YieldBearingNegRiskAdapter
	} else {
		negRiskAdapterAddr = c.contractInterface.GetConfig().NegRiskAdapter
	}

	txHash, err := c.contractInterface.MergeNegRisk(ctx, negRiskAdapterAddr, conditionID, amountWei)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to merge neg-risk: %w", err)
	}

	return txHash, nil
}

// RedeemNegRisk redeems positions for a resolved neg-risk market
// amounts specifies the amount of each outcome token to redeem
// isYieldBearing specifies whether to use Yield Bearing contracts
func (c *Client) RedeemNegRisk(ctx context.Context, conditionID [32]byte, amounts []*big.Int, isYieldBearing bool) (common.Hash, error) {
	if c.contractInterface == nil {
		return common.Hash{}, fmt.Errorf("contract interface is not initialized. Please provide RPC URL and signer when creating the client")
	}

	// Get the appropriate contract address based on isYieldBearing
	var negRiskAdapterAddr common.Address
	if isYieldBearing {
		negRiskAdapterAddr = c.contractInterface.GetConfig().YieldBearingNegRiskAdapter
	} else {
		negRiskAdapterAddr = c.contractInterface.GetConfig().NegRiskAdapter
	}

	txHash, err := c.contractInterface.RedeemNegRisk(ctx, negRiskAdapterAddr, conditionID, amounts)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to redeem neg-risk: %w", err)
	}

	return txHash, nil
}
