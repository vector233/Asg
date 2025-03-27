package ui

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/vector233/AsgGPT/internal/automation"
	"github.com/vector233/AsgGPT/internal/i18n"
)

// getMousePosition captures the current mouse position
func (g *GUI) getMousePosition() {
	infoDialog := dialog.NewInformation(
		i18n.T("get_position_title"),
		i18n.T("get_position_desc"),
		g.window)
	infoDialog.SetOnClosed(func() {
		time.Sleep(1 * time.Second)

		go func() {
			x, y, err := automation.GetMouseClickPosition(10) // 10 seconds timeout
			fmt.Println("GetMouseClickPosition: ", x, y)

			if err != nil {
				dialog.ShowError(err, g.window)
				return
			}

			g.statusLabel.SetText(i18n.Tf("position_captured", x, y))

			clickAction := fmt.Sprintf(`{
      "type": "move",
      "x": %d,
      "y": %d
    },
    {
      "type": "click",
      "button": "left"
    }`, x, y)

			g.window.Clipboard().SetContent(clickAction)

			dialog.ShowCustomConfirm(
				i18n.T("use_position"),
				i18n.T("insert_to_editor"),
				i18n.T("close"),
				widget.NewLabel(i18n.Tf("position_copied", x, y)),
				func(insert bool) {
					if insert {
						g.insertOrReplaceJSON(clickAction)
					}
				},
				g.window,
			)
		}()
	})
	infoDialog.Show()
}

// getProcessInfo retrieves information about running processes
func (g *GUI) getProcessInfo() {
	progress := dialog.NewProgressInfinite(
		i18n.T("get_app_list"),
		i18n.T("getting_app_list"),
		g.window)
	progress.Show()

	go func() {
		processes, err := automation.GetRunningProcesses()
		progress.Hide()

		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}

		if len(processes) == 0 {
			dialog.ShowInformation(
				i18n.T("no_apps_found_title"),
				i18n.T("no_apps_found"),
				g.window)
			return
		}

		var allItems []string
		for _, p := range processes {
			allItems = append(allItems, p.Name)
		}

		searchEntry := widget.NewEntry()
		searchEntry.SetPlaceHolder(i18n.T("search_app"))

		filteredItems := allItems

		processList := widget.NewList(
			func() int {
				return len(filteredItems)
			},
			func() fyne.CanvasObject {
				return widget.NewLabel("")
			},
			func(id widget.ListItemID, obj fyne.CanvasObject) {
				obj.(*widget.Label).SetText(filteredItems[id])
			},
		)

		searchEntry.OnChanged = func(text string) {
			if text == "" {
				filteredItems = allItems
			} else {
				filteredItems = []string{}
				for _, item := range allItems {
					if strings.Contains(strings.ToLower(item), strings.ToLower(text)) {
						filteredItems = append(filteredItems, item)
					}
				}
			}
			processList.Refresh()
		}

		processList.OnSelected = func(id widget.ListItemID) {
			selectedName := filteredItems[id]
			var selectedProcess automation.ProcessInfo
			for _, p := range processes {
				if p.Name == selectedName {
					selectedProcess = p
					break
				}
			}

			// 如果是Windows系统，尝试获取窗口句柄
			if runtime.GOOS == "windows" && selectedProcess.WindowHandle == 0 {
				// 显示进度对话框
				progressBar := widget.NewProgressBarInfinite()
				handleProgress := dialog.NewCustomWithoutButtons(
					i18n.T("getting_window_handle"),
					container.NewVBox(
						widget.NewLabel(i18n.T("searching_windows")),
						progressBar,
					),
					g.window)
				handleProgress.Show()

				go func() {
					// 尝试通过进程名获取窗口句柄
					windows, err := automation.GetWindowHandlesByProcessName(selectedProcess.Name)
					handleProgress.Hide()

					if err == nil && len(windows) > 0 {
						// 使用找到的第一个窗口句柄
						selectedProcess.WindowHandle = windows[0].WindowHandle
						selectedProcess.WindowTitle = windows[0].WindowTitle

						// 如果找到多个窗口，可以考虑让用户选择
						if len(windows) > 1 {
							g.statusLabel.SetText(i18n.Tf("found_multiple_windows", len(windows)))
						} else {
							g.statusLabel.SetText(i18n.Tf("found_window_handle", selectedProcess.WindowHandle))
						}
					}

					g.generateActivateConfig(selectedProcess)
				}()
			} else {
				g.generateActivateConfig(selectedProcess)
			}
		}

		content := container.NewBorder(
			container.NewVBox(
				widget.NewLabel(i18n.T("select_app_desc")),
				searchEntry,
			),
			nil, nil, nil,
			container.NewScroll(processList),
		)

		content.Resize(fyne.NewSize(500, 400))

		listDialog := dialog.NewCustom(
			i18n.T("select_app"),
			i18n.T("cancel"),
			content,
			g.window)

		g.activeDialogs = append(g.activeDialogs, listDialog)

		listDialog.Resize(fyne.NewSize(550, 450))
		listDialog.Show()
	}()
}

