package commands

import (
	"bufio"
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"

	"transfer-tool/internal/config"
	"transfer-tool/internal/wallet"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v2"
)

// SendCommand 单笔转账命令
func SendCommand(c *cli.Context) error {
	// 检查参数
	if c.NArg() != 2 {
		return fmt.Errorf("用法: transfer-tool send <recipient_address> <amount_in_eth>")
	}

	recipientStr := c.Args().Get(0)
	amountStr := c.Args().Get(1)
	network := c.String("network")
	appConfig := config.LoadAppConfig()
	envFile := appConfig.EnvFile
	skipConfirm := c.Bool("yes")

	// 验证接收地址
	if err := wallet.ValidateAddress(recipientStr); err != nil {
		return err
	}

	// 解析金额
	amount, err := wallet.ParseAmount(amountStr)
	if err != nil {
		return err
	}

	// 尝试加载全局RPC配置
	var customRPCs map[string]string
	if globalRPC, err := config.LoadGlobalRPCConfig(); err == nil {
		customRPCs = globalRPC
	}

	// 创建钱包管理器
	wm, err := wallet.NewManagerWithRPC(envFile, network, customRPCs)
	if err != nil {
		return err
	}
	defer wm.GetClient().Close()

	// 获取发送方地址（第一个钱包）
	fromAddress := wm.GetFirstAddress()
	if fromAddress == (common.Address{}) {
		return fmt.Errorf("没有可用的钱包地址")
	}

	// 获取接收方地址
	toAddress := common.HexToAddress(recipientStr)

	// 检查余额
	balance, err := wm.GetBalance(fromAddress)
	if err != nil {
		return fmt.Errorf("查询余额失败: %v", err)
	}

	// 估算Gas费用
	gasPrice, err := wm.GetGasPrice()
	if err != nil {
		return fmt.Errorf("获取Gas价格失败: %v", err)
	}

	gasLimit, err := wm.EstimateGas(fromAddress, toAddress, amount, nil)
	if err != nil {
		return fmt.Errorf("估算Gas失败: %v", err)
	}

	totalCost := new(big.Int).Add(amount, new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit))))

	if balance.Cmp(totalCost) < 0 {
		return fmt.Errorf("余额不足: 需要 %s ETH，当前余额 %s ETH",
			wallet.FormatAmount(totalCost), wallet.FormatAmount(balance))
	}

	// 显示交易信息
	fmt.Printf("📤 转账信息:\n")
	fmt.Printf("   发送方: %s\n", fromAddress.Hex())
	fmt.Printf("   接收方: %s\n", toAddress.Hex())
	fmt.Printf("   金额: %s ETH\n", wallet.FormatAmount(amount))
	fmt.Printf("   Gas价格: %s Gwei\n", wallet.FormatAmount(new(big.Int).Div(gasPrice, big.NewInt(1e9))))
	fmt.Printf("   Gas限制: %d\n", gasLimit)
	fmt.Printf("   网络: %s\n", wm.GetNetworkConfig().Name)

	// 主网额外确认
	if network == "mainnet" {
		fmt.Printf("\n⚠️  警告: 您正在主网执行转账操作！\n")
		if !skipConfirm {
			fmt.Printf("请输入 'MAINNET' 确认: ")
			reader := bufio.NewReader(os.Stdin)
			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(confirm)
			if confirm != "MAINNET" {
				return fmt.Errorf("操作已取消")
			}
		}
	} else if !skipConfirm {
		// 其他网络确认
		fmt.Printf("\n确认执行转账? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))
		if confirm != "y" && confirm != "yes" {
			return fmt.Errorf("操作已取消")
		}
	}

	// 执行转账
	txHash, err := executeTransfer(wm, fromAddress, toAddress, amount, gasPrice, gasLimit)
	if err != nil {
		return fmt.Errorf("转账失败: %v", err)
	}

	fmt.Printf("\n✅ 转账成功!\n")
	fmt.Printf("   交易哈希: %s\n", txHash)
	fmt.Printf("   区块浏览器: %s\n", wm.GetExplorerURL(txHash))

	return nil
}

// executeTransfer 执行转账
func executeTransfer(wm *wallet.Manager, from, to common.Address, value, gasPrice *big.Int, gasLimit uint64) (string, error) {
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
