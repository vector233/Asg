package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/vector233/AsgGPT/internal/ai"
	"github.com/vector233/AsgGPT/internal/i18n"
)

// showAISettings 显示AI设置对话框
func (g *GUI) showAISettings() {
	// 创建设置窗口
	w := fyne.CurrentApp().NewWindow(i18n.T("settings_title"))
	w.Resize(fyne.NewSize(500, 400))

	// 当前平台
	currentPlatform := g.aiConfig.Type
	if currentPlatform == "" {
		currentPlatform = "openai" // 默认平台
	}

	// 创建平台选择下拉框
	platforms := []string{"openai", "azure", "anthropic", "gemini", "deepseek"} // 添加 deepseek
	platformSelect := widget.NewSelect(platforms, nil)
	platformSelect.SetSelected(currentPlatform)

	// 创建API密钥输入框
	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetText(g.aiConfig.APIKey)

	// 创建模型选择下拉框
	modelSelect := widget.NewSelect([]string{}, nil)
	modelSelect.SetSelected(g.aiConfig.Model)

	// 更新模型选项的函数
	updateModelOptions := func(platform string) {
		var models []string
		switch platform {
		case "openai":
			models = []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"}
		case "azure":
			models = []string{"gpt-4", "gpt-3.5-turbo"}
		case "anthropic":
			models = []string{"claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"}
		case "gemini":
			models = []string{"gemini-pro", "gemini-1.5-pro"}
		case "deepseek":
			models = []string{"deepseek-chat", "deepseek-chat"} // DeepSeek 模型选项
		default:
			models = []string{}
		}

		modelSelect.Options = models

		// 获取当前平台的配置
		platformConfig, err := ai.GetAIConfigByType(platform)
		if err == nil && platformConfig.Model != "" {
			// 如果该平台之前有设置过模型，优先使用该模型
			found := false
			for _, m := range models {
				if m == platformConfig.Model {
					found = true
					break
				}
			}
			if !found {
				// 如果之前设置的模型不在预设列表中，添加它
				modelSelect.Options = append(modelSelect.Options, platformConfig.Model)
			}
			modelSelect.SetSelected(platformConfig.Model)
		} else if len(models) > 0 {
			// 否则使用该平台的第一个默认模型
			modelSelect.SetSelected(models[0])
		}
	}

	// 初始化模型选项
	updateModelOptions(currentPlatform)

	endpointEntry := widget.NewEntry()
	endpointEntry.SetText(g.aiConfig.Endpoint)

	apiVersionEntry := widget.NewEntry()
	apiVersionEntry.SetText(g.aiConfig.APIVersion)

	proxyEntry := widget.NewEntry()
	proxyEntry.SetText(g.aiConfig.ProxyURL)
	proxyEntry.SetPlaceHolder("http://127.0.0.1:7890")

	// 平台切换处理
	platformSelect.OnChanged = func(selected string) {
		if selected != currentPlatform {
			// 切换平台配置
			newConfig, err := ai.SwitchAIConfig(selected)
			if err != nil {
				dialog.ShowError(fmt.Errorf(i18n.Tf("platform_switch_failed", err)), w)
				return
			}

			// 更新表单
			apiKeyEntry.SetText(newConfig.APIKey)
			endpointEntry.SetText(newConfig.Endpoint)
			apiVersionEntry.SetText(newConfig.APIVersion)
			proxyEntry.SetText(newConfig.ProxyURL)

			// 为 DeepSeek 设置默认端点
			if selected == "deepseek" && newConfig.Endpoint == "" {
				endpointEntry.SetText("https://api.deepseek.com/v1/chat/completions")
			}

			// 更新当前平台
			currentPlatform = selected

			// 更新模型选项
			updateModelOptions(selected)
		}
	}

	// 创建表单
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: i18n.T("ai_platform"), Widget: platformSelect},
			{Text: i18n.T("api_key"), Widget: apiKeyEntry},
			{Text: i18n.T("model"), Widget: modelSelect},
			{Text: i18n.T("api_endpoint"), Widget: endpointEntry},
			{Text: i18n.T("api_version"), Widget: apiVersionEntry},
			{Text: i18n.T("proxy_url"), Widget: proxyEntry},
		},
		OnSubmit: func() {
			// 保存设置
			newConfig := ai.AIConfig{
				Type:       platformSelect.Selected,
				APIKey:     apiKeyEntry.Text,
				Model:      modelSelect.Selected,
				Endpoint:   endpointEntry.Text,
				APIVersion: apiVersionEntry.Text,
				ProxyURL:   proxyEntry.Text,
			}

			// 保存配置
			err := ai.SaveAIConfig(newConfig)
			if err != nil {
				dialog.ShowError(fmt.Errorf(i18n.Tf("save_config_failed", err)), w)
				return
			}

			// 更新当前配置
			g.aiConfig = newConfig
			g.client = ai.NewAIClient(newConfig)
			g.statusLabel.SetText(i18n.T("settings_updated"))

			// 关闭窗口
			w.Close()
		},
		SubmitText: i18n.T("save"),
	}

	w.SetContent(form)
	w.Show()
}
