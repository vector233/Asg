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

// getMousePosition 获取鼠标位置
func (g *GUI) getMousePosition() {
	// 显示提示对话框
	infoDialog := dialog.NewInformation(
		i18n.T("get_position_title"),
		i18n.T("get_position_desc"),
		g.window)
	infoDialog.SetOnClosed(func() {
		// 等待一秒让用户准备
		time.Sleep(1 * time.Second)

		// 获取坐标
		go func() {
			x, y, err := automation.GetMouseClickPosition(10) // 10秒超时
			fmt.Println("GetMouseClickPosition: ", x, y)

			if err != nil {
				dialog.ShowError(err, g.window)
				return
			}

			// 将坐标显示在状态栏
			g.statusLabel.SetText(fmt.Sprintf("获取到坐标: X=%d, Y=%d", x, y))

			// 创建一个包含坐标的点击操作（拆分为移动和点击两个独立操作）
			clickAction := fmt.Sprintf(`{
      "type": "move",
      "x": %d,
      "y": %d
    },
    {
      "type": "click",
      "button": "left"
    }`, x, y)

			// 复制到剪贴板
			g.window.Clipboard().SetContent(clickAction)

			// 显示对话框让用户选择如何使用坐标
			dialog.ShowCustomConfirm(
				i18n.T("use_position"),
				i18n.T("insert_to_editor"),
				i18n.T("close"),
				widget.NewLabel(fmt.Sprintf("已获取坐标 X=%d, Y=%d 并复制到剪贴板", x, y)),
				func(insert bool) {
					if insert {
						// 如果JSON编辑器为空，直接设置
						if g.jsonEditor.Text == "" {
							g.jsonEditor.SetText(clickAction)
						} else {
							// 尝试添加到现有JSON
							currentText := g.jsonEditor.Text
							if strings.HasSuffix(strings.TrimSpace(currentText), "}") {
								// 如果是完整的JSON对象，尝试添加为新操作
								currentText = strings.TrimRight(strings.TrimSpace(currentText), "}") + ",\n  " + clickAction + "\n}"
								g.jsonEditor.SetText(currentText)
							} else {
								// 否则直接替换
								g.jsonEditor.SetText(clickAction)
							}
						}
					}
				},
				g.window,
			)
		}()
	})
	infoDialog.Show()
}

// getProcessInfo 获取程序信息
func (g *GUI) getProcessInfo() {
	// 显示加载对话框
	progress := dialog.NewProgressInfinite(
		i18n.T("get_app_list"),
		i18n.T("getting_app_list"),
		g.window)
	progress.Show()

	// 在后台获取应用列表
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

		// 创建应用列表
		var allItems []string
		for _, p := range processes {
			allItems = append(allItems, p.Name)
		}

		// 创建搜索框
		searchEntry := widget.NewEntry()
		searchEntry.SetPlaceHolder(i18n.T("search_app"))

		// 创建过滤后的应用列表
		filteredItems := allItems

		// 创建应用列表控件
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

		// 设置搜索框变更事件
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

		// 设置列表选择事件
		processList.OnSelected = func(id widget.ListItemID) {
			// 查找选中的进程
			selectedName := filteredItems[id]
			var selectedProcess automation.ProcessInfo
			for _, p := range processes {
				if p.Name == selectedName {
					selectedProcess = p
					break
				}
			}

			g.generateActivateConfig(selectedProcess)
		}

		// 创建列表选择对话框
		content := container.NewBorder(
			container.NewVBox(
				widget.NewLabel(i18n.T("select_app_desc")),
				searchEntry,
			),
			nil, nil, nil,
			container.NewScroll(processList),
		)

		// 设置较大的尺寸
		content.Resize(fyne.NewSize(500, 400))

		listDialog := dialog.NewCustom(
			i18n.T("select_app"),
			i18n.T("cancel"),
			content,
			g.window)

		// 将程序列表对话框添加到活动对话框列表
		g.activeDialogs = append(g.activeDialogs, listDialog)

		listDialog.Resize(fyne.NewSize(550, 450))
		listDialog.Show()
	}()
}

