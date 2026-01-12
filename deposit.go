package predictclob

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ivanzzeth/ethtypes/contracts/erc20"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
)

// GetUSDTAddress returns the USDT contract address for the given chain ID
func GetUSDTAddress(chainID int64) (string, error) {
	switch chainID {
	case constants.ChainIDBNBMainnet:
		return constants.USDTAddressBNBMainnet, nil
	case constants.ChainIDBNBTestnet:
		return constants.USDTAddressBNBTestnet, nil
	default:
		return "", fmt.Errorf("unsupported chain ID: %d", chainID)
	}
}

// GetChainName returns the chain name for the given chain ID
func GetChainName(chainID int64) string {
	switch chainID {
	case constants.ChainIDBNBMainnet:
		return "BNB Mainnet"
	case constants.ChainIDBNBTestnet:
		return "BNB Testnet"
	default:
		return fmt.Sprintf("Chain %d", chainID)
	}
}

// GetUSDTBalance gets the USDT balance of a wallet address on-chain
// Requires an RPC URL to connect to the blockchain
func GetUSDTBalance(rpcURL string, chainID int64, address string) (*types.USDTBalance, error) {
	// Get USDT contract address
	usdtAddress, err := GetUSDTAddress(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get USDT address: %w", err)
	}

	// Connect to RPC
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	// Create ERC20 contract instance
	erc20Contract, err := erc20.NewErc20(common.HexToAddress(usdtAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to create ERC20 contract: %w", err)
	}

	// Get balance
	balanceWei, err := erc20Contract.BalanceOf(nil, common.HexToAddress(address))
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Convert to human-readable format (USDT has 18 decimals)
	balanceUSDT := new(big.Float).Quo(
		new(big.Float).SetInt(balanceWei),
		new(big.Float).SetInt(big.NewInt(1e18)),
	)

	return &types.USDTBalance{
		Address:     address,
		BalanceWei:  balanceWei,
		BalanceUSDT: balanceUSDT.Text('f', 6), // 6 decimal places
	}, nil
}

// GetDepositInfo returns deposit information for a wallet address
func GetDepositInfo(chainID int64, walletAddress string) (*types.DepositInfo, error) {
	// Ensure address is checksum format
	address := common.HexToAddress(walletAddress)
	walletAddress = address.Hex()

	// Get USDT address
	usdtAddress, err := GetUSDTAddress(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get USDT address: %w", err)
	}

	// Get chain name
	chainName := GetChainName(chainID)

	return &types.DepositInfo{
		WalletAddress: walletAddress,
		USDTAddress:   usdtAddress,
		ChainID:        chainID,
		ChainName:      chainName,
	}, nil
}

// GetDepositInfoFromPrivateKey returns deposit information from a private key
func GetDepositInfoFromPrivateKey(chainID int64, privateKeyHex string) (*types.DepositInfo, common.Address, error) {
	// Remove 0x prefix if present
	if len(privateKeyHex) >= 2 && privateKeyHex[0:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	// Parse private key to get address
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, common.Address{}, fmt.Errorf("failed to get public key")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Get deposit info
	depositInfo, err := GetDepositInfo(chainID, address.Hex())
	if err != nil {
		return nil, common.Address{}, err
	}

	return depositInfo, address, nil
}
