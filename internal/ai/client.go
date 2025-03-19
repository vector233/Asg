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

// Supported AI platforms
const (
	AITypeOpenAI   = "openai"
	AITypeAzure    = "azure"
	AITypeBaidu    = "baidu"
	AITypeDeepSeek = "deepseek"
)

// GenerateJSON generates JSON configuration using AI
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

// AIConfigs represents a collection of AI platform configurations
type AIConfigs struct {
	Configs map[string]AIConfig `json:"configs"` // Platform configurations
	Current string             `json:"current"`  // Currently active platform
}

// AIClient represents an AI client instance
type AIClient struct {
	Config AIConfig
}

// NewAIClient creates a new AI client instance
func NewAIClient(config AIConfig) *AIClient {
	return &AIClient{
		Config: config,
	}
}

// AI config file path
var AIConfigFile string

// Initialize config file path
func init() {
    configDir := utils.GetConfigDir()
    AIConfigFile = filepath.Join(configDir, "ai_configs.json")
}

// GetConfigDir returns OS-specific config directory
func GetConfigDir() string {
    var appDataDir string

    switch runtime.GOOS {
    case "windows":
        appData := os.Getenv("APPDATA")
        if appData != "" {
            appDataDir = filepath.Join(appData, "AsgGPT", "configs")
        }
    case "darwin":
        homeDir, err := os.UserHomeDir()
        if err == nil {
            appDataDir = filepath.Join(homeDir, "Library", "Application Support", "AsgGPT", "configs")
        }
    }

    if appDataDir == "" {
        homeDir, err := os.UserHomeDir()
        if err == nil {
            appDataDir = filepath.Join(homeDir, ".AsgGPT", "configs")
        } else {
            appDataDir = filepath.Join(".", "configs")
        }
    }

    return appDataDir
}

// LoadAIConfig loads current AI configuration
func LoadAIConfig() (AIConfig, error) {
    configs, err := LoadAllAIConfigs()
    if err != nil {
        return AIConfig{}, err
    }

    if len(configs.Configs) == 0 || configs.Configs[configs.Current].Type == "" {
        return AIConfig{
            Type:  AITypeOpenAI,
            Model: "gpt-3.5-turbo",
        }, nil
    }

    return configs.Configs[configs.Current], nil
}

// LoadAllAIConfigs loads all AI platform configurations
func LoadAllAIConfigs() (AIConfigs, error) {
    var configs AIConfigs

    if _, err := os.Stat(AIConfigFile); os.IsNotExist(err) {
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

    configData, err := os.ReadFile(AIConfigFile)
    if err != nil {
        return configs, fmt.Errorf(i18n.T("read_config_failed"), err)
    }

    if err := json.Unmarshal(configData, &configs); err != nil {
        return configs, fmt.Errorf(i18n.T("parse_config_failed"), err)
    }

    if configs.Configs == nil {
        configs.Configs = make(map[string]AIConfig)
    }

    return configs, nil
}

// SaveAIConfig saves AI configuration
func SaveAIConfig(config AIConfig) error {
    configs, err := LoadAllAIConfigs()
    if err != nil {
        return err
    }

    configs.Configs[config.Type] = config
    configs.Current = config.Type

    configDir := filepath.Dir(AIConfigFile)
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return fmt.Errorf(i18n.T("create_config_dir_failed"), err)
    }

    configData, err := json.MarshalIndent(configs, "", "  ")
    if err != nil {
        return fmt.Errorf(i18n.T("serialize_config_failed"), err)
    }

    if err := os.WriteFile(AIConfigFile, configData, 0644); err != nil {
        return fmt.Errorf(i18n.T("save_config_failed"), err)
    }

    return nil
}

// SwitchAIConfig switches current AI platform
func SwitchAIConfig(platformType string) (AIConfig, error) {
    configs, err := LoadAllAIConfigs()
    if err != nil {
        return AIConfig{}, err
    }

    config, exists := configs.Configs[platformType]
    if !exists {
        config = AIConfig{
            Type: platformType,
        }

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

    configs.Current = platformType

    configData, err := json.MarshalIndent(configs, "", "  ")
    if err != nil {
        return config, fmt.Errorf(i18n.T("serialize_config_failed"), err)
    }

    configDir := filepath.Dir(AIConfigFile)
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return config, fmt.Errorf(i18n.T("create_config_dir_failed"), err)
    }

    if err := os.WriteFile(AIConfigFile, configData, 0644); err != nil {
        return config, fmt.Errorf(i18n.T("save_config_failed"), err)
    }

    return config, nil
}

// GetAvailablePlatforms returns all configured platforms
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

