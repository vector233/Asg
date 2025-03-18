package automation

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/vector233/AsgGPT/internal/i18n"
)

// ProcessInfo 存储进程信息
type ProcessInfo struct {
	Name         string
	BundleID     string // macOS 特有
	Path         string
	PID          int    // Windows 可能会用到
	WindowTitle  string // 窗口标题
	WindowHandle int64  // 窗口句柄，用于精确激活
}

// GetRunningProcesses 获取当前运行的应用程序列表
func GetRunningProcesses() ([]ProcessInfo, error) {
	switch runtime.GOOS {
	case "darwin":
		return getRunningProcessesMac()
	case "windows":
		return getRunningProcessesWindows()
	default:
		return nil, fmt.Errorf(i18n.T("unsupported_os"), runtime.GOOS)
	}
}

// GetForegroundProcess 获取当前前台应用程序信息
func GetForegroundProcess() (ProcessInfo, error) {
	switch runtime.GOOS {
	case "darwin":
		return getForegroundProcessMac()
	case "windows":
		return getForegroundProcessWindows()
	default:
		return ProcessInfo{}, fmt.Errorf(i18n.T("unsupported_os"), runtime.GOOS)
	}
}

// macOS 实现
func getRunningProcessesMac() ([]ProcessInfo, error) {
	// 使用AppleScript获取运行中的应用程序信息
	cmd := exec.Command("osascript", "-e", `
		set appList to {}
		tell application "System Events"
			set appProcesses to every process where background only is false
			repeat with appProcess in appProcesses
				set appName to name of appProcess
				set end of appList to appName
			end repeat
		end tell
		return appList
	`)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf(i18n.T("get_app_list_failed"), err)
	}

	// 解析输出
	appNames := strings.Split(strings.TrimSpace(string(output)), ", ")
	result := make([]ProcessInfo, 0, len(appNames))

	// 获取每个应用的详细信息
	for _, appName := range appNames {
		appName = strings.TrimSpace(appName)
		if appName == "" {
			continue
		}

		// 获取应用的Bundle ID
		bundleIDCmd := exec.Command("osascript", "-e", fmt.Sprintf(`
			try
				tell application "System Events"
					return bundle identifier of application process "%s"
				end tell
			on error
				return ""
			end try
		`, appName))

		bundleIDOutput, err := bundleIDCmd.Output()
		bundleID := ""
		if err == nil {
			bundleID = strings.TrimSpace(string(bundleIDOutput))
		}

		// 获取应用的路径
		pathCmd := exec.Command("osascript", "-e", fmt.Sprintf(`
			try
				tell application "System Events"
					return path of application process "%s"
				end tell
			on error
				return ""
			end try
		`, appName))

		pathOutput, err := pathCmd.Output()
		path := ""
		if err == nil {
			path = strings.TrimSpace(string(pathOutput))
		}

		result = append(result, ProcessInfo{
			Name:     appName,
			BundleID: bundleID,
			Path:     path,
		})
	}

	return result, nil
}

func getForegroundProcessMac() (ProcessInfo, error) {
	// 使用AppleScript获取前台应用信息
	cmd := exec.Command("osascript", "-e", `
		tell application "System Events"
			set frontApp to first process whose frontmost is true
			set frontAppName to name of frontApp
			
			try
				set frontAppBundleID to bundle identifier of frontApp
			on error
				set frontAppBundleID to ""
			end try
			
			try
				set frontAppPath to path of frontApp
			on error
				set frontAppPath to ""
			end try
			
			try
				set frontWindowTitle to name of first window of frontApp
			on error
				set frontWindowTitle to ""
			end try
			
			return {frontAppName, frontAppBundleID, frontAppPath, frontWindowTitle}
		end tell
	`)

	output, err := cmd.Output()
	if err != nil {
		return ProcessInfo{}, fmt.Errorf(i18n.T("get_foreground_app_failed"), err)
	}

	// 解析输出
	parts := strings.Split(strings.TrimSpace(string(output)), ", ")
	if len(parts) < 3 {
		return ProcessInfo{}, fmt.Errorf(i18n.T("parse_foreground_app_failed"))
	}

	// 获取窗口标题，如果有的话
	windowTitle := ""
	if len(parts) >= 4 {
		windowTitle = strings.TrimSpace(parts[3])
	}

	return ProcessInfo{
		Name:        strings.TrimSpace(parts[0]),
		BundleID:    strings.TrimSpace(parts[1]),
		Path:        strings.TrimSpace(parts[2]),
		WindowTitle: windowTitle,
	}, nil
}

