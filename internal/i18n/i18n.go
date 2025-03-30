package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/vector233/Asg/pkg/utils"
)

// Supported languages
const (
	LangZH = "zh"
	LangEN = "en"
)

var (
	currentLang  = LangEN // Default language is English
	translations = make(map[string]map[string]string)
	mutex        sync.RWMutex
)

// Initialize translation data
func init() {
	err := loadTranslations()
	if err != nil {
		fmt.Printf("Failed to load translation files: %v\n", err)
	}
}

// Load translation files
func loadTranslations() error {
	configDir := utils.GetConfigDir()
	i18nDir := filepath.Join(configDir, "i18n")

	// Ensure directory exists
	if err := os.MkdirAll(i18nDir, 0755); err != nil {
		return err
	}

	// Get default translations
	defaultZHTranslations := getDefaultZHTranslations()
	defaultENTranslations := getDefaultENTranslations()

	// Load Chinese translations
	zhPath := filepath.Join(i18nDir, "zh.json")
	zhTranslations, err := loadAndUpdateTranslation(zhPath, defaultZHTranslations)
	if err != nil {
		return err
	}

	// Load English translations
	enPath := filepath.Join(i18nDir, "en.json")
	enTranslations, err := loadAndUpdateTranslation(enPath, defaultENTranslations)
	if err != nil {
		return err
	}

	// Save translations
	mutex.Lock()
	translations[LangZH] = zhTranslations
	translations[LangEN] = enTranslations
	mutex.Unlock()

	return nil
}

// Load and update translation file
func loadAndUpdateTranslation(path string, defaultTranslations map[string]string) (map[string]string, error) {
	var translations map[string]string
	var updated bool

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		translations = defaultTranslations
		updated = true
	} else {
		// Read existing translation file
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		// Parse translations
		if err := json.Unmarshal(data, &translations); err != nil {
			return nil, err
		}

		// Check and add missing translations
		for key, value := range defaultTranslations {
			if _, exists := translations[key]; !exists {
				translations[key] = value
				updated = true
			}
		}
	}

	// Save to file if updated
	if updated {
		data, err := json.MarshalIndent(translations, "", "  ")
		if err != nil {
			return nil, err
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			return nil, err
		}
	}

	return translations, nil
}