// ChatMessage represents a chat message structure
type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// generateWithOpenAI generates JSON using OpenAI API
func (c *AIClient) generateWithOpenAI(prompt string) (string, error) {
    endpoint := "https://api.openai.com/v1/chat/completions"
    if c.Config.Endpoint != "" {
        endpoint = c.Config.Endpoint
    }

    systemPrompt := i18n.GetSystemPrompt()

    requestBody := map[string]interface{}{
        "model": c.Config.Model,
        "messages": []ChatMessage{
            {Role: "system", Content: systemPrompt},
            {Role: "user", Content: prompt},
        },
        "temperature": 0.7,
    }

    requestData, err := json.Marshal(requestBody)
    if err != nil {
        return "", fmt.Errorf(i18n.T("serialize_request_failed"), err)
    }

    // Create request
    req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestData))
    if err != nil {
        return "", fmt.Errorf(i18n.T("create_request_failed"), err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)

    // Send request
    client := c.createHTTPClient()
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf(i18n.T("send_request_failed"), err)
    }
    defer resp.Body.Close()

    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf(i18n.T("read_response_failed"), err)
    }

    // Check HTTP status code
    if resp.StatusCode != http.StatusOK {
        // Try to parse error message
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

    // Parse response
    var response map[string]interface{}
    err = json.Unmarshal(body, &response)
    if err != nil {
        return "", fmt.Errorf(i18n.T("parse_response_failed"), err, string(body))
    }

    // Extract generated text
    choices, ok := response["choices"].([]interface{})
    if !ok || len(choices) == 0 {
        return "", fmt.Errorf(i18n.T("response_format_error_choices"), string(body))
    }

    choice, ok := choices[0].(map[string]interface{})
    if !ok {
        return "", fmt.Errorf(i18n.T("response_format_error_choice"), string(body))
    }

    // Try different paths to get content
    var content string

    // Path 1: choices[0].message.content
    if message, ok := choice["message"].(map[string]interface{}); ok {
        if contentStr, ok := message["content"].(string); ok {
            content = contentStr
        }
    }

    // Path 2: choices[0].text
    if content == "" {
        if text, ok := choice["text"].(string); ok {
            content = text
        }
    }

    // Path 3: choices[0].content
    if content == "" {
        if contentStr, ok := choice["content"].(string); ok {
            content = contentStr
        }
    }

    if content == "" {
        return "", fmt.Errorf(i18n.T("response_format_error_content"), string(body))
    }

    // Extract JSON content
    jsonStr := extractJSON(content)
    if jsonStr == "" {
        // If content itself looks like JSON, return it directly
        if strings.HasPrefix(strings.TrimSpace(content), "{") && strings.HasSuffix(strings.TrimSpace(content), "}") {
            return content, nil
        }
        return "", fmt.Errorf(i18n.T("no_valid_json"), content)
    }

    return jsonStr, nil
}

// generateWithAzure generates JSON using Azure OpenAI API
func (c *AIClient) generateWithAzure(prompt string) (string, error) {
    endpoint := c.Config.Endpoint
    if !strings.HasSuffix(endpoint, "/chat/completions") {
        endpoint = strings.TrimSuffix(endpoint, "/") + "/chat/completions"
    }

    systemPrompt := i18n.GetSystemPrompt()

    requestBody := map[string]interface{}{
        "messages": []ChatMessage{
            {Role: "system", Content: systemPrompt},
            {Role: "user", Content: prompt},
        },
        "temperature": 0.7,
    }

    requestData, err := json.Marshal(requestBody)
    if err != nil {
        return "", fmt.Errorf(i18n.T("serialize_request_failed"), err)
    }

    // Create request
    req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestData))
    if err != nil {
        return "", fmt.Errorf(i18n.T("create_request_failed"), err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("api-key", c.Config.APIKey)
    if c.Config.APIVersion != "" {
        req.Header.Set("api-version", c.Config.APIVersion)
    }

    // Send request
    client := c.createHTTPClient()
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf(i18n.T("send_request_failed"), err)
    }
    defer resp.Body.Close()

    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf(i18n.T("read_response_failed"), err)
    }

    // Parse response
    var response map[string]interface{}
    err = json.Unmarshal(body, &response)
    if err != nil {
        return "", fmt.Errorf(i18n.T("parse_response_failed"), err)
    }

    // Extract generated text
    result, ok := response["result"].(string)
    if !ok {
        return "", fmt.Errorf(i18n.T("response_format_error"))
    }

    // Extract JSON content
    jsonStr := extractJSON(result)
    if jsonStr == "" {
        return "", fmt.Errorf(i18n.T("no_valid_json_simple"))
    }

    return jsonStr, nil
}

