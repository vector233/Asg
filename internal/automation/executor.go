package automation

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"

	"github.com/vector233/Asg/internal/i18n"
	"github.com/vector233/Asg/internal/model"
)

// ExecuteConfigFile executes the specified configuration file
func ExecuteConfigFile(configFile string) error {
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf(i18n.T("read_config_file_failed"), err)
	}

	var config model.Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return fmt.Errorf(i18n.T("parse_config_file_failed"), err)
	}

	fmt.Printf(i18n.T("executing_automation_task")+"\n", config.Name)
	fmt.Printf(i18n.T("description")+"\n", config.Description)

	ExecuteActions(config.Actions)
	return nil
}

// ExecuteActions executes a series of automation actions
func ExecuteActions(actions []model.Action) {
	for i, action := range actions {
		fmt.Printf(i18n.T("executing_action")+"\n", i+1, action.Type)

		// Recovery mechanism to prevent single action crashes
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf(i18n.T("action_execution_failed")+"\n", r)
				}
			}()

			switch action.Type {
			case "move":
				robotgo.Move(action.X, action.Y)

			case "click":
				if action.Button == "" {
					action.Button = "left"
				}
				robotgo.Click(action.Button, false)

			case "type":
				robotgo.TypeStr(action.Text)

			case "key":
				safeKeyTap(action.Key, action.Modifiers)

			case "sleep":
				if action.Duration > 0 {
					time.Sleep(time.Duration(action.Duration * float64(time.Second)))
				} else {
					time.Sleep(time.Second) // Default 1 second
				}

			case "activate":
				handleActivateAction(action)

			case "if":
				handleIfAction(action)

			case "for":
				handleForAction(action)

			default:
				fmt.Printf(i18n.T("unknown_action_type")+"\n", action.Type)
			}
		}()
	}
}

// handleActivateAction handles the activate action type
func handleActivateAction(action model.Action) {
	if action.WindowHandle != 0 {
		err := ActivateWindowByHandle(action.WindowHandle)
		if err != nil {
			fmt.Printf(i18n.T("activate_window_by_handle_failed")+"\n", err)
			fallbackActivation(action)
		} else {
			fmt.Println(i18n.T("window_activation_executed"))
		}
	} else {
		fallbackActivation(action)
	}
}

// fallbackActivation attempts alternative activation methods
func fallbackActivation(action model.Action) {
	if action.ProcessName != "" {
		activateProcess(action.ProcessName)
	} else if action.BundleID != "" {
		err := activateApplicationByBundleID(action.BundleID)
		if err != nil {
			fmt.Printf(i18n.T("activate_by_bundle_id_failed")+"\n", err)
			if action.ProcessName != "" {
				activateProcess(action.ProcessName)
			} else if action.AppPath != "" {
				activateApplicationByPath(action.AppPath)
			}
		} else {
			fmt.Println(i18n.T("app_activation_executed"))
		}
	} else if action.AppPath != "" {
		activateApplicationByPath(action.AppPath)
	} else {
		fmt.Println(i18n.T("activate_requires_identifier"))
	}
}

// handleIfAction handles the if action type
func handleIfAction(action model.Action) {
	conditionMet := evaluateCondition(action.Condition)
	fmt.Printf(i18n.T("condition_result")+"\n", action.Condition, conditionMet)

	if conditionMet {
		fmt.Println(i18n.T("executing_then_branch"))
		ExecuteActions(action.ThenActions)
	} else if len(action.ElseActions) > 0 {
		fmt.Println(i18n.T("executing_else_branch"))
		ExecuteActions(action.ElseActions)
	}
}

// handleForAction handles the for action type
func handleForAction(action model.Action) {
	count := action.Count
	if count <= 0 {
		count = 1 // Default to at least one iteration
	}

	fmt.Printf(i18n.T("start_loop")+"\n", count)
	for j := 0; j < count; j++ {
		fmt.Printf(i18n.T("loop_iteration")+"\n", j+1)
		ExecuteActions(action.LoopActions)
	}
}

