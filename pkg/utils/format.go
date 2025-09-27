package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/params"
	"github.com/olekukonko/tablewriter"
)

// FormatBalance 格式化余额显示
func FormatBalance(balance *big.Int, unit string) string {
	if balance == nil {
		return "0"
	}

	switch strings.ToLower(unit) {
	case "wei":
		return balance.String()
	case "gwei":
		gwei := new(big.Int).Div(balance, big.NewInt(params.GWei))
		return gwei.String()
	case "ether", "eth":
		ether := new(big.Float).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt(big.NewInt(params.Ether)))
		return ether.Text('f', 6)
	default:
		return balance.String()
	}
}

// ParseAmount 解析金额字符串为wei
func ParseAmount(amount, unit string) (*big.Int, error) {
	if amount == "" {
		return big.NewInt(0), nil
	}

	// 解析浮点数
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的金额格式: %s", amount)
	}

	// 转换为big.Float
	amountBigFloat := big.NewFloat(amountFloat)

	// 根据单位转换为wei
	var multiplier *big.Int
	switch strings.ToLower(unit) {
	case "wei":
		multiplier = big.NewInt(1)
	case "gwei":
		multiplier = big.NewInt(params.GWei)
	case "ether", "eth":
		multiplier = big.NewInt(params.Ether)
	default:
		multiplier = big.NewInt(1)
	}

	// 计算最终金额
	result := new(big.Float).Mul(amountBigFloat, new(big.Float).SetInt(multiplier))
	
	// 转换为big.Int
	amountWei, _ := result.Int(nil)
	return amountWei, nil
}

// BalanceData 余额数据结构
type BalanceData struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
	Unit    string `json:"unit"`
}

// TransferResultData 转账结果数据结构
type TransferResultData struct {
	From            string `json:"from"`
	To              string `json:"to"`
	Amount          string `json:"amount"`
	TxHash          string `json:"tx_hash"`
	Status          string `json:"status"`
	GasUsed         string `json:"gas_used,omitempty"`
	GasPrice        string `json:"gas_price,omitempty"`
	Error           string `json:"error,omitempty"`
	ConfirmationTime string `json:"confirmation_time,omitempty"`
}

// OutputBalances 输出余额信息
func OutputBalances(balances []BalanceData, format string) error {
	switch strings.ToLower(format) {
	case "json":
		return outputBalancesJSON(balances)
	case "csv":
		return outputBalancesCSV(balances)
	case "table":
		return outputBalancesTable(balances)
	default:
		return outputBalancesTable(balances)
	}
}

// OutputTransferResults 输出转账结果
func OutputTransferResults(results []TransferResultData, format string) error {
	switch strings.ToLower(format) {
	case "json":
		return outputTransferResultsJSON(results)
	case "csv":
		return outputTransferResultsCSV(results)
	case "table":
		return outputTransferResultsTable(results)
	default:
		return outputTransferResultsTable(results)
	}
}

// outputBalancesJSON 输出JSON格式余额
func outputBalancesJSON(balances []BalanceData) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(balances)
}

// outputBalancesCSV 输出CSV格式余额
func outputBalancesCSV(balances []BalanceData) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// 写入标题行
	if err := writer.Write([]string{"Address", "Balance", "Unit"}); err != nil {
		return err
	}

	// 写入数据行
	for _, balance := range balances {
		if err := writer.Write([]string{balance.Address, balance.Balance, balance.Unit}); err != nil {
			return err
		}
	}

	return nil
}

// outputBalancesTable 输出表格格式余额
func outputBalancesTable(balances []BalanceData) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Address", "Balance", "Unit"})
	table.SetBorder(true)
	table.SetRowLine(true)

	for _, balance := range balances {
		table.Append([]string{balance.Address, balance.Balance, balance.Unit})
	}

	table.Render()
	return nil
}

// outputTransferResultsJSON 输出JSON格式转账结果
func outputTransferResultsJSON(results []TransferResultData) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

// outputTransferResultsCSV 输出CSV格式转账结果
func outputTransferResultsCSV(results []TransferResultData) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// 写入标题行
	headers := []string{"From", "To", "Amount", "TxHash", "Status", "GasUsed", "GasPrice", "Error", "ConfirmationTime"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// 写入数据行
	for _, result := range results {
		row := []string{
			result.From,
			result.To,
			result.Amount,
			result.TxHash,
			result.Status,
			result.GasUsed,
			result.GasPrice,
			result.Error,
			result.ConfirmationTime,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// outputTransferResultsTable 输出表格格式转账结果
func outputTransferResultsTable(results []TransferResultData) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"From", "To", "Amount", "TxHash", "Status"})
	table.SetBorder(true)
	table.SetRowLine(true)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	for _, result := range results {
		status := result.Status
		if result.Error != "" {
			status = fmt.Sprintf("%s (%s)", status, result.Error)
		}
		
		// 截断长地址和交易哈希以便显示
		from := TruncateAddress(result.From)
		to := TruncateAddress(result.To)
		txHash := TruncateHash(result.TxHash)
		
		table.Append([]string{from, to, result.Amount, txHash, status})
	}

	table.Render()
	return nil
}

// TruncateAddress 截断地址显示
func TruncateAddress(address string) string {
	if len(address) <= 10 {
		return address
	}
	return address[:6] + "..." + address[len(address)-4:]
}

// TruncateHash 截断哈希显示
func TruncateHash(hash string) string {
	if len(hash) <= 16 {
		return hash
	}
	return hash[:8] + "..." + hash[len(hash)-8:]
}

// PrintSummary 打印转账摘要
func PrintSummary(results []TransferResultData) {
	total := len(results)
	successful := 0
	failed := 0

	for _, result := range results {
		if result.Status == "success" {
			successful++
		} else {
			failed++
		}
	}

	fmt.Printf("\n=== 转账摘要 ===\n")
	fmt.Printf("总计: %d\n", total)
	fmt.Printf("成功: %d\n", successful)
	fmt.Printf("失败: %d\n", failed)
	fmt.Printf("成功率: %.2f%%\n", float64(successful)/float64(total)*100)
}