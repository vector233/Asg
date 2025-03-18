package ai

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/vector233/AsgGPT/internal/i18n"
	"github.com/vector233/AsgGPT/pkg/utils"
)

// 支持的 AI 平台类型
const (
	AITypeOpenAI   = "openai"
	AITypeAzure    = "azure"
	AITypeBaidu    = "baidu"
	AITypeDeepSeek = "deepseek"
)

// GenerateJSON 通过 AI 生成 JSON 配置
func (c *AIClient) GenerateJSON(prompt string) (string, error) {
	switch c.Config.Type {
	case AITypeOpenAI:
		return c.generateWithOpenAI(prompt)
	case AITypeAzure:
		return c.generateWithAzure(prompt)
	case AITypeBaidu:
		return c.generateWithBaidu(prompt)
	case AITypeDeepSeek:
		return c.generateWithDeepSeek(prompt)
	default:
		return "", fmt.Errorf(i18n.T("unsupported_ai_platform"), c.Config.Type)
	}
}

// AIConfigs 表示多个 AI 平台的配置集合
type AIConfigs struct {
	Configs map[string]AIConfig `json:"configs"` // 各平台配置
	Current string              `json:"current"` // 当前使用的平台
}

// AIClient 表示 AI 客户端
type AIClient struct {
	Config AIConfig
}

// NewAIClient 创建一个新的 AI 客户端
func NewAIClient(config AIConfig) *AIClient {
	return &AIClient{
		Config: config,
	}
}

// AIConfigFile 表示 AI 配置文件路径
var AIConfigFile string

// init 初始化配置文件路径
func init() {
	// 获取配置目录
	configDir := utils.GetConfigDir()

	// 使用 filepath.Join 构建跨平台路径
	AIConfigFile = filepath.Join(configDir, "ai_configs.json")
}

// GetConfigDir 获取适合当前操作系统的配置目录
func GetConfigDir() string {
	// 首先尝试获取应用程序数据目录
	var appDataDir string

	// 根据不同操作系统获取适当的配置目录
	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\AsgGPT\configs
		appData := os.Getenv("APPDATA")
		if appData != "" {
			appDataDir = filepath.Join(appData, "AsgGPT", "configs")
		}
	case "darwin":
		// macOS: ~/Library/Application Support/AsgGPT/configs
		homeDir, err := os.UserHomeDir()
		if err == nil {
			appDataDir = filepath.Join(homeDir, "Library", "Application Support", "AsgGPT", "configs")
		}
	}

	// 如果无法确定应用数据目录，则使用用户主目录下的 .AsgGPT 目录
	if appDataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			appDataDir = filepath.Join(homeDir, ".AsgGPT", "configs")
		} else {
			// 最后的后备方案：使用当前目录
			appDataDir = filepath.Join(".", "configs")
		}
	}

	return appDataDir
}

// LoadAIConfig 加载当前 AI 配置
func LoadAIConfig() (AIConfig, error) {
	configs, err := LoadAllAIConfigs()
	if err != nil {
		return AIConfig{}, err
	}

	// 如果没有配置或当前平台不存在，返回默认配置
	if len(configs.Configs) == 0 || configs.Configs[configs.Current].Type == "" {
		return AIConfig{
			Type:  AITypeOpenAI,
			Model: "gpt-3.5-turbo",
		}, nil
	}

	return configs.Configs[configs.Current], nil
}

// LoadAllAIConfigs 加载所有 AI 平台配置
func LoadAllAIConfigs() (AIConfigs, error) {
	var configs AIConfigs

	// 检查配置文件是否存在
	if _, err := os.Stat(AIConfigFile); os.IsNotExist(err) {
		// 创建默认配置
		configs = AIConfigs{
			Configs: map[string]AIConfig{
				AITypeOpenAI: {
					Type:  AITypeOpenAI,
					Model: "gpt-3.5-turbo",
				},
			},
			Current: AITypeOpenAI,
		}
		return configs, nil
	}

	// 读取现有配置
	configData, err := os.ReadFile(AIConfigFile)
	if err != nil {
		return configs, fmt.Errorf(i18n.T("read_config_failed"), err)
	}

	err = json.Unmarshal(configData, &configs)
	if err != nil {
		return configs, fmt.Errorf(i18n.T("parse_config_failed"), err)
	}

	// 确保 Configs 字段已初始化
	if configs.Configs == nil {
		configs.Configs = make(map[string]AIConfig)
	}

	return configs, nil
}