// safeKeyTap performs keyboard operations safely with error handling
func safeKeyTap(key string, modifiers []string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf(i18n.T("keyboard_operation_failed")+"\n", r)
		}
	}()

	// For macOS, use AppleScript for better key handling
	if runtime.GOOS == "darwin" && len(modifiers) > 0 {
		if err := macKeyTap(key, modifiers); err == nil {
			return
		}
	}

	// Standardize modifiers
	standardModifiers := standardizeModifiers(modifiers)

	// Map special key names
	specialKeyMap := map[string]string{
		"enter": "enter", "return": "enter", "tab": "tab", "space": "space",
		"backspace": "backspace", "delete": "delete", "escape": "escape", "esc": "escape",
		"up": "up", "down": "down", "left": "left", "right": "right",
		"home": "home", "end": "end", "page_up": "pageup", "page_down": "pagedown",
	}

	// Handle special key mapping
	if mapped, ok := specialKeyMap[strings.ToLower(key)]; ok {
		key = mapped
	}

	// Press all modifier keys
	for _, mod := range standardModifiers {
		robotgo.KeyToggle(mod, "down")
		// Add delay between modifier key presses
		time.Sleep(100 * time.Millisecond)
	}

	// Add delay before main key press
	time.Sleep(200 * time.Millisecond)

	// Handle main key
	if len(key) == 1 {
		// For single characters, use KeyToggle to press and release
		robotgo.KeyToggle(key, "down")
		time.Sleep(50 * time.Millisecond)
		robotgo.KeyToggle(key, "up")
	} else {
		// For special keys, use KeyTap
		robotgo.KeyTap(key)
	}

	// Add delay before releasing modifier keys
	time.Sleep(200 * time.Millisecond)

	// Release all modifier keys in reverse order
	for i := len(standardModifiers) - 1; i >= 0; i-- {
		robotgo.KeyToggle(standardModifiers[i], "up")
		// Add delay between modifier key releases
		time.Sleep(100 * time.Millisecond)
	}

	// Final delay to ensure all keyboard events are completed
	time.Sleep(100 * time.Millisecond)
}

// macKeyTap uses AppleScript to perform key combinations on macOS
func macKeyTap(key string, modifiers []string) error {
	// Map for special keys in AppleScript
	specialKeyMap := map[string]string{
		"enter": "return", "return": "return", "tab": "tab", "space": "space",
		"backspace": "delete", "delete": "delete", "escape": "escape", "esc": "escape",
		"up": "up arrow", "down": "down arrow", "left": "left arrow", "right": "right arrow",
		"home": "home", "end": "end", "page_up": "page up", "page_down": "page down",
	}

	// Convert modifiers to AppleScript format
	var scriptModifiers []string
	for _, mod := range modifiers {
		switch strings.ToLower(mod) {
		case "command", "cmd", "super":
			scriptModifiers = append(scriptModifiers, "command down")
		case "control", "ctrl":
			scriptModifiers = append(scriptModifiers, "control down")
		case "alt", "option":
			scriptModifiers = append(scriptModifiers, "option down")
		case "shift":
			scriptModifiers = append(scriptModifiers, "shift down")
		}
	}

	// Determine the key to use in AppleScript
	var scriptKey string
	var useKeystroke bool = true
	
	if len(key) == 1 {
		// For single characters
		scriptKey = key
	} else if mapped, ok := specialKeyMap[key]; ok {
		// For special keys
		scriptKey = mapped
		// For special keys like space, use key code instead of keystroke
		if key == "space" {
			useKeystroke = false
		}
	} else {
		// For function keys
		if strings.HasPrefix(key, "f") && len(key) <= 3 {
			scriptKey = key
		} else {
			return fmt.Errorf("unsupported key: %s", key)
		}
	}

	// Build the AppleScript
	var script string
	if useKeystroke {
		script = `
			tell application "System Events"
				keystroke "` + scriptKey + `" using {` + strings.Join(scriptModifiers, ", ") + `}
			end tell
		`
	} else {
		// For special keys like space, use key code
		var keyCode string
		switch key {
		case "space":
			keyCode = "49" // Key code for space key
		}
		
		script = `
			tell application "System Events"
				key code ` + keyCode + ` using {` + strings.Join(scriptModifiers, ", ") + `}
			end tell
		`
	}

	// Execute the AppleScript
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// standardizeModifiers converts modifier names to standard format
func standardizeModifiers(modifiers []string) []string {
	result := make([]string, len(modifiers))
	for i, mod := range modifiers {
		switch strings.ToLower(mod) {
		case "command", "cmd", "super":
			result[i] = "command"
		case "control", "ctrl":
			result[i] = "control"
		case "alt", "option":
			result[i] = "alt"
		case "shift":
			result[i] = "shift"
		default:
			result[i] = mod
		}
	}
	return result
}

// evaluateCondition evaluates a condition expression
func evaluateCondition(condition string) bool {
	if strings.HasPrefix(condition, "window_exists:") {
		processName := strings.TrimPrefix(condition, "window_exists:")
		return checkProcessExists(processName)
	}

	if strings.HasPrefix(condition, "pixel_color:") {
		parts := strings.Split(strings.TrimPrefix(condition, "pixel_color:"), ",")
		if len(parts) == 3 {
			x := parseInt(parts[0])
			y := parseInt(parts[1])
			expectedColor := parts[2]
			color := robotgo.GetPixelColor(x, y)
			return strings.EqualFold(color, expectedColor)
		}
	}

	return true
}

// parseInt converts string to integer, returns 0 on error
func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// checkProcessExists checks if a process exists
func checkProcessExists(processName string) bool {
	processes, err := robotgo.Process()
	if err != nil {
		fmt.Printf(i18n.T("get_process_list_failed")+"\n", err)
		return false
	}

	for _, proc := range processes {
		if strings.EqualFold(proc.Name, processName) {
			return true
		}
	}

	return false
}

// activateProcess activates a window by process name
func activateProcess(processName string) {
	if processName == "" {
		fmt.Println(i18n.T("process_name_empty"))
		return
	}

	switch runtime.GOOS {
	case "darwin":
		activateProcessMac(processName)
	case "windows":
		activateProcessWindows(processName)
	default:
		fmt.Printf(i18n.T("unsupported_os")+"\n", runtime.GOOS)
	}
}

// activateProcessMac activates a process on macOS
func activateProcessMac(processName string) {
	script := fmt.Sprintf(`
		tell application "System Events"
			set appRunning to exists (processes where name is "%s")
			if appRunning then
				set frontmost of process "%s" to true
			else
				try
					tell application "%s" to activate
				on error
					return false
				end try
			end if
		end tell
		return true
	`, processName, processName, processName)

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(output)) == "false" {
		fallbackActivateProcess(processName)
	}
}

