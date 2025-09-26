package main

import (
	"fmt"
	"log"
	"os"

	"blockchain-wallet-tester/blockchain"
	"blockchain-wallet-tester/config"
	"blockchain-wallet-tester/transfer"
	"blockchain-wallet-tester/wallet"

	"github.com/spf13/cobra"
)

var (
	configFile      string
	interactiveMode bool
	balanceOnly     bool
	concurrent      bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "wallet-tester",
		Short: "Blockchain wallet transfer tester",
		Long:  "A tool for testing multiple wallet transfers on blockchain networks",
		Run: func(cmd *cobra.Command, args []string) {
			runApplication()
		},
	}

	rootCmd.Flags().StringVar(&configFile, "config", "", "config file path")
	rootCmd.Flags().BoolVar(&interactiveMode, "interactive", false, "interactive mode for private key input")
	rootCmd.Flags().BoolVar(&balanceOnly, "balance-only", false, "only show wallet balances without transferring")
	rootCmd.Flags().BoolVar(&concurrent, "concurrent", false, "enable concurrent transfers")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runApplication() {
	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 如果指定了命令行参数，覆盖配置
	if concurrent {
		cfg.Concurrent = true
	}

	// 加载钱包
	var wallets []*wallet.Wallet
	if interactiveMode {
		wallets, err = wallet.LoadWalletsInteractive()
	} else {
		wallets, err = wallet.LoadWallets(cfg)
	}
	if err != nil {
		log.Fatalf("Failed to load wallets: %v", err)
	}

	if len(wallets) == 0 {
		log.Fatal("No wallets loaded")
	}

	// 创建区块链客户端
	client, err := blockchain.NewClient(cfg.RPCURL)
	if err != nil {
		log.Fatalf("Failed to create blockchain client: %v", err)
	}

	// 显示初始余额
	fmt.Println("Initial wallet balances:")
	for _, w := range wallets {
		balance, err := client.GetBalance(w.Address)
		if err != nil {
			log.Printf("Failed to get balance for %s: %v", w.Address.Hex(), err)
			continue
		}
		fmt.Printf("Address %s: %s wei\n", w.Address.Hex(), balance.String())
	}

	// 如果只显示余额，退出
	if balanceOnly {
		return
	}

	// 执行批量转账
	fmt.Printf("\nStarting batch transfer with %d wallets to %d addresses\n", 
		len(wallets), len(cfg.TargetAddresses))

	batchTransfer := transfer.NewBatchTransfer(client, cfg)
	transactions, err := batchTransfer.Execute(wallets, cfg.TargetAddresses)
	if err != nil {
		log.Fatalf("Batch transfer failed: %v", err)
	}

	fmt.Printf("Successfully sent %d transactions\n", len(transactions))

	// 等待确认
	if cfg.WaitConfirmations {
		fmt.Println("\nWaiting for transaction confirmations...")
		err = batchTransfer.WaitForConfirmations(transactions)
		if err != nil {
			log.Printf("Some transactions may have failed: %v", err)
		}
	}

	// 显示最终余额
	fmt.Println("\nFinal wallet balances:")
	for _, w := range wallets {
		balance, err := client.GetBalance(w.Address)
		if err != nil {
			log.Printf("Failed to get balance for %s: %v", w.Address.Hex(), err)
			continue
		}
		fmt.Printf("Address %s: %s wei\n", w.Address.Hex(), balance.String())
	}
}