// SaveAIConfig 保存 AI 配置
func SaveAIConfig(config AIConfig) error {
	// 加载所有配置
	configs, err := LoadAllAIConfigs()
	if err != nil {
		return err
	}

	// 更新当前平台的配置
	configs.Configs[config.Type] = config
	configs.Current = config.Type

	// 确保配置目录存在
	configDir := filepath.Dir(AIConfigFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf(i18n.T("create_config_dir_failed"), err)
	}

	// 保存配置
	configData, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf(i18n.T("serialize_config_failed"), err)
	}

	err = os.WriteFile(AIConfigFile, configData, 0644)
	if err != nil {
		return fmt.Errorf(i18n.T("save_config_failed"), err)
	}

	return nil
}

// SwitchAIConfig 切换当前使用的 AI 平台
func SwitchAIConfig(platformType string) (AIConfig, error) {
	// 加载所有配置
	configs, err := LoadAllAIConfigs()
	if err != nil {
		return AIConfig{}, err
	}

	// 检查目标平台配置是否存在
	config, exists := configs.Configs[platformType]
	if !exists {
		// 如果不存在，创建默认配置
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

		configs.Configs[platformType] = config
	}

	// 更新当前平台
	configs.Current = platformType

	// 保存配置
	configData, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return config, fmt.Errorf(i18n.T("serialize_config_failed"), err)
	}

	// 确保配置目录存在
	configDir := filepath.Dir(AIConfigFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return config, fmt.Errorf(i18n.T("create_config_dir_failed"), err)
	}

	err = os.WriteFile(AIConfigFile, configData, 0644)
	if err != nil {
		return config, fmt.Errorf(i18n.T("save_config_failed"), err)
	}

	return config, nil
}

// GetAvailablePlatforms 获取所有已配置的平台
func GetAvailablePlatforms() ([]string, string, error) {
	configs, err := LoadAllAIConfigs()
	if err != nil {
		return nil, "", err
	}

	platforms := make([]string, 0, len(configs.Configs))
	for platform := range configs.Configs {
		platforms = append(platforms, platform)
	}

	return platforms, configs.Current, nil
}

// ChatMessage 表示聊天消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// generateWithOpenAI 使用 OpenAI API 生成 JSON
func (c *AIClient) generateWithOpenAI(prompt string) (string, error) {
	endpoint := "https://api.openai.com/v1/chat/completions"
	if c.Config.Endpoint != "" {
		endpoint = c.Config.Endpoint
	}

	// 获取系统提示（根据当前语言）
	systemPrompt := i18n.GetSystemPrompt()

	// 构建请求体
	requestBody := map[string]interface{}{
		"model": c.Config.Model,
		"messages": []ChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		"temperature": 0.7,
	}

	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf(i18n.T("serialize_request_failed"), err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestData))
	if err != nil {
		return "", fmt.Errorf(i18n.T("create_request_failed"), err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)

	// 发送请求
	client := c.createHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf(i18n.T("send_request_failed"), err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf(i18n.T("read_response_failed"), err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		// 尝试解析错误信息
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if errMsg, ok := errorResp["error"].(map[string]interface{}); ok {
				if message, ok := errMsg["message"].(string); ok {
					return "", fmt.Errorf(i18n.T("api_error"), message)
				}
			}
		}
		return "", fmt.Errorf(i18n.T("api_request_failed"), resp.StatusCode, string(body))
	}

	// 解析响应
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf(i18n.T("parse_response_failed"), err, string(body))
	}

	// 提取生成的文本
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf(i18n.T("response_format_error_choices"), string(body))
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf(i18n.T("response_format_error_choice"), string(body))
	}

	// 尝试不同的路径获取内容
	var content string

	// 路径1: choices[0].message.content
	if message, ok := choice["message"].(map[string]interface{}); ok {
		if contentStr, ok := message["content"].(string); ok {
			content = contentStr
		}
	}

	// 路径2: choices[0].text
	if content == "" {
		if text, ok := choice["text"].(string); ok {
			content = text
		}
	}

	// 路径3: choices[0].content
	if content == "" {
		if contentStr, ok := choice["content"].(string); ok {
			content = contentStr
		}
	}

	if content == "" {
		return "", fmt.Errorf(i18n.T("response_format_error_content"), string(body))
	}

	// 提取 JSON 部分
	jsonStr := extractJSON(content)
	if jsonStr == "" {
		// 如果内容本身看起来像 JSON，直接返回
		if strings.HasPrefix(strings.TrimSpace(content), "{") && strings.HasSuffix(strings.TrimSpace(content), "}") {
			return content, nil
		}
		return "", fmt.Errorf(i18n.T("no_valid_json"), content)
	}

	return jsonStr, nil
}

