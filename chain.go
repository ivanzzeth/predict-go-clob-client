package predictclob

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
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
// Uses cache if enabled (useCache=true) as split only needs conditionID and isYieldBearing
// Automatically uses SplitNegRisk if market.IsNegRisk is true
func (c *Client) Split(ctx context.Context, marketID types.MarketID, amount decimal.Decimal) (common.Hash, error) {
	if c.contractInterface == nil {
		return common.Hash{}, fmt.Errorf("contract interface is not initialized. Please provide RPC URL and signer when creating the client")
	}

	// Get market details to retrieve conditionId, isYieldBearing, and isNegRisk
	// useCache=true for split/merge/redeem operations as they only need conditionID and isYieldBearing
	market, err := c.GetMarket(marketID, true)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get market: %w", err)
	}

	// Convert amount to wei (TokenDecimals decimals for USDT on BNB)
	amountWei := amount.Shift(constants.TokenDecimals).BigInt()

	// Use NegRisk methods if market is neg-risk
	if market.IsNegRisk {
		// Get the appropriate adapter address based on isYieldBearing
		var negRiskAdapterAddr common.Address
		if market.IsYieldBearing {
			negRiskAdapterAddr = c.contractInterface.GetConfig().YieldBearingNegRiskAdapter
		} else {
			negRiskAdapterAddr = c.contractInterface.GetConfig().NegRiskAdapter
		}

		txHash, err := c.contractInterface.SplitNegRisk(ctx, negRiskAdapterAddr, market.ConditionID, amountWei)
		if err != nil {
			return common.Hash{}, fmt.Errorf("failed to split neg-risk: %w", err)
		}
		return txHash, nil
	}

	// Use regular split for non-neg-risk markets
	// Get the appropriate contract address based on isYieldBearing
	var ctfAddr common.Address
	if market.IsYieldBearing {
		ctfAddr = c.contractInterface.GetConfig().YieldBearingConditionalTokens
	} else {
		ctfAddr = c.contractInterface.GetConfig().ConditionalTokens
	}

	txHash, err := c.contractInterface.Split(ctx, ctfAddr, market.ConditionID, amountWei)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to split: %w", err)
	}

	return txHash, nil
}

// Merge merges outcome tokens back into collateral
// amount is in human-readable decimal format (e.g., "1.5" means 1.5 tokens)
// Uses cache if enabled (useCache=true) as merge only needs conditionID and isYieldBearing
// Automatically uses MergeNegRisk if market.IsNegRisk is true
func (c *Client) Merge(ctx context.Context, marketID types.MarketID, amount decimal.Decimal) (common.Hash, error) {
	if c.contractInterface == nil {
		return common.Hash{}, fmt.Errorf("contract interface is not initialized. Please provide RPC URL and signer when creating the client")
	}

	// Get market details to retrieve conditionId, isYieldBearing, and isNegRisk
	// useCache=true for split/merge/redeem operations as they only need conditionID and isYieldBearing
	market, err := c.GetMarket(marketID, true)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get market: %w", err)
	}

	// Convert amount to wei (TokenDecimals decimals)
	amountWei := amount.Shift(constants.TokenDecimals).BigInt()

	// Use NegRisk methods if market is neg-risk
	if market.IsNegRisk {
		// Get the appropriate adapter address based on isYieldBearing
		var negRiskAdapterAddr common.Address
		if market.IsYieldBearing {
			negRiskAdapterAddr = c.contractInterface.GetConfig().YieldBearingNegRiskAdapter
		} else {
			negRiskAdapterAddr = c.contractInterface.GetConfig().NegRiskAdapter
		}

		txHash, err := c.contractInterface.MergeNegRisk(ctx, negRiskAdapterAddr, market.ConditionID, amountWei)
		if err != nil {
			return common.Hash{}, fmt.Errorf("failed to merge neg-risk: %w", err)
		}
		return txHash, nil
	}

	// Use regular merge for non-neg-risk markets
	// Get the appropriate contract address based on isYieldBearing
	var ctfAddr common.Address
	if market.IsYieldBearing {
		ctfAddr = c.contractInterface.GetConfig().YieldBearingConditionalTokens
	} else {
		ctfAddr = c.contractInterface.GetConfig().ConditionalTokens
	}

	txHash, err := c.contractInterface.Merge(ctx, ctfAddr, market.ConditionID, amountWei)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to merge: %w", err)
	}

	return txHash, nil
}

// Redeem redeems positions for a resolved market
// Uses cache if enabled (useCache=true) as redeem only needs conditionID and isYieldBearing
// Automatically uses RedeemNegRisk if market.IsNegRisk is true
// amounts is required for neg-risk markets, optional (nil) for regular markets
func (c *Client) Redeem(ctx context.Context, marketID types.MarketID, amounts []*big.Int) (common.Hash, error) {
	if c.contractInterface == nil {
		return common.Hash{}, fmt.Errorf("contract interface is not initialized. Please provide RPC URL and signer when creating the client")
	}

	// Get market details to retrieve conditionId, isYieldBearing, and isNegRisk
	// useCache=true for split/merge/redeem operations as they only need conditionID and isYieldBearing
	market, err := c.GetMarket(marketID, true)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get market: %w", err)
	}

	// Use NegRisk methods if market is neg-risk
	if market.IsNegRisk {
		if len(amounts) == 0 {
			return common.Hash{}, fmt.Errorf("amounts parameter is required for neg-risk market redeem")
		}

		// Get the appropriate adapter address based on isYieldBearing
		var negRiskAdapterAddr common.Address
		if market.IsYieldBearing {
			negRiskAdapterAddr = c.contractInterface.GetConfig().YieldBearingNegRiskAdapter
		} else {
			negRiskAdapterAddr = c.contractInterface.GetConfig().NegRiskAdapter
		}

		txHash, err := c.contractInterface.RedeemNegRisk(ctx, negRiskAdapterAddr, market.ConditionID, amounts)
		if err != nil {
			return common.Hash{}, fmt.Errorf("failed to redeem neg-risk: %w", err)
		}
		return txHash, nil
	}

	// Use regular redeem for non-neg-risk markets
	// Get the appropriate contract address based on isYieldBearing
	var ctfAddr common.Address
	if market.IsYieldBearing {
		ctfAddr = c.contractInterface.GetConfig().YieldBearingConditionalTokens
	} else {
		ctfAddr = c.contractInterface.GetConfig().ConditionalTokens
	}

	txHash, err := c.contractInterface.Redeem(ctx, ctfAddr, market.ConditionID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to redeem: %w", err)
	}

	return txHash, nil
}

// SplitNegRisk splits collateral into outcome tokens for a neg-risk market
// amount is in human-readable decimal format (e.g., "1.5" means 1.5 USDT)
// isYieldBearing specifies whether to use Yield Bearing contracts
func (c *Client) SplitNegRisk(ctx context.Context, conditionID [32]byte, amount decimal.Decimal, isYieldBearing bool) (common.Hash, error) {
	if c.contractInterface == nil {
		return common.Hash{}, fmt.Errorf("contract interface is not initialized. Please provide RPC URL and signer when creating the client")
	}

	// Convert amount to wei (TokenDecimals decimals for USDT on BNB)
	amountWei := amount.Shift(constants.TokenDecimals).BigInt()

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

	// Convert amount to wei (TokenDecimals decimals)
	amountWei := amount.Shift(constants.TokenDecimals).BigInt()

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
