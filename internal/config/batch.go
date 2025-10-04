package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"
)

// BatchConfig 批量转账配置
type BatchConfig struct {
	Transfer struct {
		TokenAddress string `yaml:"token_address"`
	} `yaml:"transfer"`
	DataSources struct {
		RecipientsXlsx string `yaml:"recipients_xlsx"`
	} `yaml:"data_sources"`
	RPCConfig map[string]string `yaml:"rpc_config,omitempty"`
	APIKeys   map[string]string `yaml:"api_keys,omitempty"`
}

// Recipient 接收方信息
type Recipient struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}

// BatchReport 批量转账报告
type BatchReport struct {
	Timestamp time.Time         `json:"timestamp"`
	Network   string            `json:"network"`
	ChainID   string            `json:"chain_id"`
	Summary   *BatchSummary     `json:"summary"`
	Details   []*TransferDetail `json:"details"`
}

// BatchSummary 批量转账汇总
type BatchSummary struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

// TransferDetail 转账详情
type TransferDetail struct {
	Index     int `json:"index"`
	Recipient `json:"recipient"`
	Sender    string `json:"sender"`
	TxHash    string `json:"tx_hash,omitempty"`
	Explorer  string `json:"explorer,omitempty"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

// LoadBatchConfig 加载批量转账配置
func LoadBatchConfig(configFile string) (*BatchConfig, error) {
	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configFile)
	}

	// 读取文件内容
	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML
	var config BatchConfig
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置
	if config.DataSources.RecipientsXlsx == "" {
		return nil, fmt.Errorf("配置文件中缺少 recipients_xlsx 字段")
	}

	// 如果没有配置RPC，尝试加载全局RPC配置
	if config.RPCConfig == nil {
		globalRPC, err := LoadGlobalRPCConfig()
		if err == nil {
			config.RPCConfig = globalRPC
		}
	}

	// 处理环境变量替换
	config = processEnvVariables(config)

	return &config, nil
}

// LoadGlobalRPCConfig 加载全局RPC配置
func LoadGlobalRPCConfig() (map[string]string, error) {
	// 尝试加载合并后的配置文件
	configFile := "configs/config.yaml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configFile)
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML配置
	var config struct {
		RPCConfig map[string]string `yaml:"rpc_config"`
	}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	if config.RPCConfig == nil {
		return nil, fmt.Errorf("配置文件中未找到rpc_config部分")
	}

	// 处理环境变量替换
	for key, value := range config.RPCConfig {
		config.RPCConfig[key] = expandEnvVariables(value)
	}

	return config.RPCConfig, nil
}

// LoadRecipients 从Excel文件加载接收方数据
func LoadRecipients(xlsxFile string) ([]Recipient, error) {
	// 检查文件是否存在
	if _, err := os.Stat(xlsxFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("Excel文件不存在: %s", xlsxFile)
	}

	// 打开Excel文件
	f, err := excelize.OpenFile(xlsxFile)
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("Excel文件中没有工作表")
	}

	// 获取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取Excel数据失败: %v", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel文件至少需要2行数据（标题行+数据行）")
	}

	// 查找address和amount列
	headerRow := rows[0]
	addressCol := -1
	amountCol := -1

	for i, cell := range headerRow {
		switch strings.ToLower(strings.TrimSpace(cell)) {
		case "address":
			addressCol = i
		case "amount":
			amountCol = i
		}
	}

	if addressCol == -1 {
		return nil, fmt.Errorf("Excel文件缺少 'address' 列")
	}
	if amountCol == -1 {
		return nil, fmt.Errorf("Excel文件缺少 'amount' 列")
	}

	// 解析数据行
	recipients := make([]Recipient, 0, len(rows)-1)
	for i, row := range rows[1:] {
		rowNum := i + 2 // Excel行号从1开始，加上标题行

		// 检查列数
		if len(row) <= addressCol || len(row) <= amountCol {
			return nil, fmt.Errorf("第%d行数据不完整", rowNum)
		}

		address := strings.TrimSpace(row[addressCol])
		amountStr := strings.TrimSpace(row[amountCol])

		// 验证地址
		if address == "" {
			return nil, fmt.Errorf("第%d行地址为空", rowNum)
		}

		// 验证金额
		if amountStr == "" {
			return nil, fmt.Errorf("第%d行金额为空", rowNum)
		}

		// 解析金额
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return nil, fmt.Errorf("第%d行金额格式错误: %s", rowNum, amountStr)
		}

		if amount <= 0 {
			return nil, fmt.Errorf("第%d行金额必须大于0: %f", rowNum, amount)
		}

		recipients = append(recipients, Recipient{
			Address: address,
			Amount:  amount,
		})
	}

	if len(recipients) == 0 {
		return nil, fmt.Errorf("没有有效的接收方数据")
	}

	return recipients, nil
}

// SaveReport 保存批量转账报告
func SaveReport(report *BatchReport, filename string) error {
	// 创建JSON文件
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建报告文件失败: %v", err)
	}
	defer file.Close()

	// 编码为JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(report)
	if err != nil {
		return fmt.Errorf("编码报告失败: %v", err)
	}

	return nil
}

// AddSuccessDetail 添加成功记录
func (r *BatchReport) AddSuccessDetail(index int, recipient Recipient, sender, txHash, explorer string) {
	r.Details = append(r.Details, &TransferDetail{
		Index:     index,
		Recipient: recipient,
		Sender:    sender,
		TxHash:    txHash,
		Explorer:  explorer,
		Status:    "success",
	})
	r.Summary.Success++
}

// AddFailedDetail 添加失败记录
func (r *BatchReport) AddFailedDetail(index int, recipient Recipient, errorMsg string) {
	r.Details = append(r.Details, &TransferDetail{
		Index:     index,
		Recipient: recipient,
		Status:    "failed",
		Error:     errorMsg,
	})
	r.Summary.Failed++
}

// processEnvVariables 处理环境变量替换
func processEnvVariables(config BatchConfig) BatchConfig {
	// 处理RPC配置中的环境变量
	if config.RPCConfig != nil {
		for key, value := range config.RPCConfig {
			config.RPCConfig[key] = expandEnvVariables(value)
		}
	}

	// 处理API密钥配置中的环境变量
	if config.APIKeys != nil {
		for key, value := range config.APIKeys {
			config.APIKeys[key] = expandEnvVariables(value)
		}
	}

	return config
}

// expandEnvVariables 展开环境变量
func expandEnvVariables(value string) string {
	// 简单的环境变量替换，支持 ${VAR_NAME} 格式
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
		if envValue := GetEnvVar(envVar); envValue != "" {
			return envValue
		}
	}
	return value
}