// GetDefaultZHTranslations 获取默认中文翻译
func getDefaultZHTranslations() map[string]string {
	return map[string]string{
		// GUI通用
		"app_title": "自动化脚本生成助手",
		"save":      "保存",
		"cancel":    "取消",
		"close":     "关闭",
		"confirm":   "确定",

		// 主界面
		"chat_area":         "对话区域",
		"json_config":       "JSON 配置 (可编辑)",
		"copy_last_message": "复制最后消息",
		"send":              "发送",
		"execute_config":    "执行配置",
		"save_config":       "保存配置",
		"get_position":      "获取坐标",
		"get_process_info":  "获取程序信息",
		"ai_settings":       "AI 设置",

		// 对话
		"welcome_message":   "欢迎使用自动化脚本生成助手！请描述您想要实现的自动化任务，我会帮您生成相应的配置。",
		"thinking":          "正在思考...",
		"config_generated":  "已生成配置，请在右侧查看和编辑",
		"input_placeholder": "请描述您想要实现的自动化任务...",

		// 工具
		"get_position_title":    "获取坐标",
		"get_position_desc":     "点击确定后，请在3秒内点击屏幕上的目标位置。\n应用将最小化以便您点击其他窗口。",
		"position_copied":       "已获取坐标 X=%d, Y=%d 并复制到剪贴板",
		"insert_to_editor":      "插入到编辑器",
		"activate_config_title": "使用程序信息",
		"format_json":           "格式化 JSON",
		"json_formatted":        "JSON 已格式化",
		"json_format_error":     "JSON 格式错误: %v",
		"json_format_failed":    "格式化 JSON 失败: %v",

		// 进程信息
		"get_process_info_title":   "获取程序信息",
		"get_foreground_app":       "获取前台应用",
		"foreground_app_title":     "获取前台应用",
		"foreground_app_desc":      "点击确定后，请切换到您想要自动化的应用程序窗口，然后等待1秒钟。",
		"list_all_apps":            "列出所有应用",
		"select_method":            "请选择获取程序信息的方式",
		"get_foreground_desc":      "点击确定后，请切换到您想要自动化的应用程序窗口，然后等待3秒钟。",
		"getting_app_list":         "正在获取运行中的应用程序列表...",
		"no_apps_found":            "未能找到正在运行的应用程序",
		"select_app":               "请选择要自动化的应用程序:",
		"select_config_method":     "请选择配置方式:",
		"use_process_name":         "使用进程名称",
		"use_bundle_id":            "使用Bundle ID",
		"use_app_path":             "使用应用路径",
		"app_name":                 "应用名称: %s",
		"bundle_id":                "Bundle ID: %s",
		"process_id":               "进程 ID: %d",
		"app_path":                 "应用路径: %s",
		"activate_config_copied":   "已生成 activate 配置并复制到剪贴板",
		"getting_window_handle":    "获取窗口句柄",
		"searching_windows":        "正在搜索窗口...",
		"found_window_handle":      "找到窗口句柄: %d",
		"found_multiple_windows":   "找到 %d 个窗口，使用第一个",
		"get_window_handle_failed": "获取窗口句柄失败: %v",

		// 设置
		"settings_title":   "AI 设置",
		"ai_platform":      "AI 平台",
		"api_key":          "API 密钥",
		"model":            "模型",
		"api_endpoint":     "API 端点",
		"api_version":      "API 版本",
		"proxy_url":        "代理 URL",
		"settings_updated": "AI 设置已更新",
		"deepseek_model":   "DeepSeek 模型",
		"deepseek_api_key": "DeepSeek API 密钥",

		// 状态消息
		"no_config":                     "没有可执行的配置",
		"executing":                     "正在执行配置...",
		"execution_failed":              "执行配置失败: %v",
		"execution_complete":            "配置执行完成!",
		"no_savable_config":             "没有可保存的配置",
		"config_saved":                  "配置已保存",
		"last_message_copied":           "最后一条消息已复制到剪贴板",
		"no_copyable_message":           "没有可复制的消息",
		"platform_switch_failed":        "切换平台失败: %v",
		"restart_required":              "语言设置已更改，请重启应用以应用更改",
		"config_refreshed":              "配置文件列表已刷新",
		"config_file":                   "配置文件:",
		"config_dir":                    "配置目录",
		"load_ai_config_failed":         "加载 AI 配置失败: %v",
		"json_editor_placeholder":       "生成的 JSON 配置将显示在这里，您可以直接编辑...",
		"generate_config_failed":        "生成配置失败: %v",
		"you":                           "你",
		"ai":                            "AI",
		"config_loaded":                 "已加载配置: %s",
		"unsupported_ai_platform":       "不支持的 AI 平台类型: %s",
		"read_config_failed":            "读取配置文件失败: %v",
		"parse_config_failed":           "解析配置文件失败: %v",
		"create_config_dir_failed":      "创建配置目录失败: %v",
		"serialize_config_failed":       "序列化配置失败: %v",
		"save_config_failed":            "保存配置失败: %v",
		"serialize_request_failed":      "序列化请求失败: %v",
		"create_request_failed":         "创建请求失败: %v",
		"send_request_failed":           "发送请求失败: %v",
		"read_response_failed":          "读取响应失败: %v",
		"api_error":                     "API 错误: %s",
		"api_request_failed":            "API 请求失败，状态码: %d, 响应: %s",
		"parse_response_failed":         "解析响应失败: %v, 响应内容: %s",
		"response_format_error_choices": "响应格式错误: 未找到 choices 字段或为空, 响应: %s",
		"response_format_error_choice":  "响应格式错误: choices[0] 不是对象, 响应: %s",
		"response_format_error_content": "响应格式错误: 无法提取内容, 响应: %s",
		"response_format_error":         "响应格式错误",
		"no_valid_json":                 "未找到有效的 JSON, 内容: %s",
		"no_valid_json_simple":          "未找到有效的 JSON",

		// tools.go
		"position_captured":    "获取到坐标: X=%d, Y=%d",
		"use_position":         "使用坐标",
		"get_app_list":         "获取应用列表",
		"no_apps_found_title":  "没有找到应用",
		"select_app_desc":      "请选择要自动化的应用程序:",
		"search_app":           "搜索应用...",
		"use_window_handle":    "使用窗口句柄",
		"no_activation_method": "没有可用的激活方法",
		// config.go
		"config_dir_set":                   "配置目录已设置为: %s",
		"load_config_failed":               "加载配置失败: %v",
		"create_temp_file_failed":          "创建临时文件失败: %v",
		"write_config_failed":              "写入配置失败: %v",
		"read_config_file_failed":          "读取配置文件失败: %v",
		"parse_config_file_failed":         "解析配置文件失败: %v",
		"executing_automation_task":        "执行自动化任务: %s",
		"description":                      "描述: %s",
		"executing_action":                 "执行操作 %d: %s",
		"action_execution_failed":          "操作执行失败: %v",
		"warning_single_char_modifiers":    "警告: 对于单个字符 '%s'，无法应用修饰键。使用 TypeStr 替代",
		"condition_result":                 "条件 '%s' 结果: %v",
		"executing_then_branch":            "执行 then 分支",
		"executing_else_branch":            "执行 else 分支",
		"start_loop":                       "开始循环，执行 %d 次",
		"loop_iteration":                   "循环第 %d 次",
		"unknown_action_type":              "未知操作类型: %s",
		"keyboard_operation_failed":        "键盘操作失败: %v",
		"get_process_list_failed":          "获取进程列表失败: %v",
		"process_name_empty":               "进程名称为空",
		"unsupported_os":                   "不支持的操作系统: %s",
		"activate_window_by_handle_failed": "使用窗口句柄激活窗口失败: %v，尝试其他方法",
		"window_activation_executed":       "窗口激活操作已执行",
		"activate_by_bundle_id_failed":     "使用 Bundle ID 激活应用失败: %v，尝试其他方法",
		"app_activation_executed":          "应用激活操作已执行",
		"activate_requires_identifier":     "activate操作需要指定window_handle、process_name、bundle_id或app_path",
		"activate_by_new_method_failed":    "使用新方法激活应用失败: %v，尝试备选方法",
		"found_process":                    "找到进程 %s, PID: %d",
		"process_not_found":                "未找到名为 %s 的进程",
		"window_activated":                 "已激活窗口: %s",
		"bundle_id_empty":                  "Bundle ID为空",
		"bundle_id_mac_only":               "Bundle ID 只在 macOS 上支持",
		"app_activation_failed":            "激活应用失败: %v",
		"bundle_id_app_not_found":          "未找到Bundle ID为 %s 的应用",
		"app_not_found":                    "未找到应用",
		"app_activated_bundle_id":          "已激活应用 (Bundle ID: %s)",
		"app_path_empty":                   "应用路径为空",
		"app_path_not_found":               "未找到路径为 %s 的应用",
		"app_activated_path":               "已激活应用 (路径: %s)",
		"app_launch_failed":                "启动应用失败: %v",
		"app_launched_path":                "已启动应用 (路径: %s)",

		// 图像匹配相关
		"image_file_not_found": "找不到图像文件: %s",
		"failed_to_open_image": "无法打开图像文件: %s",
		"image_found_at":       "在屏幕上找到图像 %s，位置: X=%d, Y=%d",
		"image_not_found":      "在屏幕上未找到图像: %s",

		// AI系统提示
		"system_info_macos":   "当前系统是 macOS，请生成适用于 macOS 的自动化配置。",
		"system_info_windows": "当前系统是 Windows，请生成适用于 Windows 的自动化配置。",
		"ai_system_prompt": `你是一个自动化脚本生成助手。请根据用户的描述，生成符合以下格式的 JSON 配置：
{
  "name": "任务名称",
  "description": "任务描述",
  "actions": [
    {
      "type": "操作类型", // 支持的类型: move, click, type, key, sleep, activate, if, for
      // 其他字段根据操作类型而定
    }
  ]
}

支持的操作类型和字段:
1. move: 移动鼠标 - x, y (坐标)
2. click: 点击鼠标 - button (left/right/center)
3. type: 输入文本 - text (要输入的文本)
4. key: 按键 - key (键名), modifiers (修饰键数组，如 ["control", "shift"])
5. sleep: 等待 - duration (秒)
6. activate: 激活窗口
   - macOS: 支持 process_name (进程名) 或 bundle_id (应用程序包标识符)
   - Windows: 支持 process_name (进程名) 或 window_handle (窗口句柄)
7. if: 条件判断 - condition (条件), then_actions (满足条件时的操作), else_actions (不满足条件时的操作)
8. for: 循环 - count (次数), loop_actions (循环操作)

请只返回 JSON 格式的配置，不要包含其他解释。`,

		// 语言设置
		"language":    "语言",
		"language_zh": "中文",
		"language_en": "English",
	}
}

