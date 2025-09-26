package transfer

import (
	"fmt"
	"math/big"
	"sync"

	"blockchain-wallet-tester/blockchain"
	"blockchain-wallet-tester/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type BatchTransfer struct {
	client         *blockchain.Client
	config         *config.Config
	transferAmount *big.Int
}

func NewBatchTransfer(client *blockchain.Client, cfg *config.Config) *BatchTransfer {
	amount := new(big.Int)
	amount.SetString(cfg.TransferAmount, 10)
	
	return &BatchTransfer{
		client:         client,
		config:         cfg,
		transferAmount: amount,
	}
}

func (bt *BatchTransfer) Execute(wallets []*Wallet, targets []common.Address) ([]*types.Transaction, error) {
	var transactions []*types.Transaction
	
	if bt.config.Concurrent {
		transactions = bt.executeConcurrent(wallets, targets)
	} else {
		transactions = bt.executeSequential(wallets, targets)
	}
	
	return transactions, nil
}

func (bt *BatchTransfer) executeSequential(wallets []*Wallet, targets []common.Address) []*types.Transaction {
	var transactions []*types.Transaction
	
	for i, wallet := range wallets {
		target := targets[i%len(targets)]
		fmt.Printf("Sending transaction from %s to %s\n", wallet.Address.Hex(), target.Hex())
		
		tx, err := bt.client.SendTransaction(wallet, target, bt.transferAmount)
		if err != nil {
			fmt.Printf("Failed to send transaction: %v\n", err)
			continue
		}
		
		fmt.Printf("Transaction hash: %s\n", tx.Hash().Hex())
		transactions = append(transactions, tx)
	}
	
	return transactions
}

func (bt *BatchTransfer) executeConcurrent(wallets []*Wallet, targets []common.Address) []*types.Transaction {
	var transactions []*types.Transaction
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	for i, wallet := range wallets {
		wg.Add(1)
		go func(w *Wallet, target common.Address, idx int) {
			defer wg.Done()
			
			fmt.Printf("Sending transaction from %s to %s\n", w.Address.Hex(), target.Hex())
			tx, err := bt.client.SendTransaction(w, target, bt.transferAmount)
			if err != nil {
				fmt.Printf("Failed to send transaction: %v\n", err)
				return
			}
			
			fmt.Printf("Transaction hash: %s\n", tx.Hash().Hex())
			mu.Lock()
			transactions = append(transactions, tx)
			mu.Unlock()
		}(wallet, targets[i%len(targets)], i)
	}
	
	wg.Wait()
	return transactions
}

func (bt *BatchTransfer) WaitForConfirmations(transactions []*types.Transaction) error {
	var wg sync.WaitGroup
	var errors []error
	var mu sync.Mutex
	
	for _, tx := range transactions {
		wg.Add(1)
		go func(transaction *types.Transaction) {
			defer wg.Done()
			
			err := bt.client.WaitForConfirmation(transaction.Hash(), bt.config.Confirmations)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("tx %s: %v", transaction.Hash().Hex(), err))
				mu.Unlock()
			}
		}(tx)
	}
	
	wg.Wait()
	
	if len(errors) > 0 {
		return fmt.Errorf("some transactions failed: %v", errors)
	}
	
	return nil
}