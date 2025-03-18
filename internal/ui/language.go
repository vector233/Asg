package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/vector233/AsgGPT/internal/i18n"
)

// createLanguageSelector 创建语言选择器
func (g *GUI) createLanguageSelector() fyne.CanvasObject {
	// 创建语言选择下拉框
	langSelect := widget.NewSelect(
		[]string{i18n.T("language_zh"), i18n.T("language_en")},
		func(selected string) {
			var lang string
			if selected == i18n.T("language_zh") {
				lang = i18n.LangZH
			} else {
				lang = i18n.LangEN
			}

			// 设置语言
			err := i18n.SetLang(lang)
			if err != nil {
				g.statusLabel.SetText(err.Error())
				return
			}

			// 提示用户重启应用
			g.statusLabel.SetText(i18n.T("restart_required"))
		},
	)

	// 设置当前选中的语言
	currentLang := i18n.GetCurrentLang()
	if currentLang == i18n.LangZH {
		langSelect.SetSelected(i18n.T("language_zh"))
	} else {
		langSelect.SetSelected(i18n.T("language_en"))
	}

	// 创建语言选择容器
	return container.NewHBox(
		widget.NewLabel(i18n.T("language")),
		langSelect,
	)
}