// generateWithAzure 使用 Azure OpenAI API 生成 JSON
func (c *AIClient) generateWithAzure(prompt string) (string, error) {
	// Azure OpenAI API 的实现类似于 OpenAI
	// 主要区别在于端点和认证方式
	endpoint := c.Config.Endpoint
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint = strings.TrimSuffix(endpoint, "/") + "/chat/completions"
	}

	// 获取系统提示（根据当前语言）
	systemPrompt := i18n.GetSystemPrompt()

	// 构建请求体
	requestBody := map[string]interface{}{
		"messages": []ChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		"temperature": 0.7,
	}

	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf(i18n.T("serialize_request_failed"), err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestData))
	if err != nil {
		return "", fmt.Errorf(i18n.T("create_request_failed"), err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.Config.APIKey)
	if c.Config.APIVersion != "" {
		req.Header.Set("api-version", c.Config.APIVersion)
	}

	// 发送请求
	client := c.createHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf(i18n.T("send_request_failed"), err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf(i18n.T("read_response_failed"), err)
	}

	// 解析响应
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf(i18n.T("parse_response_failed"), err)
	}

	// 提取生成的文本
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf(i18n.T("response_format_error"))
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf(i18n.T("response_format_error"))
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf(i18n.T("response_format_error"))
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf(i18n.T("response_format_error"))
	}

	// 提取 JSON 部分
	jsonStr := extractJSON(content)
	if jsonStr == "" {
		return "", fmt.Errorf(i18n.T("no_valid_json_simple"))
	}

	return jsonStr, nil
}

// generateWithBaidu 使用百度文心 API 生成 JSON
func (c *AIClient) generateWithBaidu(prompt string) (string, error) {
	// 百度文心 API 的实现
	endpoint := c.Config.Endpoint
	if endpoint == "" {
		endpoint = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions"
	}

	// 获取系统提示（根据当前语言）
	systemPrompt := i18n.GetSystemPrompt()

	// 构建请求体
	requestBody := map[string]interface{}{
		"messages": []ChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		"temperature": 0.7,
	}

	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf(i18n.T("serialize_request_failed"), err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestData))
	if err != nil {
		return "", fmt.Errorf(i18n.T("create_request_failed"), err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)

	// 发送请求
	client := c.createHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf(i18n.T("send_request_failed"), err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf(i18n.T("read_response_failed"), err)
	}

	// 解析响应
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf(i18n.T("parse_response_failed"), err)
	}

	// 提取生成的文本 (百度文心的响应格式与OpenAI不同)
	result, ok := response["result"].(string)
	if !ok {
		return "", fmt.Errorf(i18n.T("response_format_error"))
	}

	// 提取 JSON 部分
	jsonStr := extractJSON(result)
	if jsonStr == "" {
		return "", fmt.Errorf(i18n.T("no_valid_json_simple"))
	}

	return jsonStr, nil
}

