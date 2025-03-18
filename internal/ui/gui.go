package ui

import (
	"encoding/json"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/vector233/AsgGPT/internal/ai"
	"github.com/vector233/AsgGPT/internal/i18n"
)

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Content string
	IsUser  bool
	Time    time.Time
}

// GUI 结构体用于存储GUI相关的状态和组件
type GUI struct {
	window              fyne.Window
	aiConfig            ai.AIConfig
	client              *ai.AIClient
	chatMessages        []ChatMessage
	chatDisplay         *widget.RichText
	chatScrollContainer *container.Scroll
	messageInput        *widget.Entry
	jsonEditor          *widget.Entry
	statusLabel         *widget.Label
	configDir           string
	configFiles         []string
	configSelect        *widget.Select

	// 新增字段：跟踪活动的对话框
	activeDialogs []dialog.Dialog
}

// RunGUI starts the graphical user interface
func RunGUI() {
	a := app.New()
	a.Settings().SetTheme(theme.DefaultTheme())
	w := a.NewWindow(i18n.T("app_title"))
	w.Resize(fyne.NewSize(1000, 700))

	// 创建GUI实例
	gui := &GUI{
		window: w,
	}

	// 初始化AI配置
	gui.initAIConfig()

	// 初始化聊天界面
	gui.initChatInterface()

	// 初始化JSON编辑器
	gui.initJSONEditor()

	// 初始化配置目录
	gui.initConfigDir()

	// 创建主布局
	content := gui.createMainLayout()

	w.SetContent(content)
	w.ShowAndRun()
}

// initAIConfig 初始化AI配置
func (g *GUI) initAIConfig() {
	aiConfig, err := ai.LoadAIConfig()
	if err != nil {
		dialog.ShowError(fmt.Errorf(i18n.Tf("load_ai_config_failed"), err), g.window)
	}
	g.aiConfig = aiConfig
	g.client = ai.NewAIClient(aiConfig)
}

// initJSONEditor 初始化JSON编辑器
func (g *GUI) initJSONEditor() {
	g.jsonEditor = widget.NewMultiLineEntry()
	g.jsonEditor.SetPlaceHolder(i18n.T("json_editor_placeholder"))
	g.jsonEditor.Wrapping = fyne.TextWrapWord
	g.jsonEditor.SetMinRowsVisible(3)
}

// createJSONContainer 创建JSON编辑器容器
func (g *GUI) createJSONContainer() fyne.CanvasObject {
	// 添加格式化按钮
	formatButton := widget.NewButtonWithIcon(i18n.T("format_json"), theme.DocumentIcon(), func() {
		g.formatJSON()
	})

	// 创建标题栏，包含标题和格式化按钮
	titleBar := container.NewBorder(
		nil, nil, nil, formatButton,
		widget.NewLabel(i18n.T("json_config")),
	)

	return container.NewBorder(
		titleBar,
		nil, nil, nil,
		container.NewScroll(g.jsonEditor),
	)
}

// 添加 formatJSON 方法用于格式化 JSON
func (g *GUI) formatJSON() {
	// 获取当前文本
	currentText := g.jsonEditor.Text
	if currentText == "" {
		return
	}

	// 解析并格式化 JSON
	var jsonData interface{}
	err := json.Unmarshal([]byte(currentText), &jsonData)
	if err != nil {
		dialog.ShowError(fmt.Errorf(i18n.Tf("json_format_error"), err), g.window)
		return
	}

	// 重新格式化 JSON
	formattedJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		dialog.ShowError(fmt.Errorf(i18n.Tf("json_format_failed"), err), g.window)
		return
	}

	// 更新编辑器内容
	g.jsonEditor.SetText(string(formattedJSON))
	g.statusLabel.SetText(i18n.T("json_formatted"))
}

// createMainLayout 创建主布局
func (g *GUI) createMainLayout() fyne.CanvasObject {
	// 创建聊天容器
	chatContainer := g.createChatContainer()

	// 创建JSON编辑器容器
	jsonContainer := g.createJSONContainer()

	// 创建分割视图
	split := container.NewHSplit(
		chatContainer,
		jsonContainer,
	)
	split.SetOffset(0.5) // 设置分割比例

	// 创建工具栏
	toolbar := g.createToolbar()

	// 创建按钮容器
	buttonContainer := g.createButtonContainer()

	// 创建主布局
	return container.NewBorder(
		toolbar,
		container.NewVBox(g.statusLabel, buttonContainer),
		nil,
		nil,
		split,
	)
}

// createButtonContainer 创建按钮容器
func (g *GUI) createButtonContainer() fyne.CanvasObject {
	// 执行按钮
	executeBtn := widget.NewButtonWithIcon(i18n.T("execute_config"), theme.MediaPlayIcon(), func() {
		g.executeConfig()
	})

	// 保存按钮
	saveBtn := widget.NewButtonWithIcon(i18n.T("save_config"), theme.DocumentSaveIcon(), func() {
		g.saveConfig()
	})

	// 获取坐标按钮
	getPositionBtn := widget.NewButtonWithIcon(i18n.T("get_position"), theme.VisibilityIcon(), func() {
		g.getMousePosition()
	})

	// 获取程序信息按钮
	getProcessBtn := widget.NewButtonWithIcon(i18n.T("get_process_info"), theme.ComputerIcon(), func() {
		g.getProcessInfo()
	})
	getForegroundAppBtn := widget.NewButtonWithIcon(i18n.T("get_foreground_app"), theme.ComputerIcon(), func() {
		g.getForegroundApp()
	})

	// 设置按钮
	settingsBtn := widget.NewButtonWithIcon(i18n.T("ai_settings"), theme.SettingsIcon(), func() {
		g.showAISettings()
	})

	// 返回按钮容器
	return container.NewHBox(
		executeBtn,
		saveBtn,
		getPositionBtn,
		getProcessBtn,
		getForegroundAppBtn,
		settingsBtn,
	)
}
