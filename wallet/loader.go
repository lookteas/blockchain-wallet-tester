package wallet

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"blockchain-wallet-tester/config"
	"golang.org/x/term"
)

func LoadWallets(cfg *config.Config) ([]*Wallet, error) {
	if len(cfg.PrivateKeys) == 0 {
		return nil, fmt.Errorf("no private keys provided in configuration")
	}

	var wallets []*Wallet
	for i, pk := range cfg.PrivateKeys {
		pk = strings.TrimSpace(pk)
		if pk == "" {
			continue
		}

		wallet, err := NewWallet(pk)
		if err != nil {
			return nil, fmt.Errorf("failed to create wallet %d: %v", i+1, err)
		}
		wallets = append(wallets, wallet)
	}

	return wallets, nil
}

func LoadWalletsInteractive() ([]*Wallet, error) {
	fmt.Print("Enter number of wallets: ")
	var count int
	_, err := fmt.Scanf("%d", &count)
	if err != nil {
		return nil, fmt.Errorf("invalid input: %v", err)
	}

	if count <= 0 {
		return nil, fmt.Errorf("number of wallets must be greater than 0")
	}

	reader := bufio.NewReader(os.Stdin)
	var wallets []*Wallet

	for i := 0; i < count; i++ {
		fmt.Printf("Enter private key for wallet %d (without 0x prefix): ", i+1)
		
		// 使用安全的密码输入
		privateKeyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %v", err)
		}
		fmt.Println() // 添加换行
		
		privateKey := strings.TrimSpace(string(privateKeyBytes))
		if privateKey == "" {
			return nil, fmt.Errorf("private key cannot be empty")
		}

		wallet, err := NewWallet(privateKey)
		if err != nil {
			return nil, fmt.Errorf("invalid private key for wallet %d: %v", i+1, err)
		}
		wallets = append(wallets, wallet)
	}

	return wallets, nil
}