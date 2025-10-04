package main

import (
	"log"
	"os"
	"strings"

	"transfer-tool/internal/commands"

	"github.com/urfave/cli/v2"
)

// init 程序初始化时自动加载环境变量
func init() {
	// 自动查找并加载.env文件
	envFiles := []string{".env", "configs/.env", "configs/env.example"}

	for _, envFile := range envFiles {
		if _, err := os.Stat(envFile); err == nil {
			loadEnvFile(envFile)
			break
		}
	}
}

// loadEnvFile 加载环境变量文件
func loadEnvFile(envFile string) {
	// 读取文件内容
	content, err := os.ReadFile(envFile)
	if err != nil {
		return
	}

	// 解析环境变量并设置到系统环境变量中
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// 只有当环境变量不存在时才设置
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

func main() {
	app := &cli.App{
		Name:  "transfer-tool",
		Usage: "基于 urfave/cli/v2 的多钱包转账与查询 CLI 工具",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "network",
				Aliases: []string{"n"},
				Value:   "sepolia",
				Usage:   "网络选择: sepolia, goerli, mainnet, bnb, polygon",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "send",
				Aliases:   []string{"s"},
				Usage:     "单笔转账",
				ArgsUsage: "<recipient_address> <amount_in_eth>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "跳过确认提示",
					},
				},
				Action: commands.SendCommand,
			},
			{
				Name:    "balance",
				Aliases: []string{"b"},
				Usage:   "查询所有钱包余额",
				Action:  commands.BalanceCommand,
			},
			{
				Name:      "batch",
				Aliases:   []string{"batch"},
				Usage:     "批量轮询转账",
				ArgsUsage: "--config <config_file>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"c"},
						Usage:    "配置文件路径",
						Required: true,
					},
				},
				Action: commands.BatchCommand,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
