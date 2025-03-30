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

	"github.com/vector233/Asg/internal/ai"
	"github.com/vector233/Asg/internal/i18n"
)

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Content string
	IsUser  bool
	Time    time.Time
}

// GUI struct stores GUI-related states and components
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

	// Track active dialogs
	activeDialogs     []dialog.Dialog
	defaultConfigFile string
}

// RunGUI starts the graphical user interface
func RunGUI() {
	a := app.New()
	a.Settings().SetTheme(theme.DefaultTheme())
	w := a.NewWindow(i18n.T("app_title"))
	w.Resize(fyne.NewSize(1000, 700))

	// Create GUI instance
	gui := &GUI{
		window: w,
		statusLabel: widget.NewLabel(""), // Initialize status label
	}

	// Initialize AI configuration
	gui.initAIConfig()

	// Initialize chat interface
	gui.initChatInterface()

	// Initialize JSON editor
	gui.initJSONEditor()

	// Initialize configuration directory
	gui.initConfigDir()

	// Create main layout
	content := gui.createMainLayout()

	w.SetContent(content)
	
	// Ensure default config file is selected before showing the interface
	if gui.defaultConfigFile != "" && gui.configSelect != nil {
		gui.configSelect.SetSelected(gui.defaultConfigFile)
		gui.loadConfigFile(gui.defaultConfigFile)
	}
	
	w.ShowAndRun()
}

// initAIConfig initializes AI configuration
func (g *GUI) initAIConfig() {
	aiConfig, err := ai.LoadAIConfig()
	if err != nil {
		dialog.ShowError(fmt.Errorf(i18n.Tf("load_ai_config_failed"), err), g.window)
	}
	g.aiConfig = aiConfig
	g.client = ai.NewAIClient(aiConfig)
}

// initJSONEditor initializes JSON editor
func (g *GUI) initJSONEditor() {
	g.jsonEditor = widget.NewMultiLineEntry()
	g.jsonEditor.SetPlaceHolder(i18n.T("json_editor_placeholder"))
	g.jsonEditor.Wrapping = fyne.TextWrapWord
	g.jsonEditor.SetMinRowsVisible(3)
}

// createJSONContainer creates JSON editor container
func (g *GUI) createJSONContainer() fyne.CanvasObject {
	// Add format button
	formatButton := widget.NewButtonWithIcon(i18n.T("format_json"), theme.DocumentIcon(), func() {
		g.formatJSON()
	})

	// Create title bar with title and format button
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

// formatJSON formats the JSON content
func (g *GUI) formatJSON() {
	// Get current text
	currentText := g.jsonEditor.Text
	if currentText == "" {
		return
	}

	// Parse and format JSON
	var jsonData interface{}
	err := json.Unmarshal([]byte(currentText), &jsonData)
	if err != nil {
		dialog.ShowError(fmt.Errorf(i18n.Tf("json_format_error"), err), g.window)
		return
	}

	// Reformat JSON
	formattedJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		dialog.ShowError(fmt.Errorf(i18n.Tf("json_format_failed"), err), g.window)
		return
	}

	// Update editor content
	g.jsonEditor.SetText(string(formattedJSON))
	g.statusLabel.SetText(i18n.T("json_formatted"))
}

// createMainLayout creates the main layout
func (g *GUI) createMainLayout() fyne.CanvasObject {
	// Create chat container
	chatContainer := g.createChatContainer()

	// Create JSON editor container
	jsonContainer := g.createJSONContainer()

	// Create split view
	split := container.NewHSplit(
		chatContainer,
		jsonContainer,
	)
	split.SetOffset(0.5) // Set split ratio

	// Create toolbar
	toolbar := g.createToolbar()

	// Create button container
	buttonContainer := g.createButtonContainer()

	// Create main layout
	return container.NewBorder(
		toolbar,
		container.NewVBox(g.statusLabel, buttonContainer),
		nil,
		nil,
		split,
	)
}

// createButtonContainer creates button container
func (g *GUI) createButtonContainer() fyne.CanvasObject {
	executeBtn := widget.NewButtonWithIcon(i18n.T("execute_config"), theme.MediaPlayIcon(), func() {
		g.executeConfig()
	})

	saveBtn := widget.NewButtonWithIcon(i18n.T("save_config"), theme.DocumentSaveIcon(), func() {
		g.saveConfig()
	})

	getPositionBtn := widget.NewButtonWithIcon(i18n.T("get_position"), theme.VisibilityIcon(), func() {
		g.getMousePosition()
	})

	getProcessBtn := widget.NewButtonWithIcon(i18n.T("get_process_info"), theme.ComputerIcon(), func() {
		g.getProcessInfo()
	})

	getForegroundAppBtn := widget.NewButtonWithIcon(i18n.T("get_foreground_app"), theme.ComputerIcon(), func() {
		g.getForegroundApp()
	})

	settingsBtn := widget.NewButtonWithIcon(i18n.T("ai_settings"), theme.SettingsIcon(), func() {
		g.showAISettings()
	})

	return container.NewHBox(
		executeBtn,
		saveBtn,
		getPositionBtn,
		getProcessBtn,
		getForegroundAppBtn,
		settingsBtn,
	)
}
