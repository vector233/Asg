package ai

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vector233/AsgGPT/internal/automation"
)

// RunAIDialog starts the AI dialog mode
func RunAIDialog() error {
	config, err := LoadAIConfig()
	if err != nil {
		return fmt.Errorf("failed to load AI config: %v", err)
	}

	client := NewAIClient(config)

	fmt.Println("=== Automation Script Generator ===")
	fmt.Println("Please describe the automation task you want to achieve, and I'll generate the configuration for you.")
	fmt.Println("Enter 'exit' or 'quit' to end the dialog.")
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

		fmt.Println("Generating configuration, please wait...")
		jsonStr, err := client.GenerateJSON(input)
		if err != nil {
			fmt.Printf("Failed to generate configuration: %v\n", err)
			continue
		}

		fmt.Println("\n=== Generated Configuration ===")
		fmt.Println(jsonStr)

		fmt.Print("\nSave this configuration? (y/n): ")
		if !scanner.Scan() {
			break
		}

		if strings.ToLower(scanner.Text()) == "y" {
			fmt.Print("Enter filename (default: config.json): ")
			if !scanner.Scan() {
				break
			}

			filename := scanner.Text()
			if filename == "" {
				filename = "config.json"
			}

			if !strings.HasSuffix(filename, ".json") {
				filename += ".json"
			}

			savePath := filepath.Join("/Users/huangjiaorong/personal/auto/examples", filename)

			err := os.WriteFile(savePath, []byte(jsonStr), 0644)
			if err != nil {
				fmt.Printf("Failed to save configuration: %v\n", err)
			} else {
				fmt.Printf("Configuration saved to: %s\n", savePath)

				fmt.Print("Execute this configuration now? (y/n): ")
				if !scanner.Scan() {
					break
				}

				if strings.ToLower(scanner.Text()) == "y" {
					fmt.Println("Executing configuration...")
					err := automation.ExecuteConfigFile(savePath)
					if err != nil {
						fmt.Printf("Failed to execute configuration: %v\n", err)
					} else {
						fmt.Println("Configuration execution completed!")
					}
				}
			}
		}

		fmt.Println()
	}

	return nil
}
