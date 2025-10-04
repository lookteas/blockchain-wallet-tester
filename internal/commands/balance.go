package commands

import (
	"fmt"
	"math/big"

	"transfer-tool/internal/config"
	"transfer-tool/internal/wallet"

	"github.com/urfave/cli/v2"
)

// BalanceCommand 余额查询命令
func BalanceCommand(c *cli.Context) error {
	network := c.String("network")
	appConfig := config.LoadAppConfig()
	envFile := appConfig.EnvFile

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

	// 获取所有地址
	addresses := wm.GetAddresses()
	if len(addresses) == 0 {
		return fmt.Errorf("没有可用的钱包地址")
	}

	// 查询余额
	fmt.Printf("Wallet Balances on %s (ChainID: %s):\n",
		wm.GetNetworkConfig().Name, wm.GetChainID().String())

	totalBalance := big.NewInt(0)
	hasZeroBalance := false

	for _, address := range addresses {
		balance, err := wm.GetBalance(address)
		if err != nil {
			fmt.Printf("- %s : 查询失败 (%v)\n", address.Hex(), err)
			continue
		}

		balanceStr := wallet.FormatAmount(balance)
		if balance.Cmp(big.NewInt(0)) == 0 {
			fmt.Printf("- %s : %s ETH ⚠️\n", address.Hex(), balanceStr)
			hasZeroBalance = true
		} else {
			fmt.Printf("- %s : %s ETH\n", address.Hex(), balanceStr)
		}

		totalBalance.Add(totalBalance, balance)
	}

	fmt.Printf("Total: %s ETH\n", wallet.FormatAmount(totalBalance))

	if hasZeroBalance {
		fmt.Printf("\n⚠️  部分钱包余额为0，可能影响转账操作\n")
	}

	return nil
}
