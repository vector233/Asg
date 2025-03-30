package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/vector233/Asg/internal/automation"
	"github.com/vector233/Asg/internal/i18n"
	"github.com/vector233/Asg/pkg/utils"
)

// UIConfig stores UI related configuration
type UIConfig struct {
	ConfigDir string `json:"config_dir"`
}

// Get UI configuration file path
func getUIConfigPath() string {
	configDir := utils.GetConfigDir()
	return filepath.Join(configDir, "ui_config.json")
}

// Save configuration directory setting
func saveConfigDir(dir string) error {
	config := UIConfig{
		ConfigDir: dir,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	configPath := getUIConfigPath()
	// Ensure directory exists
	os.MkdirAll(filepath.Dir(configPath), 0755)

	return os.WriteFile(configPath, data, 0644)
}

// Load configuration directory setting
func loadConfigDir() (string, error) {
	configPath := getUIConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	var config UIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return "", err
	}

	return config.ConfigDir, nil
}

// initConfigDir initializes the configuration directory
func (g *GUI) initConfigDir() {
	// Get base configuration directory
	baseDir := filepath.Dir(utils.GetConfigDir())
	// Set configuration directory to examples subdirectory
	g.configDir = filepath.Join(baseDir, "examples")

	// Try to load configuration directory from settings
	if dir, err := loadConfigDir(); err == nil && dir != "" {
		g.configDir = dir
	}

	// Ensure examples directory exists
	os.MkdirAll(g.configDir, 0755)

	// Generate example configuration files
	g.generateExampleConfigs()

	// Update configuration file list
	g.updateConfigFiles()
}

// generateExampleConfigs generates example configuration files
func (g *GUI) generateExampleConfigs() {
	// Get current language
	currentLang := i18n.GetCurrentLang()

	// Select example configuration content based on language
	var defaultExample string

	if currentLang == i18n.LangZH {
		// Chinese example
		defaultExample = `{
  "name": "自动化任务示例",
  "description": "这是一个示例任务，展示了支持的各种操作类型",
  "actions": [
    {
      "type": "move",
      "x": 500,
      "y": 500,
      "description": "移动鼠标到屏幕中央位置"
    },
    {
      "type": "sleep",
      "duration": 1,
      "description": "等待1秒"
    },
    {
      "type": "click",
      "button": "left",
      "description": "执行左键点击"
    },
    {
      "type": "sleep",
      "duration": 0.5,
      "description": "等待0.5秒"
    },
    {
      "type": "type",
      "text": "这是通过Asg自动化工具输入的文本",
      "description": "输入文本"
    },
    {
      "type": "key",
      "key": "return",
      "description": "按回车键"
    }
  ]
}`
	} else {
		// English example
		defaultExample = `{
  "name": "Automation Task Example",
  "description": "This is an example task that demonstrates various supported operation types",
  "actions": [
    {
      "type": "move",
      "x": 500,
      "y": 500,
      "description": "Move mouse to the center of the screen"
    },
    {
      "type": "sleep",
      "duration": 1,
      "description": "Wait for 1 second"
    },
    {
      "type": "click",
      "button": "left",
      "description": "Perform left click"
    },
    {
      "type": "sleep",
      "duration": 0.5,
      "description": "Wait for 0.5 seconds"
    },
    {
      "type": "type",
      "text": "This is text input through the Asg automation tool",
      "description": "Input text"
    },
    {
      "type": "key",
      "key": "return",
      "description": "Press Enter key"
    }
  ]
}`
	}

	// Target file path
	targetPath := filepath.Join(g.configDir, "default.json")

	// Check if target file already exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		// Target file doesn't exist, create example configuration
		os.WriteFile(targetPath, []byte(defaultExample), 0644)
	}
}

// updateConfigFiles updates the configuration file list
func (g *GUI) updateConfigFiles() {
	g.configFiles = []string{}
	files, err := os.ReadDir(g.configDir)
	if err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
				g.configFiles = append(g.configFiles, file.Name())
			}
		}
	}

	// If dropdown has already been created, update its options
	if g.configSelect != nil {
		g.configSelect.Options = g.configFiles
	}
}