// Windows 实现
func getRunningProcessesWindows() ([]ProcessInfo, error) {
	// 使用 tasklist 命令获取进程列表
	cmd := exec.Command("tasklist", "/fo", "csv", "/nh")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf(i18n.T("get_process_list_failed"), err)
	}

	// 解析输出
	lines := strings.Split(string(output), "\n")
	result := make([]ProcessInfo, 0, len(lines))

	// 获取系统进程黑名单
	systemProcesses := getSystemProcessBlacklist()

	// 用于跟踪已添加的进程，避免重复
	addedProcesses := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析 CSV 格式
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}

		// 去除引号
		name := strings.Trim(parts[0], "\"")

		// 跳过系统进程
		if systemProcesses[name] {
			continue
		}

		// 跳过服务窗口会话
		if len(parts) >= 3 && strings.Trim(parts[2], "\"") == "Services" {
			continue
		}

		// 如果这个进程名已经添加过，则跳过
		if addedProcesses[name] {
			continue
		}

		pidStr := strings.Trim(parts[1], "\"")
		fmt.Println(name, pidStr)

		var pid int
		fmt.Sscanf(pidStr, "%d", &pid)

		// 获取可执行文件路径
		pathCmd := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%d", pid), "get", "ExecutablePath", "/format:list")
		pathOutput, _ := pathCmd.Output()

		path := ""
		if pathLines := strings.Split(string(pathOutput), "\n"); len(pathLines) > 0 {
			for _, pathLine := range pathLines {
				if strings.HasPrefix(pathLine, "ExecutablePath=") {
					path = strings.TrimPrefix(pathLine, "ExecutablePath=")
					path = strings.TrimSpace(path)
					break
				}
			}
		}
		fmt.Println(path)

		// 如果没有可执行文件路径，可能是系统进程，跳过
		if path == "" {
			continue
		}

		// 标记该进程名已添加
		addedProcesses[name] = true

		result = append(result, ProcessInfo{
			Name: name,
			Path: path,
			PID:  pid,
		})
	}

	return result, nil
}

// 修改 getForegroundProcessWindows 函数，获取窗口句柄和标题
func getForegroundProcessWindows() (ProcessInfo, error) {
	script := `
		Add-Type @"
		using System;
		using System.Runtime.InteropServices;
		using System.Diagnostics;
		using System.Text;
		
		public class WindowInfo {
			[DllImport("user32.dll")]
			public static extern IntPtr GetForegroundWindow();
			
			[DllImport("user32.dll")]
			public static extern int GetWindowThreadProcessId(IntPtr hWnd, out int processId);
			
			[DllImport("user32.dll")]
			public static extern int GetWindowText(IntPtr hWnd, StringBuilder text, int count);
			
			public static string GetForegroundWindowInfo() {
				IntPtr hwnd = GetForegroundWindow();
				int pid = 0;
				GetWindowThreadProcessId(hwnd, out pid);
				
				Process process = null;
				try {
					process = Process.GetProcessById(pid);
				} catch {
					return "Unknown,0,Unknown,Unknown," + hwnd.ToInt64();
				}
				
				StringBuilder title = new StringBuilder(256);
				GetWindowText(hwnd, title, 256);
				
				return string.Format("{0},{1},{2},{3},{4}", 
					process.ProcessName, 
					process.Id, 
					process.MainModule != null ? process.MainModule.FileName : "", 
					title.ToString(),
					hwnd.ToInt64());
			}
		}
"@
		
		# 等待一秒，让用户有时间切换到目标窗口
		Start-Sleep -Seconds 1
		
		# 获取前台窗口信息
		[WindowInfo]::GetForegroundWindowInfo()
	`

	// 创建一个临时文件来存储 PowerShell 脚本
	tempFile, err := os.CreateTemp("", "foreground_*.ps1")
	if err != nil {
		return ProcessInfo{}, fmt.Errorf(i18n.T("create_temp_script_failed"), err)
	}
	defer os.Remove(tempFile.Name())

	// 写入脚本内容
	if _, err := tempFile.WriteString(script); err != nil {
		return ProcessInfo{}, fmt.Errorf(i18n.T("write_script_failed"), err)
	}
	tempFile.Close()

	// 使用 -WindowStyle Hidden 参数运行 PowerShell 脚本，避免 PowerShell 窗口成为前台
	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-ExecutionPolicy", "Bypass", "-File", tempFile.Name())
	output, err := cmd.Output()
	if err != nil {
		return ProcessInfo{}, fmt.Errorf(i18n.T("get_foreground_window_failed"), err)
	}

	// 解析输出
	parts := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(parts) < 5 {
		return ProcessInfo{}, fmt.Errorf(i18n.T("parse_foreground_window_failed"))
	}

	var pid int
	fmt.Sscanf(parts[1], "%d", &pid)

	var hwnd int64
	fmt.Sscanf(parts[4], "%d", &hwnd)

	return ProcessInfo{
		Name:         parts[0],
		Path:         parts[2],
		PID:          pid,
		WindowTitle:  parts[3],
		WindowHandle: hwnd,
	}, nil
}

