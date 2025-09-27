package cmd

import (
	"fmt"
	"os"

	"wallet-transfer/pkg/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	appConfig  *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wallet-transfer",
	Short: "区块链钱包转账工具",
	Long: `wallet-transfer 是一个功能强大的区块链钱包转账工具，支持：

- 多种区块链网络（Ethereum、BSC、Polygon等）
- 批量转账操作（一对一、一对多、多对一、多对多）
- 并发执行和速率控制
- 安全的私钥管理
- 余额查询和交易监控
- 灵活的配置选项`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 加载配置
		var err error
		appConfig, err = config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 验证配置
		if err := appConfig.ValidateConfig(); err != nil {
			return fmt.Errorf("配置验证失败: %w", err)
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

// GetConfig 获取应用配置
func GetConfig() *config.Config {
	return appConfig
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认为 ./config/config.yaml)")
	rootCmd.PersistentFlags().String("network", "sepolia", "区块链网络 (ethereum, goerli, sepolia, bsc, polygon, mumbai)")
	rootCmd.PersistentFlags().String("rpc-url", "", "自定义RPC URL (覆盖网络默认设置)")
	rootCmd.PersistentFlags().String("private-keys", "env", "私钥来源 (env, file, interactive)")
	rootCmd.PersistentFlags().Bool("concurrent", false, "启用并发执行")
	rootCmd.PersistentFlags().Int("workers", 10, "并发工作线程数")
	rootCmd.PersistentFlags().Int("confirmations", 1, "交易确认数")
	rootCmd.PersistentFlags().Int("timeout", 300, "操作超时时间（秒）")
	rootCmd.PersistentFlags().String("output", "table", "输出格式 (table, json, csv)")

	// 绑定标志到viper
	viper.BindPFlag("network", rootCmd.PersistentFlags().Lookup("network"))
	viper.BindPFlag("rpc-url", rootCmd.PersistentFlags().Lookup("rpc-url"))
	viper.BindPFlag("private-keys", rootCmd.PersistentFlags().Lookup("private-keys"))
	viper.BindPFlag("concurrent", rootCmd.PersistentFlags().Lookup("concurrent"))
	viper.BindPFlag("workers", rootCmd.PersistentFlags().Lookup("workers"))
	viper.BindPFlag("confirmations", rootCmd.PersistentFlags().Lookup("confirmations"))
	viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}

// initConfig reads in config file and ENV variables.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in current directory and config directory
		viper.AddConfigPath("./config")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "使用配置文件:", viper.ConfigFileUsed())
	}
}