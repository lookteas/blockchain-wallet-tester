package commands

import (
	"bufio"
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"transfer-tool/internal/config"
	"transfer-tool/internal/wallet"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v2"
)

// BatchCommand 批量转账命令
func BatchCommand(c *cli.Context) error {
	configFile := c.String("config")
	network := c.String("network")
	appConfig := config.LoadAppConfig()
	envFile := appConfig.EnvFile

	// 检查配置文件
	if configFile == "" {
		return fmt.Errorf("必须指定配置文件: --config <config_file>")
	}

	// 加载配置
	batchConfig, err := config.LoadBatchConfig(configFile)
	if err != nil {
		return fmt.Errorf("加载配置文件失败: %v", err)
	}

	// 创建钱包管理器（使用自定义RPC配置）
	wm, err := wallet.NewManagerWithRPC(envFile, network, batchConfig.RPCConfig)
	if err != nil {
		return err
	}
	defer wm.GetClient().Close()

	// 获取所有地址
	addresses := wm.GetAddresses()
	if len(addresses) == 0 {
		return fmt.Errorf("没有可用的钱包地址")
	}

	// 加载接收方数据
	recipients, err := config.LoadRecipients(batchConfig.DataSources.RecipientsXlsx)
	if err != nil {
		return fmt.Errorf("加载接收方数据失败: %v", err)
	}

	fmt.Printf("📋 批量转账信息:\n")
	fmt.Printf("   钱包数量: %d\n", len(addresses))
	fmt.Printf("   接收方数量: %d\n", len(recipients))
	fmt.Printf("   网络: %s\n", wm.GetNetworkConfig().Name)
	fmt.Printf("   配置文件: %s\n", configFile)

	// 确认执行
	fmt.Printf("\n确认执行批量转账? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm != "y" && confirm != "yes" {
		return fmt.Errorf("操作已取消")
	}

	// 执行批量转账
	report, err := executeBatchTransfer(wm, recipients, batchConfig)
	if err != nil {
		return fmt.Errorf("批量转账失败: %v", err)
	}

	// 生成报告
	reportFile := fmt.Sprintf("data/batch_report_%d.md", time.Now().Unix())
	err = config.SaveReport(report, reportFile)
	if err != nil {
		fmt.Printf("⚠️  报告保存失败: %v\n", err)
	} else {
		fmt.Printf("📊 报告已保存: %s\n", reportFile)
	}

	// 显示汇总
	fmt.Printf("\n📊 批量转账完成:\n")
	fmt.Printf("   成功: %d\n", report.Summary.Success)
	fmt.Printf("   失败: %d\n", report.Summary.Failed)
	fmt.Printf("   总计: %d\n", report.Summary.Total)

	if report.Summary.Failed > 0 {
		return fmt.Errorf("部分转账失败，请查看报告详情")
	}

	return nil
}

// executeBatchTransfer 执行批量转账
func executeBatchTransfer(wm *wallet.Manager, recipients []config.Recipient, batchConfig *config.BatchConfig) (*config.BatchReport, error) {
	report := &config.BatchReport{
		Timestamp: time.Now(),
		Network:   wm.GetNetworkConfig().Name,
		ChainID:   wm.GetChainID().String(),
		Summary:   &config.BatchSummary{},
		Details:   make([]*config.TransferDetail, 0, len(recipients)),
	}

	addresses := wm.GetAddresses()
	report.Summary.Total = len(recipients)

	for i, recipient := range recipients {
		// 轮询选择发送方
		senderIndex := i % len(addresses)
		fromAddress := addresses[senderIndex]
		toAddress := common.HexToAddress(recipient.Address)

		// 转换金额
		amount, err := wallet.ParseAmount(fmt.Sprintf("%.6f", recipient.Amount))
		if err != nil {
			report.AddFailedDetail(i, recipient, fmt.Sprintf("金额解析失败: %v", err))
			continue
		}

		// 检查余额
		balance, err := wm.GetBalance(fromAddress)
		if err != nil {
			report.AddFailedDetail(i, recipient, fmt.Sprintf("查询余额失败: %v", err))
			continue
		}

		// 估算Gas费用
		gasPrice, err := wm.GetGasPrice()
		if err != nil {
			report.AddFailedDetail(i, recipient, fmt.Sprintf("获取Gas价格失败: %v", err))
			continue
		}

		gasLimit, err := wm.EstimateGas(fromAddress, toAddress, amount, nil)
		if err != nil {
			report.AddFailedDetail(i, recipient, fmt.Sprintf("估算Gas失败: %v", err))
			continue
		}

		totalCost := new(big.Int).Add(amount, new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit))))

		if balance.Cmp(totalCost) < 0 {
			report.AddFailedDetail(i, recipient, fmt.Sprintf("余额不足: 需要 %s ETH，当前 %s ETH",
				wallet.FormatAmount(totalCost), wallet.FormatAmount(balance)))
			continue
		}

		// 执行转账
		txHash, err := executeSingleTransfer(wm, fromAddress, toAddress, amount, gasPrice, gasLimit)
		if err != nil {
			report.AddFailedDetail(i, recipient, fmt.Sprintf("转账失败: %v", err))
			continue
		}

		// 记录成功
		report.AddSuccessDetail(i, recipient, fromAddress.Hex(), txHash, wm.GetExplorerURL(txHash))
	}

	return report, nil
}

// executeSingleTransfer 执行单笔转账
func executeSingleTransfer(wm *wallet.Manager, from, to common.Address, value, gasPrice *big.Int, gasLimit uint64) (string, error) {
	ctx := context.Background()

	// 获取私钥
	privateKey := wm.GetPrivateKeyByIndex(0)

	// 获取nonce
	nonce, err := wm.GetClient().PendingNonceAt(ctx, from)
	if err != nil {
		return "", fmt.Errorf("获取nonce失败: %v", err)
	}

	// 构建交易
	tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, nil)

	// 签名交易
	chainID := wm.GetChainID()
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %v", err)
	}

	// 发送交易
	err = wm.GetClient().SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("发送交易失败: %v", err)
	}

	return signedTx.Hash().Hex(), nil
}