// 添加一个新函数，通过窗口句柄激活窗口
// 修改 ActivateWindowByHandle 函数
func ActivateWindowByHandle(hwnd int64) error {
	// 脚本内容保持不变
	script := fmt.Sprintf(`
        Add-Type @"
        using System;
        using System.Runtime.InteropServices;
        public class WindowActivator {
            [DllImport("user32.dll")]
            [return: MarshalAs(UnmanagedType.Bool)]
            public static extern bool SetForegroundWindow(IntPtr hWnd);
            
            [DllImport("user32.dll")]
            public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
            
            [DllImport("user32.dll")]
            public static extern bool IsWindow(IntPtr hWnd);
        }
"@
        
        $hwnd = [IntPtr]::new(%d)
        
        # 检查窗口句柄是否有效
        $isValidWindow = [WindowActivator]::IsWindow($hwnd)
        if (-not $isValidWindow) {
            Write-Output "Invalid:窗口句柄无效"
            exit
        }
        
        # 尝试显示窗口
        $showResult = [WindowActivator]::ShowWindow($hwnd, 9) # SW_RESTORE = 9
        
        # 尝试激活窗口
        $activateResult = [WindowActivator]::SetForegroundWindow($hwnd)
        
        # 返回详细结果
        Write-Output "Show:$showResult,Activate:$activateResult"
    `, hwnd)

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(i18n.T("execute_powershell_failed"), err)
	}

	result := strings.TrimSpace(string(output))
	fmt.Printf(i18n.T("window_activation_result")+"\n", result)

	if strings.HasPrefix(result, "Invalid:") {
		return fmt.Errorf(i18n.T("invalid_window_handle"), hwnd)
	}

	// 即使SetForegroundWindow返回false，也不一定意味着激活失败
	// 有时候窗口已经被激活，但函数仍然返回false
	// 所以这里我们不再严格检查返回值

	return nil
}

