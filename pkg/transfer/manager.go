package transfer

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"wallet-transfer/pkg/blockchain"
	"wallet-transfer/pkg/concurrency"
	"wallet-transfer/pkg/wallet"
)

// ConcurrentTransferManager 并发转账管理器
type ConcurrentTransferManager struct {
	networkManager *blockchain.NetworkManager
	walletManager  *wallet.WalletManager
	config         *TransferConfig
	workerPool     *concurrency.WorkerPool
	rateLimiter    *concurrency.RateLimiter
	retryExecutor  *concurrency.RetryExecutor
	circuitBreaker *concurrency.CircuitBreaker
	stats          *TransferStats
	mutex          sync.RWMutex
}

// NewConcurrentTransferManager 创建新的并发转账管理器
func NewConcurrentTransferManager(networkManager *blockchain.NetworkManager, walletManager *wallet.WalletManager, config *TransferConfig) *ConcurrentTransferManager {
	// 创建工作池
	workerPool := concurrency.NewWorkerPool(config.Workers, config.Workers*2)

	// 创建速率限制器
	rateLimiter := concurrency.NewRateLimiter(int(config.RateLimit))

	// 创建重试配置
	retryConfig := &concurrency.RetryConfig{
		MaxRetries:    config.MaxRetries,
		BaseDelay:     time.Second,
		MaxDelay:      time.Minute * 5,
		BackoffFactor: 2.0,
	}

	retryExecutor := concurrency.NewRetryExecutor(retryConfig)

	// 创建熔断器
	circuitBreaker := concurrency.NewCircuitBreaker(10, time.Minute*5)

	return &ConcurrentTransferManager{
		networkManager: networkManager,
		walletManager:  walletManager,
		config:         config,
		workerPool:     workerPool,
		rateLimiter:    rateLimiter,
		retryExecutor:  retryExecutor,
		circuitBreaker: circuitBreaker,
		stats: &TransferStats{
			StartTime:    time.Now(),
			TotalAmount:  big.NewInt(0),
			TotalGasCost: big.NewInt(0),
		},
	}
}

// TransferTask 实现concurrency.Task接口
type TransferTaskImpl struct {
	task *TransferTask
	manager *ConcurrentTransferManager
}

func (t *TransferTaskImpl) Execute(ctx context.Context) error {
	return t.manager.executeTransferTask(ctx, t.task)
}

func (t *TransferTaskImpl) GetID() string {
	return t.task.ID
}

