package automation

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec" // 添加这一行
	"runtime" // 添加这一行
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/vector233/AsgGPT/internal/i18n"
)

// ExecuteConfigFile 执行指定的配置文件
func ExecuteConfigFile(configFile string) error {
	// 读取配置文件
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf(i18n.T("read_config_file_failed"), err)
	}

	// 解析配置
	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return fmt.Errorf(i18n.T("parse_config_file_failed"), err)
	}

	fmt.Printf(i18n.T("executing_automation_task")+"\n", config.Name)
	fmt.Printf(i18n.T("description")+"\n", config.Description)

	// 执行操作
	ExecuteActions(config.Actions)

	return nil
}

// ExecuteActions 执行一系列操作
func ExecuteActions(actions []Action) {
	for i, action := range actions {
		fmt.Printf(i18n.T("executing_action")+"\n", i+1, action.Type)

		// 添加恢复机制，防止单个操作崩溃导致整个程序崩溃
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
				// 安全地执行键盘操作
				safeKeyTap(action.Key, action.Modifiers)

			case "sleep":
				if action.Duration > 0 {
					time.Sleep(time.Duration(action.Duration * float64(time.Second)))
				} else {
					time.Sleep(time.Second) // 默认等待1秒
				}

			// 在 ExecuteActions 函数中的 activate 部分
			case "activate":
				if action.WindowHandle != 0 {
					// 使用窗口句柄激活窗口
					err := ActivateWindowByHandle(action.WindowHandle)
					if err != nil {
						fmt.Printf(i18n.T("activate_window_by_handle_failed")+"\n", err)
						// 如果窗口句柄失效，尝试其他方法
						if action.ProcessName != "" {
							activateProcess(action.ProcessName)
						} else if action.BundleID != "" {
							activateApplicationByBundleID(action.BundleID)
						} else if action.AppPath != "" {
							activateApplicationByPath(action.AppPath)
						}
					} else {
						fmt.Println(i18n.T("window_activation_executed"))
					}
				} else if action.ProcessName != "" {
					// 使用进程名称激活应用
					activateProcess(action.ProcessName)
				} else if action.BundleID != "" {
					// 使用Bundle ID激活应用
					err := activateApplicationByBundleID(action.BundleID)
					if err != nil {
						fmt.Printf(i18n.T("activate_by_bundle_id_failed")+"\n", err)
						// 如果 Bundle ID 失效，尝试其他方法
						if action.ProcessName != "" {
							activateProcess(action.ProcessName)
						} else if action.AppPath != "" {
							activateApplicationByPath(action.AppPath)
						}
					} else {
						fmt.Println(i18n.T("app_activation_executed"))
					}
				} else if action.AppPath != "" {
					// 使用应用路径激活应用
					activateApplicationByPath(action.AppPath)
				} else {
					fmt.Println(i18n.T("activate_requires_identifier"))
				}

			case "if":
				// 处理条件判断
				conditionMet := evaluateCondition(action.Condition)
				fmt.Printf(i18n.T("condition_result")+"\n", action.Condition, conditionMet)

				if conditionMet {
					fmt.Println(i18n.T("executing_then_branch"))
					ExecuteActions(action.ThenActions)
				} else if len(action.ElseActions) > 0 {
					fmt.Println(i18n.T("executing_else_branch"))
					ExecuteActions(action.ElseActions)
				}

			case "for":
				// 处理循环
				count := action.Count
				if count <= 0 {
					count = 1 // 默认至少执行一次
				}

				fmt.Printf(i18n.T("start_loop")+"\n", count)
				for j := 0; j < count; j++ {
					fmt.Printf(i18n.T("loop_iteration")+"\n", j+1)
					ExecuteActions(action.LoopActions)
				}

			default:
				fmt.Printf(i18n.T("unknown_action_type")+"\n", action.Type)
			}
		}()
	}
}

