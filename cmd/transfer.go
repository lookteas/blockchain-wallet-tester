package cmd

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"wallet-transfer/pkg/blockchain"
	"wallet-transfer/pkg/crypto"
	"wallet-transfer/pkg/transfer"
	"wallet-transfer/pkg/wallet"

	"github.com/ethereum/go-ethereum/common"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// transferCmd represents the transfer command
var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "执行批量转账测试",
	Long: `执行批量转账测试，支持多种转账模式：
- one-to-one: 一对一转账
- one-to-many: 一对多转账  
- many-to-one: 多对一转账
- many-to-many: 多对多转账`,
	RunE: runTransfer,
}

func init() {
	rootCmd.AddCommand(transferCmd)

	// Transfer specific flags
	transferCmd.Flags().String("mode", "one-to-one", "转账模式 (one-to-one, one-to-many, many-to-one, many-to-many)")
	transferCmd.Flags().String("recipients", "", "收款地址列表，逗号分隔")
	transferCmd.Flags().String("amount", "", "转账金额 (固定金额)")
	transferCmd.Flags().String("amount-range", "", "转账金额范围 (格式: min-max)")
	transferCmd.Flags().String("unit", "ether", "金额单位 (wei, gwei, ether)")
	transferCmd.Flags().Uint64("gas-limit", 21000, "Gas限制")
	transferCmd.Flags().String("gas-price", "", "Gas价格 (wei)")
	transferCmd.Flags().Bool("auto-gas", true, "自动估算Gas")
	transferCmd.Flags().String("data", "", "交易数据")

	// Bind flags to viper
	viper.BindPFlag("mode", transferCmd.Flags().Lookup("mode"))
	viper.BindPFlag("recipients", transferCmd.Flags().Lookup("recipients"))
	viper.BindPFlag("amount", transferCmd.Flags().Lookup("amount"))
	viper.BindPFlag("amount-range", transferCmd.Flags().Lookup("amount-range"))
	viper.BindPFlag("unit", transferCmd.Flags().Lookup("unit"))
	viper.BindPFlag("gas-limit", transferCmd.Flags().Lookup("gas-limit"))
	viper.BindPFlag("gas-price", transferCmd.Flags().Lookup("gas-price"))
	viper.BindPFlag("auto-gas", transferCmd.Flags().Lookup("auto-gas"))
	viper.BindPFlag("data", transferCmd.Flags().Lookup("data"))
}

func runTransfer(cmd *cobra.Command, args []string) error {
	// Initialize key manager
	keyManager := crypto.NewKeyManager()
	
	// Load private keys based on configuration
	privateKeysSource := viper.GetString("private-keys")
	switch privateKeysSource {
	case "env":
		if err := keyManager.LoadFromEnv(); err != nil {
			return fmt.Errorf("failed to load private keys from environment: %w", err)
		}
	case "interactive":
		if err := keyManager.LoadInteractive(); err != nil {
			return fmt.Errorf("failed to load private keys interactively: %w", err)
		}
	case "file":
		// TODO: Implement file loading
		return fmt.Errorf("file loading not yet implemented")
	default:
		return fmt.Errorf("unsupported private keys source: %s", privateKeysSource)
	}

	// Ensure we clear private keys from memory when done
	defer keyManager.Clear()

	// Initialize wallet manager
	walletManager := wallet.NewWalletManager()
	if err := walletManager.LoadWallets(keyManager.GetPrivateKeys()); err != nil {
		return fmt.Errorf("failed to load wallets: %w", err)
	}

	// Initialize network manager
	networkName := viper.GetString("network")
	rpcURL := viper.GetString("rpc-url")
	networkManager, err := blockchain.NewNetworkManager(networkName, rpcURL)
	if err != nil {
		return fmt.Errorf("failed to initialize network manager: %w", err)
	}
	defer networkManager.Close()

	// Parse transfer configuration
	transferConfig, err := parseTransferConfig()
	if err != nil {
		return fmt.Errorf("failed to parse transfer configuration: %w", err)
	}

	// Initialize transfer manager
	transferManager := transfer.NewTransferManager(networkManager, walletManager, &transferConfig)

	// Execute transfers
	fmt.Printf("开始执行转账测试...\n")
	fmt.Printf("网络: %s (Chain ID: %d)\n", networkManager.GetConfig().Name, networkManager.GetConfig().ChainID)
	fmt.Printf("转账模式: %s\n", transferConfig.Mode)
	fmt.Printf("钱包数量: %d\n", walletManager.GetWalletCount())
	fmt.Printf("收款地址数量: %d\n", len(transferConfig.Recipients))
	fmt.Println()

	result, err := transferManager.ExecuteTransfers()
	if err != nil {
		return fmt.Errorf("transfer execution failed: %w", err)
	}

	// Display results
	if err := displayResults(result); err != nil {
		return fmt.Errorf("failed to display results: %w", err)
	}

	return nil
}