// generateActivateConfig generates activation configuration for the selected process
func (g *GUI) generateActivateConfig(processInfo automation.ProcessInfo) {
	options := []string{}

	if runtime.GOOS == "darwin" {
		if processInfo.BundleID != "" {
			options = append(options, i18n.T("use_bundle_id"))
		}
	} else {
		options = append(options, i18n.T("use_window_handle"))
	}

	options = append(options, i18n.T("use_process_name"))

	if processInfo.Path != "" {
		options = append(options, i18n.T("use_app_path"))
	}

	if len(options) == 0 {
		dialog.ShowError(fmt.Errorf(i18n.T("no_activation_method")), g.window)
		return
	}

	optionSelect := widget.NewRadioGroup(options, nil)
	optionSelect.SetSelected(options[0]) // 默认选择第一个选项

	infoItems := []fyne.CanvasObject{
		widget.NewLabel(fmt.Sprintf(i18n.T("app_name"), processInfo.Name)),
	}

	if runtime.GOOS == "darwin" {
		if processInfo.BundleID != "" {
			infoItems = append(infoItems, widget.NewLabel(fmt.Sprintf(i18n.T("bundle_id"), processInfo.BundleID)))
		}
	} else if runtime.GOOS == "windows" {
		infoItems = append(infoItems, widget.NewLabel(fmt.Sprintf(i18n.T("process_id"), processInfo.PID)))
	}

	if processInfo.Path != "" {
		infoItems = append(infoItems, widget.NewLabel(fmt.Sprintf(i18n.T("app_path"), processInfo.Path)))
	}

	infoItems = append(infoItems,
		widget.NewLabel(i18n.T("select_config_method")),
		optionSelect,
		widget.NewButton(i18n.T("confirm"), func() {
			var activateConfig string

			switch optionSelect.Selected {
			case i18n.T("use_bundle_id"):
				activateConfig = fmt.Sprintf(`{
  "type": "activate",
  "bundle_id": "%s"
}`, processInfo.BundleID)
			case i18n.T("use_window_handle"):
				activateConfig = fmt.Sprintf(`{
  "type": "activate",
  "window_handle": %d,
  "process_name": "%s"
}`, processInfo.WindowHandle, processInfo.Name)
			case i18n.T("use_process_name"):
				activateConfig = fmt.Sprintf(`{
  "type": "activate",
  "process_name": "%s"
}`, processInfo.Name)
			case i18n.T("use_app_path"):
				// 转义路径中的反斜杠
				escapedPath := strings.ReplaceAll(processInfo.Path, "\\", "\\\\")
				activateConfig = fmt.Sprintf(`{
  "type": "activate",
  "app_path": "%s"
}`, escapedPath)
			default:
				return
			}

			if len(g.activeDialogs) > 0 {
				g.activeDialogs[len(g.activeDialogs)-1].Hide()
			}

			if activateConfig != "" {
				g.window.Clipboard().SetContent(activateConfig)

				confirmDialog := dialog.NewCustomConfirm(
					i18n.T("activate_config_title"),
					i18n.T("insert_to_editor"),
					i18n.T("close"),
					widget.NewLabel(i18n.T("activate_config_copied")),
					func(insert bool) {
						if insert {
							if g.jsonEditor.Text == "" {
								g.jsonEditor.SetText(activateConfig)
							} else {
								currentText := g.jsonEditor.Text
								if strings.HasSuffix(strings.TrimSpace(currentText), "}") {
									currentText = strings.TrimRight(strings.TrimSpace(currentText), "}") + ",\n  " + activateConfig + "\n}"
									g.jsonEditor.SetText(currentText)
								} else {
									g.jsonEditor.SetText(activateConfig)
								}
							}
						}

						g.closeAllDialogs()
					},
					g.window,
				)

				g.activeDialogs = append(g.activeDialogs, confirmDialog)

				confirmDialog.Show()
			}
		}),
	)

	configDialog := dialog.NewCustom(
		i18n.T("select_config_method"),
		i18n.T("cancel"),
		container.NewVBox(infoItems...),
		g.window)

	g.activeDialogs = append(g.activeDialogs, configDialog)

	configDialog.Show()
}

// closeAllDialogs closes all active dialogs
func (g *GUI) closeAllDialogs() {
	go func() {
		time.Sleep(50 * time.Millisecond)
		for _, d := range g.activeDialogs {
			d.Hide()
		}
		g.activeDialogs = nil
	}()
}

// getForegroundApp gets information about the foreground application
func (g *GUI) getForegroundApp() {
	infoDialog := dialog.NewInformation(
		i18n.T("foreground_app_title"),
		i18n.T("foreground_app_desc"),
		g.window)

	infoDialog.SetOnClosed(func() {
		time.Sleep(1 * time.Second)

		go func() {
			processInfo, err := automation.GetForegroundProcess()
			if err != nil {
				dialog.ShowError(err, g.window)
				return
			}

			if runtime.GOOS == "darwin" && processInfo.BundleID != "" {
				g.statusLabel.SetText(i18n.Tf("app_info_with_bundle", processInfo.Name, processInfo.BundleID))
			} else {
				g.statusLabel.SetText(i18n.Tf("app_info", processInfo.Name))
			}

			g.generateActivateConfig(processInfo)
		}()
	})

	infoDialog.Show()
}

// Helper function to insert or replace JSON in the editor
func (g *GUI) insertOrReplaceJSON(newJSON string) {
	if g.jsonEditor.Text == "" {
		g.jsonEditor.SetText(newJSON)
	} else {
		currentText := g.jsonEditor.Text
		if strings.HasSuffix(strings.TrimSpace(currentText), "}") {
			currentText = strings.TrimRight(strings.TrimSpace(currentText), "}") + ",\n  " + newJSON + "\n}"
			g.jsonEditor.SetText(currentText)
		} else {
			g.jsonEditor.SetText(newJSON)
		}
	}
}
