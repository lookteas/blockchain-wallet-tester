package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	client *ethclient.Client
}

func NewClient(rpcURL string) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain client: %v", err)
	}
	return &Client{client: client}, nil
}

func (c *Client) GetBalance(address common.Address) (*big.Int, error) {
	balance, err := c.client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}
	return balance, nil
}

func (c *Client) SendTransaction(from *Wallet, to common.Address, value *big.Int) (*types.Transaction, error) {
	// 获取 nonce
	nonce, err := c.client.PendingNonceAt(context.Background(), from.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	// 获取 gas price
	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}

	// 创建交易
	tx := types.NewTransaction(nonce, to, value, 21000, gasPrice, nil)

	// 签名交易
	chainID, err := c.client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), from.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	// 发送交易
	err = c.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %v", err)
	}

	return signedTx, nil
}

func (c *Client) WaitForConfirmation(txHash common.Hash, confirmations uint64) error {
	// 简化实现，实际项目中可能需要更复杂的等待逻辑
	receipt, err := bind.WaitMined(context.Background(), c.client, txHash)
	if err != nil {
		return err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("transaction failed")
	}

	// 如果需要更多确认
	if confirmations > 1 {
		// 这里可以添加额外的确认逻辑
		// 为了简化，我们只等待交易被包含在区块中
	}

	return nil
}