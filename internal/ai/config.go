package ai

// AIConfig represents the AI platform configuration
type AIConfig struct {
	Type       string `json:"type"`        // AI platform type
	APIKey     string `json:"api_key"`     // API key for authentication
	Model      string `json:"model"`       // Model name to use
	Endpoint   string `json:"endpoint"`    // Custom API endpoint
	APIVersion string `json:"api_version"` // API version (Azure specific)
	ProxyURL   string `json:"proxy_url"`   // Proxy URL for API requests
}

// GetAIConfigByType retrieves AI configuration by platform type
func GetAIConfigByType(platformType string) (AIConfig, error) {
	configs, err := LoadAllAIConfigs()
	if err != nil {
		return AIConfig{}, err
	}

	config, exists := configs.Configs[platformType]
	if !exists {
		// Return default configuration if not exists
		config = AIConfig{
			Type: platformType,
		}

		// Set default values based on platform type
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
