package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// PrivateKey 私钥包装器
type PrivateKey struct {
	key *ecdsa.PrivateKey
}

// NewPrivateKeyFromHex 从十六进制字符串创建私钥
func NewPrivateKeyFromHex(hexKey string) (*PrivateKey, error) {
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key hex: %w", err)
	}
	
	return &PrivateKey{key: key}, nil
}

// NewPrivateKeyFromECDSA 从ecdsa.PrivateKey创建私钥
func NewPrivateKeyFromECDSA(key *ecdsa.PrivateKey) (*PrivateKey, error) {
	if key == nil {
		return nil, fmt.Errorf("private key cannot be nil")
	}
	
	return &PrivateKey{key: key}, nil
}

// GetAddress 获取地址
func (pk *PrivateKey) GetAddress() common.Address {
	publicKey := pk.key.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}
	}
	
	return crypto.PubkeyToAddress(*publicKeyECDSA)
}

// SignTransaction 签名交易
func (pk *PrivateKey) SignTransaction(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), pk.key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}
	
	return signedTx, nil
}

// GetECDSAKey 获取底层的ECDSA私钥
func (pk *PrivateKey) GetECDSAKey() *ecdsa.PrivateKey {
	return pk.key
}

// ToHex 转换为十六进制字符串
func (pk *PrivateKey) ToHex() string {
	return fmt.Sprintf("%x", crypto.FromECDSA(pk.key))
}

// GetPublicKey 获取公钥
func (pk *PrivateKey) GetPublicKey() *ecdsa.PublicKey {
	publicKey := pk.key.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil
	}
	
	return publicKeyECDSA
}