func createDefaultZHTranslation(path string) error {
	data, err := json.MarshalIndent(getDefaultZHTranslations(), "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// 创建默认英文翻译
func createDefaultENTranslation(path string) error {
	data, err := json.MarshalIndent(getDefaultENTranslations(), "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// 创建默认英文翻译
func getDefaultENTranslations() map[string]string {
	return map[string]string{
		// GUI通用
		"app_title": "Automation Script Generator",
		"save":      "Save",
		"cancel":    "Cancel",
		"close":     "Close",
		"confirm":   "Confirm",

		// 主界面
		"chat_area":         "Chat Area",
		"json_config":       "JSON Config (Editable)",
		"copy_last_message": "Copy Last Message",
		"send":              "Send",
		"execute_config":    "Execute Config",
		"save_config":       "Save Config",
		"get_position":      "Get Position",
		"get_process_info":  "Get Process Info",
		"ai_settings":       "AI Settings",
		"deepseek_model":    "DeepSeek Model",
		"deepseek_api_key":  "DeepSeek API Key",

		// 对话
		"welcome_message":   "Welcome to the Automation Script Generator! Please describe the automation task you want to achieve, and I'll help you generate the corresponding configuration.",
		"thinking":          "Thinking...",
		"config_generated":  "Configuration generated, please check and edit on the right",
		"input_placeholder": "Please describe the automation task you want to achieve...",

		// 工具
		"get_position_title":    "Get Position",
		"get_position_desc":     "After clicking OK, please click on the target position on the screen within 3 seconds.\nThe application will be minimized so you can click on other windows.",
		"position_copied":       "Position X=%d, Y=%d captured and copied to clipboard",
		"insert_to_editor":      "Insert to Editor",
		"activate_config_title": "Use Program Information",
		"format_json":           "Format JSON",
		"json_formatted":        "JSON formatted",
		"json_format_error":     "JSON format error: %v",
		"json_format_failed":    "Failed to format JSON: %v",
		"search_app":            "Search applications...",
		"use_window_handle":     "Use Window Handle",
		"no_activation_method":  "No available activation method",

		// 进程信息
		"get_process_info_title":   "Get Process Info",
		"get_foreground_app":       "Get Foreground App",
		"foreground_app_title":     "Get Foreground App",
		"foreground_app_desc":      "After clicking OK, please switch to the application window you want to automate, then wait for 1 second.",
		"list_all_apps":            "List All Apps",
		"select_method":            "Please select how to get process information",
		"get_foreground_desc":      "After clicking OK, please switch to the application window you want to automate, then wait for 3 seconds.",
		"getting_app_list":         "Getting list of running applications...",
		"no_apps_found":            "No running applications found",
		"select_app":               "Please select the application to automate:",
		"select_config_method":     "Please select configuration method:",
		"use_process_name":         "Use Process Name",
		"use_bundle_id":            "Use Bundle ID",
		"use_app_path":             "Use Application Path",
		"app_name":                 "App Name: %s",
		"bundle_id":                "Bundle ID: %s",
		"process_id":               "Process ID: %d",
		"app_path":                 "App Path: %s",
		"activate_config_copied":   "Activate configuration generated and copied to clipboard",
		"getting_window_handle":    "Getting Window Handle",
		"searching_windows":        "Searching for windows...",
		"found_window_handle":      "Found window handle: %d",
		"found_multiple_windows":   "Found %d windows, using the first one",
		"get_window_handle_failed": "Failed to get window handle: %v",

		// 设置
		"settings_title":                   "AI Settings",
		"ai_platform":                      "AI Platform",
		"api_key":                          "API Key",
		"model":                            "Model",
		"api_endpoint":                     "API Endpoint",
		"api_version":                      "API Version",
		"proxy_url":                        "Proxy URL",
		"settings_updated":                 "AI settings updated",
		"config_loaded":                    "Configuration loaded: %s",
		"read_config_file_failed":          "Failed to read config file: %v",
		"parse_config_file_failed":         "Failed to parse config file: %v",
		"executing_automation_task":        "Executing automation task: %s",
		"description":                      "Description: %s",
		"executing_action":                 "Executing action %d: %s",
		"action_execution_failed":          "Action execution failed: %v",
		"warning_single_char_modifiers":    "Warning: For single character '%s', modifiers cannot be applied. Using TypeStr instead",
		"condition_result":                 "Condition '%s' result: %v",
		"executing_then_branch":            "Executing then branch",
		"executing_else_branch":            "Executing else branch",
		"start_loop":                       "Starting loop, executing %d times",
		"loop_iteration":                   "Loop iteration %d",
		"unknown_action_type":              "Unknown action type: %s",
		"keyboard_operation_failed":        "Keyboard operation failed: %v",
		"get_process_list_failed":          "Failed to get process list: %v",
		"process_name_empty":               "Process name is empty",
		"unsupported_os":                   "Unsupported operating system: %s",
		"activate_window_by_handle_failed": "Failed to activate window by handle: %v, trying other methods",
		"window_activation_executed":       "Window activation executed",
		"activate_by_bundle_id_failed":     "Failed to activate application by Bundle ID: %v, trying other methods",
		"app_activation_executed":          "Application activation executed",
		"activate_requires_identifier":     "Activate operation requires window_handle, process_name, bundle_id, or app_path",
		"activate_by_new_method_failed":    "Failed to activate application by new method: %v, trying fallback method",
		"found_process":                    "Found process %s, PID: %d",
		"process_not_found":                "Process named %s not found",
		"window_activated":                 "Window activated: %s",
		"bundle_id_empty":                  "Bundle ID is empty",
		"bundle_id_mac_only":               "Bundle ID is only supported on macOS",
		"app_activation_failed":            "Failed to activate application: %v",
		"bundle_id_app_not_found":          "Application with Bundle ID %s not found",
		"app_not_found":                    "Application not found",
		"app_activated_bundle_id":          "Application activated (Bundle ID: %s)",
		"app_path_empty":                   "Application path is empty",
		"app_path_not_found":               "Application at path %s not found",
		"app_activated_path":               "Application activated (Path: %s)",
		"app_launch_failed":                "Failed to launch application: %v",
		"app_launched_path":                "Application launched (Path: %s)",

		// 状态消息
		"no_config":                     "No configuration to execute",
		"executing":                     "Executing configuration...",
		"execution_failed":              "Execution failed: %v",
		"execution_complete":            "Configuration execution completed!",
		"no_savable_config":             "No configuration to save",
		"config_saved":                  "Configuration saved",
		"last_message_copied":           "Last message copied to clipboard",
		"no_copyable_message":           "No message to copy",
		"load_ai_config_failed":         "Failed to load AI configuration: %v",
		"json_editor_placeholder":       "Generated JSON configuration will be displayed here, you can edit it directly...",
		"platform_switch_failed":        "Failed to switch platform: %v",
		"save_config_failed":            "Failed to save configuration: %v",
		"restart_required":              "Language settings changed, please restart the application to apply changes",
		"config_refreshed":              "Configuration file list refreshed",
		"config_file":                   "Config File:",
		"config_dir":                    "Config Directory",
		"generate_config_failed":        "Failed to generate configuration: %v",
		"you":                           "You",
		"ai":                            "AI",
		"position_captured":             "Position captured: X=%d, Y=%d",
		"use_position":                  "Use Position",
		"get_app_list":                  "Get Application List",
		"no_apps_found_title":           "No Applications Found",
		"select_app_desc":               "Please select the application to automate:",
		"config_dir_set":                "Configuration directory set to: %s",
		"load_config_failed":            "Failed to load configuration: %v",
		"create_temp_file_failed":       "Failed to create temporary file: %v",
		"write_config_failed":           "Failed to write configuration: %v",
		"unsupported_ai_platform":       "Unsupported AI platform type: %s",
		"read_config_failed":            "Failed to read config file: %v",
		"parse_config_failed":           "Failed to parse config file: %v",
		"create_config_dir_failed":      "Failed to create config directory: %v",
		"serialize_config_failed":       "Failed to serialize config: %v",
		"serialize_request_failed":      "Failed to serialize request: %v",
		"create_request_failed":         "Failed to create request: %v",
		"send_request_failed":           "Failed to send request: %v",
		"read_response_failed":          "Failed to read response: %v",
		"api_error":                     "API error: %s",
		"api_request_failed":            "API request failed, status code: %d, response: %s",
		"parse_response_failed":         "Failed to parse response: %v, response content: %s",
		"response_format_error_choices": "Response format error: choices field not found or empty, response: %s",
		"response_format_error_choice":  "Response format error: choices[0] is not an object, response: %s",
		"response_format_error_content": "Response format error: unable to extract content, response: %s",
		"response_format_error":         "Response format error",
		"no_valid_json":                 "No valid JSON found, content: %s",
		"no_valid_json_simple":          "No valid JSON found",

		// 图像匹配相关
		"image_file_not_found": "Image file not found: %s",
		"failed_to_open_image": "Failed to open image file: %s",
		"image_found_at":       "Image %s found on screen at position: X=%d, Y=%d",
		"image_not_found":      "Image not found on screen: %s",

		// AI系统提示
		"system_info_macos":   "Current system is macOS, please generate automation configuration for macOS.",
		"system_info_windows": "Current system is Windows, please generate automation configuration for Windows.",
		"ai_system_prompt": `You are an automation script generator assistant. Based on the user's description, please generate a JSON configuration in the following format:
{
  "name": "Task Name",
  "description": "Task Description",
  "actions": [
    {
      "type": "action_type", // Supported types: move, click, type, key, sleep, activate, if, for
      // Other fields depend on the action type
    }
  ]
}

Supported action types and fields:
1. move: Move mouse - x, y (coordinates)
2. click: Click mouse - button (left/right/center)
3. type: Type text - text (text to type)
4. key: Press key - key (key name), modifiers (modifier key array, e.g. ["control", "shift"])
5. sleep: Wait - duration (seconds)
6. activate: Activate window
   - macOS: Supports process_name (process name) or bundle_id (application bundle identifier)
   - Windows: Supports process_name (process name) or window_handle (window handle)
7. if: Conditional - condition (condition), then_actions (actions when condition is met), else_actions (actions when condition is not met)
8. for: Loop - count (number of times), loop_actions (loop actions)

Please return only the JSON configuration without any other explanations.`,

		// 语言设置
		"language":    "Language",
		"language_zh": "中文",
		"language_en": "English",
	}
}

// Get current language
func GetCurrentLang() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return currentLang
}

// Set current language
func SetLang(lang string) error {
	if lang != LangZH && lang != LangEN {
		return fmt.Errorf("unsupported language: %s", lang)
	}

	mutex.Lock()
	currentLang = lang
	mutex.Unlock()

	return saveLangSetting(lang)
}

// Save language setting to config file
func saveLangSetting(lang string) error {
	configDir := utils.GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "language.json")
	config := map[string]string{
		"language": lang,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// Load language setting from config file
func init() {
	configDir := utils.GetConfigDir()
	configPath := filepath.Join(configDir, "language.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return // Use default language
	}

	var config map[string]string
	if err := json.Unmarshal(data, &config); err != nil {
		return
	}

	if lang, ok := config["language"]; ok {
		if lang == LangZH || lang == LangEN {
			mutex.Lock()
			currentLang = lang
			mutex.Unlock()
		}
	}
	fmt.Println("currentLang:", string(currentLang))
}

// Get translation text by key
func T(key string) string {
	mutex.RLock()
	defer mutex.RUnlock()

	if trans, ok := translations[currentLang]; ok {
		if text, ok := trans[key]; ok {
			return text
		}
	}

	// Try Chinese if current language translation not found
	if currentLang != LangZH {
		if trans, ok := translations[LangZH]; ok {
			if text, ok := trans[key]; ok {
				return text
			}
		}
	}

	// Try English if Chinese translation not found
	if currentLang != LangEN {
		if trans, ok := translations[LangEN]; ok {
			if text, ok := trans[key]; ok {
				return text
			}
		}
	}

	return key
}

// Get formatted translation text
func Tf(key string, args ...interface{}) string {
	text := T(key)
	result := fmt.Sprintf(text, args...)
	if strings.Contains(result, "%!(EXTRA") || strings.Contains(result, "%!(MISSING") {
		return fmt.Sprintf("%s: %v", text, args)
	}
	return result
}

// GetSystemPrompt returns the system prompt with OS-specific information
func GetSystemPrompt() string {
	var osInfo string
	switch runtime.GOOS {
	case "darwin":
		osInfo = T("system_info_macos")
	case "windows":
		osInfo = T("system_info_windows")
	}

	return osInfo + "\n" + T("ai_system_prompt")
}