// ExecuteTransfers 执行转账任务
func (ctm *ConcurrentTransferManager) ExecuteTransfers() (*TransferResult, error) {
	ctx := context.Background()
	startTime := time.Now()

	// 生成转账任务
	tasks, err := ctm.generateTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to generate tasks: %w", err)
	}

	// 启动工作池
	ctm.workerPool.Start()
	defer ctm.workerPool.Stop()

	// 提交任务
	for _, task := range tasks {
		taskImpl := &TransferTaskImpl{
			task: task,
			manager: ctm,
		}
		
		if err := ctm.workerPool.SubmitTask(taskImpl); err != nil {
			return nil, fmt.Errorf("failed to submit task %s: %w", task.ID, err)
		}
	}

	// 收集结果
	results := make([]*TransferTask, 0, len(tasks))
	resultChan := ctm.workerPool.GetResults()
	
	for i := 0; i < len(tasks); i++ {
		select {
		case result := <-resultChan:
			// 找到对应的任务并更新状态
			for _, task := range tasks {
				if task.ID == result.TaskID {
					if result.Success {
						task.Status = "completed"
					} else {
						task.Status = "failed"
						if result.Error != nil {
							task.Error = result.Error.Error()
						}
					}
					task.EndTime = result.EndTime
					results = append(results, task)
					break
				}
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// 计算统计信息
	return ctm.calculateResults(results, startTime), nil
}

// generateTasks 生成转账任务
func (ctm *ConcurrentTransferManager) generateTasks() ([]*TransferTask, error) {
	var tasks []*TransferTask
	
	wallets := ctm.walletManager.GetWallets()
	if len(wallets) == 0 {
		return nil, fmt.Errorf("no wallets available")
	}

	switch ctm.config.Mode {
	case "one-to-one":
		return ctm.generateOneToOneTasks(wallets)
	case "one-to-many":
		return ctm.generateOneToManyTasks(wallets)
	case "many-to-one":
		return ctm.generateManyToOneTasks(wallets)
	case "many-to-many":
		return ctm.generateManyToManyTasks(wallets)
	default:
		return nil, fmt.Errorf("unsupported transfer mode: %s", ctm.config.Mode)
	}

	return tasks, nil
}

// generateOneToOneTasks 生成一对一转账任务
func (ctm *ConcurrentTransferManager) generateOneToOneTasks(wallets []*wallet.Wallet) ([]*TransferTask, error) {
	var tasks []*TransferTask
	
	for i := 0; i < len(wallets)-1; i++ {
		fromWallet := wallets[i]
		toWallet := wallets[i+1]
		
		task := &TransferTask{
			ID:       fmt.Sprintf("one-to-one-%d", i),
			From:     fromWallet.GetAddress().Hex(),
			To:       toWallet.GetAddress().Hex(),
			Amount:   ctm.generateAmount(),
			GasPrice: ctm.config.GasPrice,
			GasLimit: ctm.config.GasLimit,
			Data:     ctm.config.Data,
			Status:   "pending",
		}
		
		tasks = append(tasks, task)
	}
	
	return tasks, nil
}

// generateOneToManyTasks 生成一对多转账任务
func (ctm *ConcurrentTransferManager) generateOneToManyTasks(wallets []*wallet.Wallet) ([]*TransferTask, error) {
	if len(wallets) == 0 {
		return nil, fmt.Errorf("no wallets available")
	}
	
	var tasks []*TransferTask
	fromWallet := wallets[0] // 使用第一个钱包作为发送方
	
	for i := 1; i < len(wallets); i++ {
		toWallet := wallets[i]
		
		task := &TransferTask{
			ID:       fmt.Sprintf("one-to-many-%d", i),
			From:     fromWallet.GetAddress().Hex(),
			To:       toWallet.GetAddress().Hex(),
			Amount:   ctm.generateAmount(),
			GasPrice: ctm.config.GasPrice,
			GasLimit: ctm.config.GasLimit,
			Data:     ctm.config.Data,
			Status:   "pending",
		}
		
		tasks = append(tasks, task)
	}
	
	return tasks, nil
}

// generateManyToOneTasks 生成多对一转账任务
func (ctm *ConcurrentTransferManager) generateManyToOneTasks(wallets []*wallet.Wallet) ([]*TransferTask, error) {
	if len(wallets) < 2 {
		return nil, fmt.Errorf("need at least 2 wallets for many-to-one transfers")
	}
	
	var tasks []*TransferTask
	toWallet := wallets[len(wallets)-1] // 使用最后一个钱包作为接收方
	
	for i := 0; i < len(wallets)-1; i++ {
		fromWallet := wallets[i]
		
		task := &TransferTask{
			ID:       fmt.Sprintf("many-to-one-%d", i),
			From:     fromWallet.GetAddress().Hex(),
			To:       toWallet.GetAddress().Hex(),
			Amount:   ctm.generateAmount(),
			GasPrice: ctm.config.GasPrice,
			GasLimit: ctm.config.GasLimit,
			Data:     ctm.config.Data,
			Status:   "pending",
		}
		
		tasks = append(tasks, task)
	}
	
	return tasks, nil
}

// generateManyToManyTasks 生成多对多转账任务
func (ctm *ConcurrentTransferManager) generateManyToManyTasks(wallets []*wallet.Wallet) ([]*TransferTask, error) {
	if len(wallets) < 2 {
		return nil, fmt.Errorf("need at least 2 wallets for many-to-many transfers")
	}
	
	var tasks []*TransferTask
	
	for i := 0; i < len(wallets); i++ {
		fromWallet := wallets[i]
		toWallet := wallets[(i+1)%len(wallets)] // 循环选择接收方
		
		task := &TransferTask{
			ID:       fmt.Sprintf("many-to-many-%d", i),
			From:     fromWallet.GetAddress().Hex(),
			To:       toWallet.GetAddress().Hex(),
			Amount:   ctm.generateAmount(),
			GasPrice: ctm.config.GasPrice,
			GasLimit: ctm.config.GasLimit,
			Data:     ctm.config.Data,
			Status:   "pending",
		}
		
		tasks = append(tasks, task)
	}
	
	return tasks, nil
}

// generateAmount 生成转账金额
func (ctm *ConcurrentTransferManager) generateAmount() *big.Int {
	if ctm.config.AmountConfig.Fixed != nil {
		return new(big.Int).Set(ctm.config.AmountConfig.Fixed)
	}
	
	if ctm.config.AmountConfig.MinRange != nil && ctm.config.AmountConfig.MaxRange != nil {
		min := ctm.config.AmountConfig.MinRange
		max := ctm.config.AmountConfig.MaxRange
		
		diff := new(big.Int).Sub(max, min)
		if diff.Sign() <= 0 {
			return new(big.Int).Set(min)
		}
		
		// 生成随机数
		randomValue := rand.Int63n(diff.Int64())
		randomBig := big.NewInt(randomValue)
		
		return new(big.Int).Add(min, randomBig)
	}
	
	// 默认值
	return big.NewInt(1000000000000000000) // 1 ETH in wei
}

// executeTransferTask 执行单个转账任务
func (ctm *ConcurrentTransferManager) executeTransferTask(ctx context.Context, task *TransferTask) error {
	// 速率限制
	if err := ctm.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	// 使用熔断器执行转账
	return ctm.circuitBreaker.Execute(func() error {
		return ctm.retryExecutor.ExecuteWithContext(ctx, func() error {
			return ctm.performTransfer(ctx, task)
		})
	})
}

// performTransfer 执行实际的转账操作
func (ctm *ConcurrentTransferManager) performTransfer(ctx context.Context, task *TransferTask) error {
	// 获取发送方钱包
	fromAddr := common.HexToAddress(task.From)
	privateKey, err := ctm.walletManager.GetPrivateKey(fromAddr)
	if err != nil {
		return fmt.Errorf("failed to get private key: %w", err)
	}

	// 获取nonce
	nonce, err := ctm.networkManager.GetNonce(ctx, fromAddr)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}

	// 创建交易
	toAddr := common.HexToAddress(task.To)
	tx := types.NewTransaction(
		nonce,
		toAddr,
		task.Amount,
		task.GasLimit,
		task.GasPrice,
		task.Data,
	)

	// 签名交易
	chainID := ctm.networkManager.GetChainID()
	signedTx, err := privateKey.SignTransaction(tx, chainID)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 发送交易
	txHash, err := ctm.networkManager.SendTransaction(ctx, signedTx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	task.TxHash = txHash.Hex()

	// 等待确认
	if ctm.config.Confirmations > 0 {
		_, err := ctm.networkManager.WaitForTransaction(ctx, txHash, ctm.config.Confirmations)
		if err != nil {
			return fmt.Errorf("failed to wait for confirmation: %w", err)
		}
	}

	return nil
}

// GetStats 获取转账统计信息
func (ctm *ConcurrentTransferManager) GetStats() TransferStats {
	ctm.mutex.RLock()
	defer ctm.mutex.RUnlock()
	
	return *ctm.stats
}

// calculateResults 计算转账结果
func (ctm *ConcurrentTransferManager) calculateResults(tasks []*TransferTask, startTime time.Time) *TransferResult {
	result := &TransferResult{
		TotalTasks:  len(tasks),
		Successful:  0,
		Failed:      0,
		TotalAmount: big.NewInt(0),
		TotalFees:   big.NewInt(0),
		Duration:    time.Since(startTime),
		Tasks:       tasks,
	}

	for _, task := range tasks {
		if task.Status == "completed" {
			result.Successful++
			result.TotalAmount.Add(result.TotalAmount, task.Amount)
			
			// 计算手续费
			fee := new(big.Int).Mul(task.GasPrice, big.NewInt(int64(task.GasLimit)))
			result.TotalFees.Add(result.TotalFees, fee)
		} else {
			result.Failed++
		}
	}

	return result
}