// safeKeyTap 安全地执行键盘操作，添加错误处理
func safeKeyTap(key string, modifiers []string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf(i18n.T("keyboard_operation_failed")+"\n", r)
		}
	}()

	// 检查键名是否有效
	validKeys := map[string]bool{
		"enter": true, "tab": true, "space": true, "backspace": true, "delete": true,
		"escape": true, "up": true, "down": true, "left": true, "right": true,
		"home": true, "end": true, "page_up": true, "page_down": true,
		"f1": true, "f2": true, "f3": true, "f4": true, "f5": true,
		"f6": true, "f7": true, "f8": true, "f9": true, "f10": true,
		"f11": true, "f12": true, "f13": true, "f14": true, "f15": true,
		"f16": true, "f17": true, "f18": true, "f19": true, "f20": true,
	}

	// 如果是单个字符，直接使用 TypeStr 而不是 KeyTap
	if len(key) == 1 && !validKeys[key] {
		if len(modifiers) > 0 {
			fmt.Printf(i18n.T("warning_single_char_modifiers")+"\n", key)
		}
		robotgo.TypeStr(key)
		return
	}

	// 对于有效的特殊键，使用 KeyTap
	if len(modifiers) > 0 {
		// Convert []string to []interface{}
		mods := make([]interface{}, len(modifiers))
		for i, mod := range modifiers {
			mods[i] = mod
		}
		robotgo.KeyTap(key, mods...)
	} else {
		robotgo.KeyTap(key)
	}
}

// evaluateCondition 评估条件表达式
func evaluateCondition(condition string) bool {
	// 这里实现一个简单的条件评估
	// 在实际应用中，你可能需要一个更复杂的表达式解析器

	// 检查窗口是否存在
	if strings.HasPrefix(condition, "window_exists:") {
		processName := strings.TrimPrefix(condition, "window_exists:")
		return checkProcessExists(processName)
	}

	// 检查屏幕上的像素颜色
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

	// 默认返回 true
	return true
}

// parseInt 将字符串转换为整数，出错时返回0
func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// checkProcessExists 检查进程是否存在
// checkProcessExists 检查进程是否存在
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

// activateProcess 激活指定名称的进程窗口
func activateProcess(processName string) {
	if processName == "" {
		fmt.Println(i18n.T("process_name_empty"))
		return
	}

	// 根据操作系统选择不同的激活方法
	switch runtime.GOOS {
	case "darwin":
		activateProcessMac(processName)
	case "windows":
		activateProcessWindows(processName)
	default:
		fmt.Printf(i18n.T("unsupported_os")+"\n", runtime.GOOS)
	}
}

// activateProcessMac 在 macOS 上激活进程
func activateProcessMac(processName string) {
	// 尝试使用 AppleScript 激活应用
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
		// 如果 AppleScript 失败，尝试使用 robotgo
		fallbackActivateProcess(processName)
	}
}

// activateProcessWindows 在 Windows 上激活进程
func activateProcessWindows(processName string) {
	// 移除 .exe 后缀（如果有）
	processNameWithoutExt := strings.TrimSuffix(processName, ".exe")

	// 使用新的 ActivateApplicationByName 函数
	err := ActivateApplicationByName(processNameWithoutExt)
	if err != nil {
		fmt.Printf(i18n.T("activate_by_new_method_failed")+"\n", err)

		// 如果新方法失败，尝试使用原来的方法作为备选
		fallbackActivateProcess(processName)
	}
}

// fallbackActivateProcess 使用 robotgo 作为备选方案激活进程
func fallbackActivateProcess(processName string) {
	// 获取所有进程
	processes, err := robotgo.Process()
	if err != nil {
		fmt.Printf(i18n.T("get_process_list_failed")+"\n", err)
		return
	}

	// 查找特定名称的进程
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

	// 激活找到的进程窗口
	robotgo.ActivePid(targetPid)
	fmt.Printf(i18n.T("window_activated")+"\n", processName)
}

// activateApplicationByBundleID 使用Bundle ID激活应用
func activateApplicationByBundleID(bundleID string) error {
	if bundleID == "" {
		return fmt.Errorf(i18n.T("bundle_id_empty"))
	}

	// Bundle ID 是 macOS 特有的概念
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

// activateApplicationByPath 使用应用路径激活应用
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

// activateApplicationByPathMac 在 macOS 上使用应用路径激活应用
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

// activateApplicationByPathWindows 在 Windows 上使用应用路径激活应用
func activateApplicationByPathWindows(appPath string) error {
	// 尝试启动应用程序
	cmd := exec.Command("cmd", "/c", "start", "", appPath)
	err := cmd.Run()
	if err != nil {
		fmt.Printf(i18n.T("app_launch_failed")+"\n", err)
		return err
	}

	fmt.Printf(i18n.T("app_launched_path")+"\n", appPath)
	return nil
}