// generateWithDeepSeek 使用 DeepSeek API 生成 JSON
func (c *AIClient) generateWithDeepSeek(prompt string) (string, error) {
	endpoint := c.Config.Endpoint
	if endpoint == "" {
		endpoint = "https://api.deepseek.com/v1/chat/completions"
	}

	// 获取系统提示（根据当前语言）
	systemPrompt := i18n.GetSystemPrompt()

	// 构建请求体
	requestBody := map[string]interface{}{
		"model": c.Config.Model,
		"messages": []ChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		"temperature": 0.7,
	}

	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf(i18n.T("serialize_request_failed"), err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestData))
	if err != nil {
		return "", fmt.Errorf(i18n.T("create_request_failed"), err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)

	// 发送请求
	client := c.createHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf(i18n.T("send_request_failed"), err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf(i18n.T("read_response_failed"), err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		// 尝试解析错误信息
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if errMsg, ok := errorResp["error"].(map[string]interface{}); ok {
				if message, ok := errMsg["message"].(string); ok {
					return "", fmt.Errorf(i18n.T("api_error"), message)
				}
			}
		}
		return "", fmt.Errorf(i18n.T("api_request_failed"), resp.StatusCode, string(body))
	}

	// 解析响应
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf(i18n.T("parse_response_failed"), err, string(body))
	}

	// 提取生成的文本
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf(i18n.T("response_format_error"))
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf(i18n.T("response_format_error"))
	}

	// 尝试不同的路径获取内容
	var content string

	// 路径1: choices[0].message.content
	if message, ok := choice["message"].(map[string]interface{}); ok {
		if contentStr, ok := message["content"].(string); ok {
			content = contentStr
		}
	}

	// 路径2: choices[0].text
	if content == "" {
		if text, ok := choice["text"].(string); ok {
			content = text
		}
	}

	// 路径3: choices[0].content
	if content == "" {
		if contentStr, ok := choice["content"].(string); ok {
			content = contentStr
		}
	}

	if content == "" {
		return "", fmt.Errorf(i18n.T("response_format_error"))
	}

	// 提取 JSON 部分
	jsonStr := extractJSON(content)
	if jsonStr == "" {
		// 如果内容本身看起来像 JSON，直接返回
		if strings.HasPrefix(strings.TrimSpace(content), "{") && strings.HasSuffix(strings.TrimSpace(content), "}") {
			return content, nil
		}
		return "", fmt.Errorf(i18n.T("no_valid_json"), content)
	}

	return jsonStr, nil
}

// extractJSON 从文本中提取 JSON 部分
func extractJSON(text string) string {
	// 尝试使用正则表达式匹配 JSON 对象
	re := regexp.MustCompile(`(?s)\{.*\}`)
	match := re.FindString(text)

	if match != "" {
		// 验证提取的内容是否为有效的 JSON
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(match), &js); err == nil {
			return match
		}

		// 如果不是有效的 JSON，尝试修复常见问题
		// 例如，将单引号替换为双引号
		fixedMatch := strings.ReplaceAll(match, "'", "\"")
		if err := json.Unmarshal([]byte(fixedMatch), &js); err == nil {
			return fixedMatch
		}
	}

	// 如果正则表达式匹配失败，尝试查找第一个 { 和最后一个 }
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")

	if start != -1 && end != -1 && start < end {
		jsonCandidate := text[start : end+1]
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(jsonCandidate), &js); err == nil {
			return jsonCandidate
		}
	}

	return ""
}

// 添加一个辅助函数来创建带有代理的 HTTP 客户端
func (c *AIClient) createHTTPClient() *http.Client {
	// 默认传输配置
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	// 如果配置了代理，则设置代理
	if c.Config.ProxyURL != "" {
		proxyURL, err := url.Parse(c.Config.ProxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	// 创建客户端
	client := &http.Client{
		Timeout:   time.Second * 60,
		Transport: transport,
	}

	return client
}
