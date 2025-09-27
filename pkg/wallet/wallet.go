package wallet

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	
	localcrypto "wallet-transfer/pkg/crypto"
)

// Wallet represents a blockchain wallet
type Wallet struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

// NewWallet creates a new wallet from a private key hex string
func NewWallet(privateKeyHex string) (*Wallet, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Wallet{
		privateKey: privateKey,
		address:    address,
	}, nil
}

// GetAddress returns the wallet address
func (w *Wallet) GetAddress() common.Address {
	return w.address
}

// GetPrivateKey returns the private key (use with caution)
func (w *Wallet) GetPrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

// SignTransaction signs a transaction with the wallet's private key
func (w *Wallet) SignTransaction(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), w.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return signedTx, nil
}

// CreateTransaction creates a new transaction
func (w *Wallet) CreateTransaction(to common.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int, nonce uint64, data []byte) *types.Transaction {
	return types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)
}

// WalletManager manages multiple wallets
type WalletManager struct {
	wallets []*Wallet
}

// NewWalletManager creates a new WalletManager
func NewWalletManager() *WalletManager {
	return &WalletManager{
		wallets: make([]*Wallet, 0),
	}
}

// LoadWallets loads wallets from private key strings
func (wm *WalletManager) LoadWallets(privateKeys []string) error {
	wm.wallets = make([]*Wallet, 0, len(privateKeys))

	for i, privateKey := range privateKeys {
		wallet, err := NewWallet(privateKey)
		if err != nil {
			return fmt.Errorf("failed to create wallet %d: %w", i, err)
		}
		wm.wallets = append(wm.wallets, wallet)
	}

	return nil
}

// GetWallets returns all wallets
func (wm *WalletManager) GetWallets() []*Wallet {
	return wm.wallets
}

// GetWallet returns a wallet by index
func (wm *WalletManager) GetWallet(index int) (*Wallet, error) {
	if index < 0 || index >= len(wm.wallets) {
		return nil, fmt.Errorf("wallet index %d out of range", index)
	}
	return wm.wallets[index], nil
}

// GetWalletCount returns the number of wallets
func (wm *WalletManager) GetWalletCount() int {
	return len(wm.wallets)
}

// GetAddresses returns all wallet addresses
func (wm *WalletManager) GetAddresses() []common.Address {
	addresses := make([]common.Address, len(wm.wallets))
	for i, wallet := range wm.wallets {
		addresses[i] = wallet.GetAddress()
	}
	return addresses
}

// GetPrivateKey 根据地址获取私钥
func (wm *WalletManager) GetPrivateKey(address common.Address) (*localcrypto.PrivateKey, error) {
	for _, wallet := range wm.wallets {
		if wallet.GetAddress() == address {
			// 将ecdsa.PrivateKey转换为crypto.PrivateKey
			privateKey, err := localcrypto.NewPrivateKeyFromECDSA(wallet.GetPrivateKey())
			if err != nil {
				return nil, fmt.Errorf("failed to convert private key: %w", err)
			}
			return privateKey, nil
		}
	}
	
	return nil, fmt.Errorf("private key not found for address: %s", address.Hex())
}