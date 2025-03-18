package ai

// AIConfig 表示 AI 配置
type AIConfig struct {
	Type       string `json:"type"`
	APIKey     string `json:"api_key"`
	Model      string `json:"model"`
	Endpoint   string `json:"endpoint"`
	APIVersion string `json:"api_version"`
	ProxyURL   string `json:"proxy_url"` // 添加代理 URL 字段
}

// GetAIConfigByType 根据类型获取 AI 配置
func GetAIConfigByType(platformType string) (AIConfig, error) {
	configs, err := LoadAllAIConfigs()
	if err != nil {
		return AIConfig{}, err
	}

	config, exists := configs.Configs[platformType]
	if !exists {
		// 如果不存在，返回该类型的默认配置
		config = AIConfig{
			Type: platformType,
		}

		// 根据平台类型设置默认值
		switch platformType {
		case AITypeOpenAI:
			config.Model = "gpt-3.5-turbo"
		case AITypeAzure:
			config.Model = "gpt-35-turbo"
			config.APIVersion = "2023-05-15"
		case AITypeBaidu:
			config.Model = "ERNIE-Bot-4"
		case AITypeDeepSeek:
			config.Model = "deepseek-chat"
		}
	}

	return config, nil
}
