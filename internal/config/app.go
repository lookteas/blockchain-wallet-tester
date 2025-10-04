package config

import (
	"os"
	"strings"
)

// AppConfig 应用配置
type AppConfig struct {
	EnvFile string
}

// LoadAppConfig 加载应用配置
func LoadAppConfig() *AppConfig {
	config := &AppConfig{
		EnvFile: getEnvFilePath(),
	}
	return config
}

// getEnvFilePath 获取环境变量文件路径
func getEnvFilePath() string {
	// 按优先级查找.env文件
	envFiles := []string{".env", "configs/.env", "configs/env.example"}

	for _, envFile := range envFiles {
		if _, err := os.Stat(envFile); err == nil {
			return envFile
		}
	}

	// 如果都没有找到，返回默认路径
	return ".env"
}

// GetEnvVar 获取环境变量值
func GetEnvVar(key string) string {
	// 先从系统环境变量获取
	if value := os.Getenv(key); value != "" {
		return value
	}

	// 如果系统环境变量中没有，尝试从.env文件读取
	return getEnvFromFile(key)
}

// getEnvFromFile 从.env文件读取环境变量
func getEnvFromFile(envVar string) string {
	// 尝试从常见的.env文件读取
	envFiles := []string{".env", "configs/.env", "configs/env.example"}

	for _, envFile := range envFiles {
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			continue
		}

		content, err := os.ReadFile(envFile)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == envVar {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}