// Windows 实现
// getSystemProcessBlacklist 返回Windows系统进程黑名单
func getSystemProcessBlacklist() map[string]bool {
	return map[string]bool{
		"System Idle Process": true,
		"System":              true,
		"Secure System":       true,
		"Registry":            true,
		"smss.exe":            true,
		"csrss.exe":           true,
		"wininit.exe":         true,
		"services.exe":        true,
		"LsaIso.exe":          true,
		"lsass.exe":           true,
		"svchost.exe":         true,
		"fontdrvhost.exe":     true,
		"WUDFHost.exe":        true,
		"winlogon.exe":        true,
		"dwm.exe":             true,
		"Memory Compression":  true,
		"conhost.exe":         true,
		"WmiPrvSE.exe":        true,
		"unsecapp.exe":        true,
		"sihost.exe":          true,
		"taskhostw.exe":       true,
		// 添加新的系统进程
		"Widgets.exe":                 true,
		"SearchHost.exe":              true,
		"StartMenuExperienceHost.exe": true,
		"WidgetService.exe":           true,
		"RuntimeBroker.exe":           true,
		"SearchIndexer.exe":           true,
		"UserOOBEBroker.exe":          true,
		"dllhost.exe":                 true,
		"ctfmon.exe":                  true,
		"LockApp.exe":                 true,
		"ChsIME.exe":                  true,
		"SecurityHealthSystray.exe":   true,
		"SecurityHealthService.exe":   true,
		"TextInputHost.exe":           true,
		"SystemSettings.exe":          true,
		"WmiApSrv.exe":                true,
		"PhoneExperienceHost.exe":     true,
		"backgroundTaskHost.exe":      true,

		// Windows Subsystem for Linux related
		"wsl.exe":       true,
		"wslhost.exe":   true,
		"wslrelay.exe":  true,
		"vmcompute.exe": true,
		"vmwp.exe":      true,
		"vmmemWSL":      true,
		"ubuntu.exe":    true, // WSL Ubuntu 发行版

		// 终端相关
		"WindowsTerminal.exe": true,
		"OpenConsole.exe":     true,
		"powershell.exe":      true,

		// Docker related processes
		"com.docker.backend.exe":    true,
		"com.docker.extensions.exe": true,
		"com.docker.dev-envs.exe":   true,
		"com.docker.build.exe":      true,
		// "Docker Desktop.exe":        true,
		// "docker.exe":                true,
		// "clash-win64.exe":           true,

		// Remote Desktop client
		"msrdc.exe": true,

		// 搜索相关进程
		"SearchProtocolHost.exe": true,
		"SearchFilterHost.exe":   true,

		// 系统UI和服务组件
		"ShellExperienceHost.exe": true,
		"ShellHost.exe":           true,
		"smartscreen.exe":         true,
		"CHXSmartScreen.exe":      true,
		"servicehost.exe":         true,
		"uihost.exe":              true,

		// 浏览器组件
		"mc-extn-browserhost.exe": true,
		"browserhost.exe":         true,

		// 系统工具和命令行工具
		"winget.exe":   true, // Windows包管理器
		"WMIC.exe":     true, // Windows管理工具
		"tasklist.exe": true, // 任务列表命令
		"audiodg.exe":  true, // 音频引擎

		// 开发工具相关
		"gopls.exe": true, // Go语言服务器

		// 其他已有的系统进程保持不变
	}
}

// ActivateWindowByName 通过窗口标题激活窗口
func ActivateWindowByName(windowTitle string) error {
	script := fmt.Sprintf(`
        Add-Type @"
        using System;
        using System.Runtime.InteropServices;
        public class WindowActivator {
            [DllImport("user32.dll")]
            [return: MarshalAs(UnmanagedType.Bool)]
            public static extern bool SetForegroundWindow(IntPtr hWnd);
            
            [DllImport("user32.dll")]
            public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
            
            [DllImport("user32.dll", SetLastError = true)]
            public static extern IntPtr FindWindow(string lpClassName, string lpWindowName);
        }
"@
        
        $hwnd = [WindowActivator]::FindWindow($null, "%s")
        if ($hwnd -ne [IntPtr]::Zero) {
            [WindowActivator]::ShowWindow($hwnd, 9) # SW_RESTORE = 9
            [WindowActivator]::SetForegroundWindow($hwnd)
            return $true
        }
        return $false
    `, windowTitle)

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(i18n.T("activate_window_failed"), err)
	}

	result := strings.TrimSpace(string(output))
	if result != "True" {
		return fmt.Errorf(i18n.T("window_not_found"), windowTitle)
	}

	return nil
}

