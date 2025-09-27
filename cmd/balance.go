package cmd

import (
	"fmt"

	"wallet-transfer/pkg/blockchain"
	"wallet-transfer/pkg/crypto"
	"wallet-transfer/pkg/utils"
	"wallet-transfer/pkg/wallet"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Query wallet balances",
	Long:  `Query balances for wallets from private keys or specified addresses`,
	RunE:  runBalance,
}

func init() {
	rootCmd.AddCommand(balanceCmd)

	balanceCmd.Flags().StringSliceP("private-keys", "k", nil, "Private keys (comma separated)")
	balanceCmd.Flags().StringSliceP("addresses", "a", nil, "Addresses to query (comma separated)")
	balanceCmd.Flags().StringP("network", "n", "", "Network name (ethereum, goerli, sepolia, bsc, polygon, mumbai)")
	balanceCmd.Flags().StringP("rpc-url", "r", "", "Custom RPC URL")
	balanceCmd.Flags().StringP("unit", "u", "ether", "Display unit (wei, gwei, ether)")
	balanceCmd.Flags().StringP("output", "o", "table", "Output format (json, table, csv)")

	viper.BindPFlag("balance.private-keys", balanceCmd.Flags().Lookup("private-keys"))
	viper.BindPFlag("balance.addresses", balanceCmd.Flags().Lookup("addresses"))
	viper.BindPFlag("balance.network", balanceCmd.Flags().Lookup("network"))
	viper.BindPFlag("balance.rpc-url", balanceCmd.Flags().Lookup("rpc-url"))
	viper.BindPFlag("balance.unit", balanceCmd.Flags().Lookup("unit"))
	viper.BindPFlag("balance.output", balanceCmd.Flags().Lookup("output"))
}

func runBalance(cmd *cobra.Command, args []string) error {
	config := GetConfig()

	// Get command line parameters
	privateKeys := viper.GetStringSlice("balance.private-keys")
	addresses := viper.GetStringSlice("balance.addresses")
	networkName := viper.GetString("balance.network")
	rpcURL := viper.GetString("balance.rpc-url")
	unit := viper.GetString("balance.unit")
	outputFormat := viper.GetString("balance.output")

	// Use default network if not specified
	if networkName == "" {
		networkName = config.Defaults.Network
	}

	// Get network configuration
	networkConfig, err := config.GetNetworkConfig(networkName)
	if err != nil {
		return fmt.Errorf("failed to get network config: %w", err)
	}

	// Use custom RPC URL if provided
	if rpcURL != "" {
		networkConfig.RPCURL = rpcURL
	}

	// Initialize network manager
	networkManager, err := blockchain.NewNetworkManager(networkConfig.Name, networkConfig.RPCURL)
	if err != nil {
		return fmt.Errorf("failed to initialize network manager: %w", err)
	}
	defer networkManager.Close()

	// Initialize key manager
	keyManager := crypto.NewKeyManager()

	// Load private keys
	if len(privateKeys) > 0 {
		// Use provided private keys
		if err := keyManager.LoadFromEnv(); err != nil {
			return fmt.Errorf("failed to load private keys from environment: %w", err)
		}
	} else {
		// Interactive input
		if err := keyManager.LoadInteractive(); err != nil {
			return fmt.Errorf("failed to load private keys interactively: %w", err)
		}
	}

	// Initialize wallet manager
	walletManager := wallet.NewWalletManager()
	if err := walletManager.LoadWallets(keyManager.GetPrivateKeys()); err != nil {
		return fmt.Errorf("failed to load wallets: %w", err)
	}

	// Collect all addresses to query
	var queryAddresses []common.Address

	// Add wallet addresses
	for _, wallet := range walletManager.GetWallets() {
		queryAddresses = append(queryAddresses, wallet.GetAddress())
	}

	// Add specified addresses
	for _, addrStr := range addresses {
		if !common.IsHexAddress(addrStr) {
			return fmt.Errorf("invalid address: %s", addrStr)
		}
		queryAddresses = append(queryAddresses, common.HexToAddress(addrStr))
	}

	if len(queryAddresses) == 0 {
		return fmt.Errorf("no addresses to query")
	}

	// Query balances
	var balanceData []utils.BalanceData
	for _, addr := range queryAddresses {
		balance, err := networkManager.GetBalance(cmd.Context(), addr)
		if err != nil {
			return fmt.Errorf("failed to get balance for %s: %w", addr.Hex(), err)
		}
		balanceData = append(balanceData, utils.BalanceData{
			Address: addr.Hex(),
			Balance: utils.FormatBalance(balance, unit),
			Unit:    unit,
		})
	}

	// Output results
	return utils.OutputBalances(balanceData, outputFormat)
}