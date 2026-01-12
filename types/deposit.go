package types

import "math/big"

// DepositInfo represents deposit information
type DepositInfo struct {
	WalletAddress   string
	USDTAddress     string
	ChainID         int64
	ChainName       string
}

// USDTBalance represents USDT balance information
type USDTBalance struct {
	Address     string
	BalanceWei  *big.Int
	BalanceUSDT string // Human-readable format
}