// ActivateApplicationByName 通过应用程序名称激活应用
func ActivateApplicationByName(appName string) error {
	// 使用进程ID直接激活窗口，而不是通过窗口标题
	script := fmt.Sprintf(`
        $processes = Get-Process | Where-Object { $_.ProcessName -eq "%s" -or $_.ProcessName -eq "%s" }
        if ($processes.Count -gt 0) {
            $process = $processes[0]  # 获取第一个匹配的进程
            $hwnd = $process.MainWindowHandle
            
            Add-Type @"
            using System;
            using System.Runtime.InteropServices;
            public class WindowActivator {
                [DllImport("user32.dll")]
                [return: MarshalAs(UnmanagedType.Bool)]
                public static extern bool SetForegroundWindow(IntPtr hWnd);
                
                [DllImport("user32.dll")]
                public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
            }
"@
            
            if ($hwnd -ne [IntPtr]::Zero) {
                [WindowActivator]::ShowWindow($hwnd, 9)  # SW_RESTORE = 9
                [WindowActivator]::SetForegroundWindow($hwnd)
                Write-Output "Success:$($process.Id)"
                exit
            }
            
            # 如果主窗口句柄为零，尝试枚举所有窗口
            Add-Type @"
            using System;
            using System.Runtime.InteropServices;
            using System.Text;
            public class WindowEnumerator {
                [DllImport("user32.dll")]
                public static extern bool EnumWindows(EnumWindowsProc enumProc, IntPtr lParam);
                
                [DllImport("user32.dll")]
                public static extern int GetWindowThreadProcessId(IntPtr hWnd, out int processId);
                
                [DllImport("user32.dll")]
                public static extern bool IsWindowVisible(IntPtr hWnd);
                
                [DllImport("user32.dll")]
                public static extern bool SetForegroundWindow(IntPtr hWnd);
                
                [DllImport("user32.dll")]
                public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
                
                public delegate bool EnumWindowsProc(IntPtr hWnd, IntPtr lParam);
            }
"@
            
            $processId = $process.Id
            $found = $false
            
            [WindowEnumerator+EnumWindowsProc]$callBack = {
                param([IntPtr]$hwnd, [IntPtr]$lParam)
                
                $pid = 0
                [void][WindowEnumerator]::GetWindowThreadProcessId($hwnd, [ref]$pid)
                
                if ($pid -eq $processId -and [WindowEnumerator]::IsWindowVisible($hwnd)) {
                    [WindowEnumerator]::ShowWindow($hwnd, 9)  # SW_RESTORE = 9
                    [WindowEnumerator]::SetForegroundWindow($hwnd)
                    $script:found = $true
                    return $false  # 停止枚举
                }
                
                return $true  # 继续枚举
            }
            
            [WindowEnumerator]::EnumWindows($callBack, [IntPtr]::Zero)
            
            if ($script:found) {
                Write-Output "Success:$processId"
            } else {
                # 如果还是找不到可见窗口，尝试启动新实例
                Start-Process $process.Path
                Write-Output "Started:$($process.Path)"
            }
        } else {
            # 如果找不到进程，尝试启动应用程序
            try {
                Start-Process "%s"
                Write-Output "Started:%s"
            } catch {
                Write-Output "Failed:$($_.Exception.Message)"
            }
        }
    `, appName, strings.TrimSuffix(appName, ".exe"), appName, appName)

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(i18n.T("activate_app_failed"), err)
	}

	result := strings.TrimSpace(string(output))
	fmt.Printf(i18n.T("app_activation_result")+"\n", result)

	if strings.HasPrefix(result, "Failed:") {
		return fmt.Errorf(i18n.T("app_activation_error"), strings.TrimPrefix(result, "Failed:"))
	}

	return nil
}

// ActivateApplicationByBundleID 通过 Bundle ID 激活 macOS 应用
func ActivateApplicationByBundleID(bundleID string) error {
	if bundleID == "" {
		return fmt.Errorf(i18n.T("bundle_id_empty"))
	}

	// 使用 AppleScript 通过 Bundle ID 激活应用
	script := fmt.Sprintf(`
		set appWasRunning to false
		
		-- 检查应用是否已经运行
		tell application "System Events"
			set appList to application processes whose bundle identifier is "%s"
			if (count of appList) > 0 then
				set appWasRunning to true
			end if
		end tell
		
		-- 如果应用已经运行，直接激活它
		if appWasRunning then
			tell application "System Events"
				set frontmost of first application process whose bundle identifier is "%s" to true
			end tell
		else
			-- 如果应用没有运行，尝试通过 Bundle ID 启动它
			try
				tell application id "%s" to activate
			on error errMsg
				-- 如果通过 Bundle ID 启动失败，尝试查找应用路径
				try
					set appPath to do shell script "mdfind 'kMDItemCFBundleIdentifier == " & quoted form of "%s" & "' | head -1"
					if appPath is not "" then
						tell application appPath to activate
					else
						return "Error: 找不到应用路径"
					end if
				on error pathErrMsg
					return "Error: " & pathErrMsg
				end try
			end try
		end if
		
		return "Success"
	`, bundleID, bundleID, bundleID, bundleID)

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(i18n.T("activate_app_failed"), err)
	}

	result := strings.TrimSpace(string(output))
	if strings.HasPrefix(result, "Error:") {
		return fmt.Errorf(i18n.T("activate_app_error"), strings.TrimPrefix(result, "Error: "))
	}

	return nil
}