// generateActivateConfig 生成activate配置
func (g *GUI) generateActivateConfig(processInfo automation.ProcessInfo) {
	// 创建配置选项
	options := []string{}

	// 根据操作系统添加不同的选项
	if runtime.GOOS == "darwin" {
		if processInfo.BundleID != "" {
			options = append(options, i18n.T("use_bundle_id"))
		}
		// 在 macOS 上不使用窗口句柄选项
	} else {
		// Windows 系统使用窗口句柄
		options = append(options, i18n.T("use_window_handle"))
	}

	// 添加通用选项
	options = append(options, i18n.T("use_process_name"))

	if processInfo.Path != "" {
		options = append(options, i18n.T("use_app_path"))
	}

	// 如果没有选项，显示错误
	if len(options) == 0 {
		dialog.ShowError(fmt.Errorf(i18n.T("no_activation_method")), g.window)
		return
	}

	// 创建选择对话框
	optionSelect := widget.NewRadioGroup(options, nil)
	optionSelect.SetSelected(options[0]) // 默认选择第一个选项

	// 创建信息显示内容
	infoItems := []fyne.CanvasObject{
		widget.NewLabel(fmt.Sprintf(i18n.T("app_name"), processInfo.Name)),
	}

	// 根据操作系统添加不同的信息
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

	// 添加选择和确认按钮
	infoItems = append(infoItems,
		widget.NewLabel(i18n.T("select_config_method")),
		optionSelect,
		widget.NewButton(i18n.T("confirm"), func() {
			var activateConfig string

			switch optionSelect.Selected {
			case i18n.T("use_bundle_id"):
				// macOS 特有的 Bundle ID 激活方式
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

			// 关闭当前对话框（子窗口A）
			if len(g.activeDialogs) > 0 {
				g.activeDialogs[len(g.activeDialogs)-1].Hide()
			}

			if activateConfig != "" {
				// 复制到剪贴板
				g.window.Clipboard().SetContent(activateConfig)

				// 显示对话框让用户选择如何使用配置（子窗口B）
				confirmDialog := dialog.NewCustomConfirm(
					i18n.T("activate_config_title"),
					i18n.T("insert_to_editor"),
					i18n.T("close"),
					widget.NewLabel(i18n.T("activate_config_copied")),
					func(insert bool) {
						if insert {
							// 如果JSON编辑器为空，直接设置
							if g.jsonEditor.Text == "" {
								g.jsonEditor.SetText(activateConfig)
							} else {
								// 尝试添加到现有JSON
								currentText := g.jsonEditor.Text
								if strings.HasSuffix(strings.TrimSpace(currentText), "}") {
									// 如果是完整的JSON对象，尝试添加为新操作
									currentText = strings.TrimRight(strings.TrimSpace(currentText), "}") + ",\n  " + activateConfig + "\n}"
									g.jsonEditor.SetText(currentText)
								} else {
									// 否则直接替换
									g.jsonEditor.SetText(activateConfig)
								}
							}
						}

						// 关闭所有活动对话框
						g.closeAllDialogs()
					},
					g.window,
				)

				// 将确认对话框添加到活动对话框列表
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

	// 将配置对话框添加到活动对话框列表
	g.activeDialogs = append(g.activeDialogs, configDialog)

	configDialog.Show()
}

// closeAllDialogs 关闭所有子窗口
func (g *GUI) closeAllDialogs() {
	// 使用延迟确保UI事件处理完成
	go func() {
		time.Sleep(50 * time.Millisecond)

		// 关闭所有存储的对话框
		for _, d := range g.activeDialogs {
			d.Hide()
		}

		// 清空对话框列表
		g.activeDialogs = nil
	}()
}

// getForegroundApp 获取前台程序信息
func (g *GUI) getForegroundApp() {
	// 显示提示对话框
	infoDialog := dialog.NewInformation(
		i18n.T("foreground_app_title"),
		i18n.T("foreground_app_desc"),
		g.window)

	infoDialog.SetOnClosed(func() {
		// 等待一秒让用户准备
		time.Sleep(1 * time.Second)

		// 获取前台程序信息
		go func() {
			processInfo, err := automation.GetForegroundProcess()
			if err != nil {
				dialog.ShowError(err, g.window)
				return
			}

			// 将程序信息显示在状态栏，根据操作系统显示不同信息
			if runtime.GOOS == "darwin" && processInfo.BundleID != "" {
				g.statusLabel.SetText(fmt.Sprintf("获取到程序: %s (Bundle ID: %s)", processInfo.Name, processInfo.BundleID))
			} else {
				g.statusLabel.SetText(fmt.Sprintf("获取到程序: %s", processInfo.Name))
			}

			// 生成激活配置
			g.generateActivateConfig(processInfo)
		}()
	})

	infoDialog.Show()
}