// createToolbar creates the toolbar
func (g *GUI) createToolbar() fyne.CanvasObject {
	// Configuration directory button
	dirButton := widget.NewButtonWithIcon(i18n.T("config_dir"), theme.FolderOpenIcon(), func() {
		g.selectConfigDir()
	})

	// Create configuration file selection dropdown
	g.configSelect = widget.NewSelect(g.configFiles, func(selected string) {
		g.loadConfigFile(selected)
	})

	// Refresh button
	refreshButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		g.updateConfigFiles()
		g.statusLabel.SetText(i18n.T("config_refreshed"))
	})

	// Configuration selection container
	configSelectContainer := container.NewBorder(
		nil, nil, nil, refreshButton,
		g.configSelect,
	)

	// Create language selector
	langSelector := g.createLanguageSelector()

	// Return toolbar
	return container.NewHBox(
		dirButton,
		widget.NewLabel(i18n.T("config_file")),
		configSelectContainer,
		langSelector,
	)
}

// selectConfigDir selects the configuration directory
func (g *GUI) selectConfigDir() {
	// Open directory selection dialog
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}
		if uri == nil {
			return
		}

		// Update configuration directory
		g.configDir = uri.Path()
		// Save configuration directory setting
		saveConfigDir(g.configDir)
		// Update configuration file list
		g.updateConfigFiles()
		g.configSelect.SetSelected("")

		g.statusLabel.SetText(i18n.Tf("config_dir_set", g.configDir))
	}, g.window)
}

// loadConfigFile loads the configuration file
func (g *GUI) loadConfigFile(selected string) {
	if selected == "" {
		return
	}

	// Load selected configuration file
	configPath := filepath.Join(g.configDir, selected)
	data, err := os.ReadFile(configPath)
	if err != nil {
		g.statusLabel.SetText(i18n.Tf("load_config_failed", err))
		return
	}

	// Format JSON for display
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		g.jsonEditor.SetText(string(data))
	} else {
		g.jsonEditor.SetText(prettyJSON.String())
	}

	g.statusLabel.SetText(fmt.Sprintf(i18n.T("config_loaded"), selected))
}

// saveConfig saves the configuration
func (g *GUI) saveConfig() {
	jsonStr := g.jsonEditor.Text
	if jsonStr == "" {
		g.statusLabel.SetText(i18n.T("no_savable_config"))
		return
	}

	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		_, err = writer.Write([]byte(jsonStr))
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}

		g.statusLabel.SetText(i18n.T("config_saved"))

		// Refresh configuration file list after saving
		g.updateConfigFiles()
	}, g.window)

	// Set default save directory and filename
	saveDialog.SetFileName("config.json")

	// Ensure directory exists
	os.MkdirAll(g.configDir, 0755)

	// Set save location to current configuration directory
	listURI, err := storage.ListerForURI(storage.NewFileURI(g.configDir))
	if err == nil {
		saveDialog.SetLocation(listURI)
	}
	saveDialog.Show()
}

// executeConfig executes the configuration
func (g *GUI) executeConfig() {
	jsonStr := g.jsonEditor.Text
	if jsonStr == "" {
		g.statusLabel.SetText(i18n.T("no_config"))
		return
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "auto-config-*.json")
	if err != nil {
		dialog.ShowError(fmt.Errorf(i18n.Tf("create_temp_file_failed", err)), g.window)
		return
	}

	// Move deletion operation to after execution completes
	tempFilePath := tempFile.Name()

	// Write configuration
	_, err = tempFile.WriteString(jsonStr)
	if err != nil {
		dialog.ShowError(fmt.Errorf(i18n.Tf("write_config_failed", err)), g.window)
		os.Remove(tempFilePath) // If writing fails, delete file immediately
		return
	}
	tempFile.Close()

	g.statusLabel.SetText(i18n.T("executing"))

	// Execute in background
	go func() {
		err := automation.ExecuteConfigFile(tempFilePath)
		// Delete temporary file after execution completes
		os.Remove(tempFilePath)

		if err != nil {
			g.statusLabel.SetText(i18n.Tf("execution_failed", err))
		} else {
			g.statusLabel.SetText(i18n.T("execution_complete"))
		}
	}()
}
