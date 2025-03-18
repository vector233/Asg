package ai

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vector233/AsgGPT/internal/automation"
)

// AIConfigFile 表示 AI 配置文件路径
// const AIConfigFile = "/Users/huangjiaorong/personal/auto/configs/ai_config.json"

// runAIDialog 运行 AI 对话模式
func RunAIDialog() error {
	// 加载 AI 配置
	config, err := LoadAIConfig()
	if err != nil {
		return fmt.Errorf("加载 AI 配置失败: %v", err)
	}

	// 创建 AI 客户端
	client := NewAIClient(config)

	fmt.Println("=== 自动化脚本生成助手 ===")
	fmt.Println("请描述您想要实现的自动化任务，AI 将为您生成相应的配置。")
	fmt.Println("输入 'exit' 或 'quit' 退出对话。")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "exit" || input == "quit" {
			break
		}

		// 生成 JSON 配置
		fmt.Println("正在生成配置，请稍候...")
		jsonStr, err := client.GenerateJSON(input)
		if err != nil {
			fmt.Printf("生成配置失败: %v\n", err)
			continue
		}

		// 显示生成的配置
		fmt.Println("\n=== 生成的配置 ===")
		fmt.Println(jsonStr)

		// 询问是否保存配置
		fmt.Print("\n是否保存此配置? (y/n): ")
		if !scanner.Scan() {
			break
		}

		if strings.ToLower(scanner.Text()) == "y" {
			// 询问保存路径
			fmt.Print("请输入保存文件名 (默认: config.json): ")
			if !scanner.Scan() {
				break
			}

			filename := scanner.Text()
			if filename == "" {
				filename = "config.json"
			}

			// 确保文件有 .json 扩展名
			if !strings.HasSuffix(filename, ".json") {
				filename += ".json"
			}

			// 构建完整路径
			savePath := filepath.Join("/Users/huangjiaorong/personal/auto/examples", filename)

			// 保存配置
			err := os.WriteFile(savePath, []byte(jsonStr), 0644)
			if err != nil {
				fmt.Printf("保存配置失败: %v\n", err)
			} else {
				fmt.Printf("配置已保存到: %s\n", savePath)

				// 询问是否立即执行
				fmt.Print("是否立即执行此配置? (y/n): ")
				if !scanner.Scan() {
					break
				}

				if strings.ToLower(scanner.Text()) == "y" {
					fmt.Println("正在执行配置...")
					err := automation.ExecuteConfigFile(savePath)
					if err != nil {
						fmt.Printf("执行配置失败: %v\n", err)
					} else {
						fmt.Println("配置执行完成!")
					}
				}
			}
		}

		fmt.Println()
	}

	return nil
}

// LoadAIConfig 加载 AI 配置
// func LoadAIConfig() (AIConfig, error) {
// 	var config AIConfig

// 	// 检查配置文件是否存在
// 	if _, err := os.Stat(AIConfigFile); os.IsNotExist(err) {
// 		// 确保配置目录存在
// 		configDir := filepath.Dir(AIConfigFile)
// 		if err := os.MkdirAll(configDir, 0755); err != nil {
// 			return config, fmt.Errorf("创建配置目录失败: %v", err)
// 		}

// 		// 创建默认配置
// 		config = AIConfig{
// 			Type:   AITypeOpenAI,
// 			APIKey: "",
// 			Model:  "gpt-3.5-turbo",
// 		}

// 		// 提示用户输入 API 密钥
// 		fmt.Println("未找到 AI 配置文件，请输入 OpenAI API 密钥:")
// 		scanner := bufio.NewScanner(os.Stdin)
// 		if scanner.Scan() {
// 			config.APIKey = scanner.Text()
// 		}

// 		// 保存配置
// 		configData, err := json.MarshalIndent(config, "", "  ")
// 		if err != nil {
// 			return config, fmt.Errorf("序列化配置失败: %v", err)
// 		}

// 		err = os.WriteFile(AIConfigFile, configData, 0644)
// 		if err != nil {
// 			return config, fmt.Errorf("保存配置失败: %v", err)
// 		}

// 		fmt.Printf("配置已保存到: %s\n", AIConfigFile)
// 	} else {
// 		// 读取现有配置
// 		configData, err := os.ReadFile(AIConfigFile)
// 		if err != nil {
// 			return config, fmt.Errorf("读取配置文件失败: %v", err)
// 		}

// 		err = json.Unmarshal(configData, &config)
// 		if err != nil {
// 			return config, fmt.Errorf("解析配置文件失败: %v", err)
// 		}
// 	}

// 	return config, nil
// }
