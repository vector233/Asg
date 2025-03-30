package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/vector233/Asg/internal/i18n"
)

// createLanguageSelector creates and returns a language selection widget
func (g *GUI) createLanguageSelector() fyne.CanvasObject {
	langSelect := widget.NewSelect(
		[]string{i18n.T("language_zh"), i18n.T("language_en")},
		func(selected string) {
			var lang string
			if selected == i18n.T("language_zh") {
				lang = i18n.LangZH
			} else {
				lang = i18n.LangEN
			}

			err := i18n.SetLang(lang)
			if err != nil {
				g.statusLabel.SetText(err.Error())
				return
			}

			g.statusLabel.SetText(i18n.T("restart_required"))
		},
	)

	currentLang := i18n.GetCurrentLang()
	if currentLang == i18n.LangZH {
		langSelect.SetSelected(i18n.T("language_zh"))
	} else {
		langSelect.SetSelected(i18n.T("language_en"))
	}

	return container.NewHBox(
		widget.NewLabel(i18n.T("language")),
		langSelect,
	)
}