func parseTransferConfig() (transfer.TransferConfig, error) {
	config := transfer.TransferConfig{
		Mode:          viper.GetString("mode"),
		Concurrent:    viper.GetBool("concurrent"),
		Workers:       viper.GetInt("workers"),
		Confirmations: viper.GetInt("confirmations"),
		Timeout:       time.Duration(viper.GetInt("timeout")) * time.Second,
		MaxRetries:    viper.GetInt("retries"),
		RetryDelay:    time.Duration(viper.GetInt("retry-delay")) * time.Second,
		RateLimit:     viper.GetFloat64("rate-limit"),
	}

	// Parse recipients
	recipientsStr := viper.GetString("recipients")
	if recipientsStr == "" {
		return config, fmt.Errorf("recipients are required")
	}
	recipientStrs := strings.Split(recipientsStr, ",")
	config.Recipients = make([]common.Address, len(recipientStrs))
	for i, recipient := range recipientStrs {
		config.Recipients[i] = common.HexToAddress(strings.TrimSpace(recipient))
	}

	// Parse amount configuration
	amount := viper.GetString("amount")
	amountRange := viper.GetString("amount-range")
	
	config.AmountConfig = &transfer.AmountConfig{}
	
	if amount != "" {
		// Parse fixed amount
		amountBig, ok := new(big.Int).SetString(amount, 10)
		if !ok {
			return config, fmt.Errorf("invalid amount format: %s", amount)
		}
		config.AmountConfig.Fixed = amountBig
	} else if amountRange != "" {
		// Parse amount range
		parts := strings.Split(amountRange, "-")
		if len(parts) != 2 {
			return config, fmt.Errorf("invalid amount range format, expected 'min-max'")
		}
		minAmount, ok1 := new(big.Int).SetString(strings.TrimSpace(parts[0]), 10)
		maxAmount, ok2 := new(big.Int).SetString(strings.TrimSpace(parts[1]), 10)
		if !ok1 || !ok2 {
			return config, fmt.Errorf("invalid amount range values")
		}
		config.AmountConfig.MinRange = minAmount
		config.AmountConfig.MaxRange = maxAmount
	} else {
		return config, fmt.Errorf("either amount or amount-range must be specified")
	}

	// Parse gas configuration
	config.AutoGas = viper.GetBool("auto-gas")
	config.GasLimit = viper.GetUint64("gas-limit")
	
	gasPriceStr := viper.GetString("gas-price")
	if gasPriceStr != "" {
		gasPrice, ok := new(big.Int).SetString(gasPriceStr, 10)
		if !ok {
			return config, fmt.Errorf("invalid gas price format: %s", gasPriceStr)
		}
		config.GasPrice = gasPrice
	}

	// Parse data
	dataStr := viper.GetString("data")
	if dataStr != "" {
		config.Data = []byte(dataStr)
	}

	return config, nil
}

func displayResults(result *transfer.TransferResult) error {
	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		return displayResultsJSON(result)
	case "table":
		return displayResultsTable(result)
	case "csv":
		return displayResultsCSV(result)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func displayResultsJSON(result *transfer.TransferResult) error {
	// Convert big.Int to string for JSON serialization
	jsonResult := struct {
		TotalTasks  int                     `json:"total_tasks"`
		Successful  int                     `json:"successful"`
		Failed      int                     `json:"failed"`
		TotalAmount string                  `json:"total_amount"`
		TotalFees   string                  `json:"total_fees"`
		Duration    string                  `json:"duration"`
		Tasks       []*transfer.TransferTask `json:"tasks"`
	}{
		TotalTasks:  result.TotalTasks,
		Successful:  result.Successful,
		Failed:      result.Failed,
		TotalAmount: result.TotalAmount.String(),
		TotalFees:   result.TotalFees.String(),
		Duration:    result.Duration.String(),
		Tasks:       result.Tasks,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonResult)
}

func displayResultsTable(result *transfer.TransferResult) error {
	// Summary table
	fmt.Println("=== 转账结果摘要 ===")
	summaryTable := tablewriter.NewWriter(os.Stdout)
	summaryTable.SetHeader([]string{"指标", "值"})
	summaryTable.Append([]string{"总任务数", strconv.Itoa(result.TotalTasks)})
	summaryTable.Append([]string{"成功", strconv.Itoa(result.Successful)})
	summaryTable.Append([]string{"失败", strconv.Itoa(result.Failed)})
	summaryTable.Append([]string{"总转账金额", result.TotalAmount.String() + " wei"})
	summaryTable.Append([]string{"总手续费", result.TotalFees.String() + " wei"})
	summaryTable.Append([]string{"执行时间", result.Duration.String()})
	summaryTable.Render()

	fmt.Println()

	// Detailed tasks table
	fmt.Println("=== 详细转账记录 ===")
	tasksTable := tablewriter.NewWriter(os.Stdout)
	tasksTable.SetHeader([]string{"任务ID", "发送地址", "接收地址", "金额(wei)", "状态", "交易哈希", "错误"})

	for _, task := range result.Tasks {
		errorStr := ""
		if task.Error != "" {
			errorStr = task.Error
			if len(errorStr) > 50 {
				errorStr = errorStr[:47] + "..."
			}
		}

		txHashStr := ""
		if task.TxHash != "" {
			txHashStr = task.TxHash[:10] + "..."
		}

		tasksTable.Append([]string{
			task.ID,
			task.From[:10] + "...",
			task.To[:10] + "...",
			task.Amount.String(),
			string(task.Status),
			txHashStr,
			errorStr,
		})
	}

	tasksTable.Render()
	return nil
}

func displayResultsCSV(result *transfer.TransferResult) error {
	fmt.Println("TaskID,From,To,Amount,Status,TxHash,Error,Duration")
	
	for _, task := range result.Tasks {
		errorStr := ""
		if task.Error != "" {
			errorStr = strings.ReplaceAll(task.Error, ",", ";")
		}

		duration := ""
		if !task.EndTime.IsZero() {
			duration = task.EndTime.Sub(task.StartTime).String()
		}

		fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s\n",
			task.ID,
			task.From,
			task.To,
			task.Amount.String(),
			task.Status,
			task.TxHash,
			errorStr,
			duration,
		)
	}

	return nil
}