// activateProcessWindows activates a process on Windows
func activateProcessWindows(processName string) {
	processNameWithoutExt := strings.TrimSuffix(processName, ".exe")

	err := ActivateApplicationByName(processNameWithoutExt)
	if err != nil {
		fmt.Printf(i18n.T("activate_by_new_method_failed")+"\n", err)
		fallbackActivateProcess(processName)
	}
}

// fallbackActivateProcess uses robotgo as a fallback method to activate process
func fallbackActivateProcess(processName string) {
	processes, err := robotgo.Process()
	if err != nil {
		fmt.Printf(i18n.T("get_process_list_failed")+"\n", err)
		return
	}

	var targetPid int
	for _, proc := range processes {
		if strings.EqualFold(proc.Name, processName) {
			targetPid = proc.Pid
			fmt.Printf(i18n.T("found_process")+"\n", proc.Name, targetPid)
			break
		}
	}

	if targetPid == 0 {
		fmt.Printf(i18n.T("process_not_found")+"\n", processName)
		return
	}

	robotgo.ActivePid(targetPid)
	fmt.Printf(i18n.T("window_activated")+"\n", processName)
}

// activateApplicationByBundleID activates an application using its bundle ID (macOS only)
func activateApplicationByBundleID(bundleID string) error {
	if bundleID == "" {
		return fmt.Errorf(i18n.T("bundle_id_empty"))
	}

	if runtime.GOOS != "darwin" {
		return fmt.Errorf(i18n.T("bundle_id_mac_only"))
	}

	script := fmt.Sprintf(`
        try
            tell application id "%s" to activate
            return true
        on error
            return false
        end try
    `, bundleID)

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf(i18n.T("app_activation_failed")+"\n", err)
		return err
	}

	result := strings.TrimSpace(string(output))
	if result == "false" {
		fmt.Printf(i18n.T("bundle_id_app_not_found")+"\n", bundleID)
		return fmt.Errorf(i18n.T("app_not_found"))
	}

	fmt.Printf(i18n.T("app_activated_bundle_id")+"\n", bundleID)
	return nil
}

// activateApplicationByPath activates an application using its path
func activateApplicationByPath(appPath string) error {
	if appPath == "" {
		return fmt.Errorf(i18n.T("app_path_empty"))
	}

	switch runtime.GOOS {
	case "darwin":
		return activateApplicationByPathMac(appPath)
	case "windows":
		return activateApplicationByPathWindows(appPath)
	default:
		return fmt.Errorf(i18n.T("unsupported_os"), runtime.GOOS)
	}
}

// activateApplicationByPathMac activates an application using its path on macOS
func activateApplicationByPathMac(appPath string) error {
	script := fmt.Sprintf(`
		try
			tell application "%s" to activate
			return true
		on error
			return false
		end try
	`, appPath)

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf(i18n.T("app_activation_failed")+"\n", err)
		return err
	}

	result := strings.TrimSpace(string(output))
	if result == "false" {
		fmt.Printf(i18n.T("app_path_not_found")+"\n", appPath)
		return fmt.Errorf(i18n.T("app_not_found"))
	}

	fmt.Printf(i18n.T("app_activated_path")+"\n", appPath)
	return nil
}

// activateApplicationByPathWindows activates an application using its path on Windows
func activateApplicationByPathWindows(appPath string) error {
	cmd := exec.Command("cmd", "/c", "start", "", appPath)
	err := cmd.Run()
	if err != nil {
		fmt.Printf(i18n.T("app_launch_failed")+"\n", err)
		return err
	}

	fmt.Printf(i18n.T("app_launched_path")+"\n", appPath)
	return nil
}
