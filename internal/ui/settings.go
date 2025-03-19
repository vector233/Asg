package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/vector233/AsgGPT/internal/ai"
	"github.com/vector233/AsgGPT/internal/i18n"
)

// showAISettings displays the AI settings dialog
func (g *GUI) showAISettings() {
	w := fyne.CurrentApp().NewWindow(i18n.T("settings_title"))
	w.Resize(fyne.NewSize(500, 400))

	currentPlatform := g.aiConfig.Type
	if currentPlatform == "" {
		currentPlatform = "openai" // Default platform
	}

	// 只保留 OpenAI 和 DeepSeek 平台
	platforms := []string{"openai", "deepseek"}
	platformSelect := widget.NewSelect(platforms, nil)
	platformSelect.SetSelected(currentPlatform)

	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetText(g.aiConfig.APIKey)

	modelSelect := widget.NewSelect([]string{}, nil)
	modelSelect.SetSelected(g.aiConfig.Model)

	// Updates model options based on selected platform
	// 修改模型选项更新函数
	updateModelOptions := func(platform string) {
		var models []string
		switch platform {
		case "openai":
			models = []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"}
		case "deepseek":
			models = []string{"deepseek-chat", "deepseek-reasoner"}
		default:
			models = []string{}
		}

		modelSelect.Options = models

		platformConfig, err := ai.GetAIConfigByType(platform)
		if err == nil && platformConfig.Model != "" {
			// Use previously configured model if available
			found := false
			for _, m := range models {
				if m == platformConfig.Model {
					found = true
					break
				}
			}
			if !found {
				modelSelect.Options = append(modelSelect.Options, platformConfig.Model)
			}
			modelSelect.SetSelected(platformConfig.Model)
		} else if len(models) > 0 {
			modelSelect.SetSelected(models[0])
		}
	}

	updateModelOptions(currentPlatform)

	endpointEntry := widget.NewEntry()
	endpointEntry.SetText(g.aiConfig.Endpoint)

	apiVersionEntry := widget.NewEntry()
	apiVersionEntry.SetText(g.aiConfig.APIVersion)

	proxyEntry := widget.NewEntry()
	proxyEntry.SetText(g.aiConfig.ProxyURL)
	proxyEntry.SetPlaceHolder("http://127.0.0.1:7890")

	// Handle platform switching
	platformSelect.OnChanged = func(selected string) {
		if selected != currentPlatform {
			newConfig, err := ai.SwitchAIConfig(selected)
			if err != nil {
				dialog.ShowError(fmt.Errorf(i18n.Tf("platform_switch_failed", err)), w)
				return
			}

			apiKeyEntry.SetText(newConfig.APIKey)
			endpointEntry.SetText(newConfig.Endpoint)
			apiVersionEntry.SetText(newConfig.APIVersion)
			proxyEntry.SetText(newConfig.ProxyURL)

			if selected == "deepseek" && newConfig.Endpoint == "" {
				endpointEntry.SetText("https://api.deepseek.com/v1/chat/completions")
			}

			currentPlatform = selected
			updateModelOptions(selected)
		}
	}

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
			newConfig := ai.AIConfig{
				Type:       platformSelect.Selected,
				APIKey:     apiKeyEntry.Text,
				Model:      modelSelect.Selected,
				Endpoint:   endpointEntry.Text,
				APIVersion: apiVersionEntry.Text,
				ProxyURL:   proxyEntry.Text,
			}

			err := ai.SaveAIConfig(newConfig)
			if err != nil {
				dialog.ShowError(fmt.Errorf(i18n.Tf("save_config_failed", err)), w)
				return
			}

			g.aiConfig = newConfig
			g.client = ai.NewAIClient(newConfig)
			g.statusLabel.SetText(i18n.T("settings_updated"))

			w.Close()
		},
		SubmitText: i18n.T("save"),
	}

	w.SetContent(form)
	w.Show()
}