// generateWithBaidu generates JSON using Baidu API
func (c *AIClient) generateWithBaidu(prompt string) (string, error) {
    endpoint := c.Config.Endpoint
    if endpoint == "" {
        endpoint = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions"
    }

    systemPrompt := i18n.GetSystemPrompt()

    requestBody := map[string]interface{}{
        "messages": []ChatMessage{
            {Role: "system", Content: systemPrompt},
            {Role: "user", Content: prompt},
        },
        "temperature": 0.7,
    }

    requestData, err := json.Marshal(requestBody)
    if err != nil {
        return "", fmt.Errorf(i18n.T("serialize_request_failed"), err)
    }

    // Create request
    req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestData))
    if err != nil {
        return "", fmt.Errorf(i18n.T("create_request_failed"), err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)

    // Send request
    client := c.createHTTPClient()
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf(i18n.T("send_request_failed"), err)
    }
    defer resp.Body.Close()

    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf(i18n.T("read_response_failed"), err)
    }

    // Parse response
    var response map[string]interface{}
    err = json.Unmarshal(body, &response)
    if err != nil {
        return "", fmt.Errorf(i18n.T("parse_response_failed"), err)
    }

    // Extract generated text
    result, ok := response["result"].(string)
    if !ok {
        return "", fmt.Errorf(i18n.T("response_format_error"))
    }

    // Extract JSON content
    jsonStr := extractJSON(result)
    if jsonStr == "" {
        return "", fmt.Errorf(i18n.T("no_valid_json_simple"))
    }

    return jsonStr, nil
}

// generateWithDeepSeek generates JSON using DeepSeek API
func (c *AIClient) generateWithDeepSeek(prompt string) (string, error) {
    endpoint := c.Config.Endpoint
    if endpoint == "" {
        endpoint = "https://api.deepseek.com/v1/chat/completions"
    }

    systemPrompt := i18n.GetSystemPrompt()

    requestBody := map[string]interface{}{
        "model": c.Config.Model,
        "messages": []ChatMessage{
            {Role: "system", Content: systemPrompt},
            {Role: "user", Content: prompt},
        },
        "temperature": 0.7,
    }

    requestData, err := json.Marshal(requestBody)
    if err != nil {
        return "", fmt.Errorf(i18n.T("serialize_request_failed"), err)
    }

    // Create request
    req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestData))
    if err != nil {
        return "", fmt.Errorf(i18n.T("create_request_failed"), err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)

    // Send request
    client := c.createHTTPClient()
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf(i18n.T("send_request_failed"), err)
    }
    defer resp.Body.Close()

    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf(i18n.T("read_response_failed"), err)
    }

    // Parse response
    var response map[string]interface{}
    err = json.Unmarshal(body, &response)
    if err != nil {
        return "", fmt.Errorf(i18n.T("parse_response_failed"), err)
    }

    // Extract generated text
    choices, ok := response["choices"].([]interface{})
    if !ok || len(choices) == 0 {
        return "", fmt.Errorf(i18n.T("response_format_error"))
    }

    choice, ok := choices[0].(map[string]interface{})
    if !ok {
        return "", fmt.Errorf(i18n.T("response_format_error"))
    }

    // Try different paths to get content
    var content string

    // Path 1: choices[0].message.content
    if message, ok := choice["message"].(map[string]interface{}); ok {
        if contentStr, ok := message["content"].(string); ok {
            content = contentStr
        }
    }

    // Path 2: choices[0].text
    if content == "" {
        if text, ok := choice["text"].(string); ok {
            content = text
        }
    }

    // Path 3: choices[0].content
    if content == "" {
        if contentStr, ok := choice["content"].(string); ok {
            content = contentStr
        }
    }

    if content == "" {
        return "", fmt.Errorf(i18n.T("response_format_error"))
    }

    // Extract JSON content
    jsonStr := extractJSON(content)
    if jsonStr == "" {
        // If content itself looks like JSON, return it directly
        if strings.HasPrefix(strings.TrimSpace(content), "{") && strings.HasSuffix(strings.TrimSpace(content), "}") {
            return content, nil
        }
        return "", fmt.Errorf(i18n.T("no_valid_json"), content)
    }

    return jsonStr, nil
}

// extractJSON extracts JSON content from text
func extractJSON(text string) string {
    // Try regex match first
    re := regexp.MustCompile(`(?s)\{.*\}`)
    match := re.FindString(text)

    if match != "" {
        // Validate extracted content
        var js map[string]interface{}
        if err := json.Unmarshal([]byte(match), &js); err == nil {
            return match
        }

        // Try fixing common issues
        fixedMatch := strings.ReplaceAll(match, "'", "\"")
        if err := json.Unmarshal([]byte(fixedMatch), &js); err == nil {
            return fixedMatch
        }
    }

    // Try manual extraction if regex fails
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

// createHTTPClient creates an HTTP client with proxy support
func (c *AIClient) createHTTPClient() *http.Client {
    // Default transport configuration
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: false,
        },
    }

    // Set proxy if configured
    if c.Config.ProxyURL != "" {
        proxyURL, err := url.Parse(c.Config.ProxyURL)
        if err == nil {
            transport.Proxy = http.ProxyURL(proxyURL)
        }
    }

    // Create client
    client := &http.Client{
        Timeout:   time.Second * 60,
        Transport: transport,
    }

    return client
}
