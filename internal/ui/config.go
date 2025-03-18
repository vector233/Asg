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
	"github.com/vector233/AsgGPT/internal/automation"
	"github.com/vector233/AsgGPT/internal/i18n"
	"github.com/vector233/AsgGPT/pkg/utils"
)

// UIConfig 存储 UI 相关配置
type UIConfig struct {
	ConfigDir string `json:"config_dir"`
}

// 获取 UI 配置文件路径
func getUIConfigPath() string {
	configDir := utils.GetConfigDir()
	return filepath.Join(configDir, "ui_config.json")
}

// 保存配置目录设置
func saveConfigDir(dir string) error {
	config := UIConfig{
		ConfigDir: dir,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	configPath := getUIConfigPath()
	// 确保目录存在
	os.MkdirAll(filepath.Dir(configPath), 0755)

	return os.WriteFile(configPath, data, 0644)
}

// 加载配置目录设置
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

// initConfigDir 初始化配置目录
func (g *GUI) initConfigDir() {
	// 获取基础配置目录
	baseDir := filepath.Dir(utils.GetConfigDir())
	// 配置目录设置为 examples 子目录
	g.configDir = filepath.Join(baseDir, "examples")

	// 尝试从设置中加载配置目录
	if dir, err := loadConfigDir(); err == nil && dir != "" {
		g.configDir = dir
	}

	// 确保 examples 目录存在
	os.MkdirAll(g.configDir, 0755)

	// 生成示例配置文件
	g.generateExampleConfigs()

	// 更新配置文件列表
	g.updateConfigFiles()
}

// generateExampleConfigs 生成示例配置文件
func (g *GUI) generateExampleConfigs() {
	// 获取当前语言
	currentLang := i18n.GetCurrentLang()

	// 根据语言选择示例配置内容
	var defaultExample string

	if currentLang == i18n.LangZH {
		// 中文示例
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
      "text": "这是通过AsgGPT自动化工具输入的文本",
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
		// 英文示例
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
      "text": "This is text input through the AsgGPT automation tool",
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

	// 目标文件路径
	targetPath := filepath.Join(g.configDir, "default.json")

	// 检查目标文件是否已存在
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		// 目标文件不存在，创建示例配置
		os.WriteFile(targetPath, []byte(defaultExample), 0644)
	}
}

// updateConfigFiles 更新配置文件列表
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

	// 如果已经创建了下拉框，更新其选项
	if g.configSelect != nil {
		g.configSelect.Options = g.configFiles
	}
}

// createToolbar 创建工具栏
func (g *GUI) createToolbar() fyne.CanvasObject {
	// 配置目录按钮
	dirButton := widget.NewButtonWithIcon(i18n.T("config_dir"), theme.FolderOpenIcon(), func() {
		g.selectConfigDir()
	})

	// 创建配置文件选择下拉框
	g.configSelect = widget.NewSelect(g.configFiles, func(selected string) {
		g.loadConfigFile(selected)
	})

	// 刷新按钮
	refreshButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		g.updateConfigFiles()
		g.statusLabel.SetText(i18n.T("config_refreshed"))
	})

	// 配置选择容器
	configSelectContainer := container.NewBorder(
		nil, nil, nil, refreshButton,
		g.configSelect,
	)

	// 创建语言选择器
	langSelector := g.createLanguageSelector()

	// 返回工具栏
	return container.NewHBox(
		dirButton,
		widget.NewLabel(i18n.T("config_file")),
		configSelectContainer,
		langSelector,
	)
}

// selectConfigDir 选择配置目录
func (g *GUI) selectConfigDir() {
	// 打开目录选择对话框
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}
		if uri == nil {
			return
		}

		// 更新配置目录
		g.configDir = uri.Path()
		// 保存配置目录设置
		saveConfigDir(g.configDir)
		// 更新配置文件列表
		g.updateConfigFiles()
		g.configSelect.SetSelected("")

		g.statusLabel.SetText(i18n.Tf("config_dir_set", g.configDir))
	}, g.window)
}

// loadConfigFile 加载配置文件
func (g *GUI) loadConfigFile(selected string) {
	if selected == "" {
		return
	}

	// 加载选中的配置文件
	configPath := filepath.Join(g.configDir, selected)
	data, err := os.ReadFile(configPath)
	if err != nil {
		g.statusLabel.SetText(i18n.Tf("load_config_failed", err))
		return
	}

	// 格式化 JSON 以便显示
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		g.jsonEditor.SetText(string(data))
	} else {
		g.jsonEditor.SetText(prettyJSON.String())
	}

	g.statusLabel.SetText(fmt.Sprintf(i18n.T("config_loaded"), selected))
}

// saveConfig 保存配置
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

		// 保存后刷新配置文件列表
		g.updateConfigFiles()
	}, g.window)

	// 设置默认保存目录和文件名
	saveDialog.SetFileName("config.json")

	// 确保目录存在
	os.MkdirAll(g.configDir, 0755)

	// 设置保存位置为当前配置目录
	listURI, err := storage.ListerForURI(storage.NewFileURI(g.configDir))
	if err == nil {
		saveDialog.SetLocation(listURI)
	}
	saveDialog.Show()
}

// executeConfig 执行配置
func (g *GUI) executeConfig() {
	jsonStr := g.jsonEditor.Text
	if jsonStr == "" {
		g.statusLabel.SetText(i18n.T("no_config"))
		return
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "auto-config-*.json")
	if err != nil {
		dialog.ShowError(fmt.Errorf(i18n.Tf("create_temp_file_failed", err)), g.window)
		return
	}

	// 将删除操作移到执行完成后
	tempFilePath := tempFile.Name()

	// 写入配置
	_, err = tempFile.WriteString(jsonStr)
	if err != nil {
		dialog.ShowError(fmt.Errorf(i18n.Tf("write_config_failed", err)), g.window)
		os.Remove(tempFilePath) // 如果写入失败，立即删除文件
		return
	}
	tempFile.Close()

	g.statusLabel.SetText(i18n.T("executing"))

	// 在后台执行
	go func() {
		err := automation.ExecuteConfigFile(tempFilePath)
		// 执行完成后删除临时文件
		os.Remove(tempFilePath)

		if err != nil {
			g.statusLabel.SetText(i18n.Tf("execution_failed", err))
		} else {
			g.statusLabel.SetText(i18n.T("execution_complete"))
		}
	}()
}
