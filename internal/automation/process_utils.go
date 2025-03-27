package automation

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/vector233/AsgGPT/internal/i18n"
)

// ProcessInfo stores process information
type ProcessInfo struct {
	Name         string
	BundleID     string // macOS specific
	Path         string
	PID          int    // Used in Windows
	WindowTitle  string // Window title
	WindowHandle int64  // Window handle for precise activation
}

// GetRunningProcesses returns a list of currently running applications
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

// GetForegroundProcess returns information about the current foreground application
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

// macOS implementation
func getRunningProcessesMac() ([]ProcessInfo, error) {
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

	appNames := strings.Split(strings.TrimSpace(string(output)), ", ")
	result := make([]ProcessInfo, 0, len(appNames))

	for _, appName := range appNames {
		appName = strings.TrimSpace(appName)
		if appName == "" {
			continue
		}

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

	parts := strings.Split(strings.TrimSpace(string(output)), ", ")
	if len(parts) < 3 {
		return ProcessInfo{}, fmt.Errorf(i18n.T("parse_foreground_app_failed"))
	}

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

// Windows implementation
func getRunningProcessesWindows() ([]ProcessInfo, error) {
	cmd := exec.Command("tasklist", "/fo", "csv", "/nh")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf(i18n.T("get_process_list_failed"), err)
	}

	lines := strings.Split(string(output), "\n")
	result := make([]ProcessInfo, 0, len(lines))
	systemProcesses := getSystemProcessBlacklist()
	addedProcesses := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}

		name := strings.Trim(parts[0], "\"")

		// 跳过系统进程和已添加的进程
		if systemProcesses[name] ||
			len(parts) >= 3 && strings.Trim(parts[2], "\"") == "Services" ||
			addedProcesses[name] {
			continue
		}

		pidStr := strings.Trim(parts[1], "\"")
		fmt.Println(name, pidStr)

		var pid int
		fmt.Sscanf(pidStr, "%d", &pid)

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

		addedProcesses[name] = true
		result = append(result, ProcessInfo{
			Name: name,
			Path: path,
			PID:  pid,
		})
	}

	return result, nil
}

// Get foreground window information on Windows
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
		Start-Sleep -Seconds 1
		[WindowInfo]::GetForegroundWindowInfo()
	`

	tempFile, err := os.CreateTemp("", "foreground_*.ps1")
	if err != nil {
		return ProcessInfo{}, fmt.Errorf(i18n.T("create_temp_script_failed"), err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(script); err != nil {
		return ProcessInfo{}, fmt.Errorf(i18n.T("write_script_failed"), err)
	}
	tempFile.Close()

	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-ExecutionPolicy", "Bypass", "-File", tempFile.Name())
	output, err := cmd.Output()
	if err != nil {
		return ProcessInfo{}, fmt.Errorf(i18n.T("get_foreground_window_failed"), err)
	}

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

// Activate window by handle on Windows
func ActivateWindowByHandle(hwnd int64) error {
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
        
        if (-not [WindowActivator]::IsWindow($hwnd)) {
            Write-Output "Invalid:Window handle is invalid"
            exit
        }
        
        $showResult = [WindowActivator]::ShowWindow($hwnd, 9)
        $activateResult = [WindowActivator]::SetForegroundWindow($hwnd)
        
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

	return nil
}

// Returns a map of Windows system processes to be excluded
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

		// System UI and services
		"ShellExperienceHost.exe":   true,
		"SearchHost.exe":            true,
		"RuntimeBroker.exe":         true,
		"SecurityHealthService.exe": true,

		// Terminal related
		"WindowsTerminal.exe": true,
		"powershell.exe":      true,

		// Development tools
		"gopls.exe": true,

		// WSL related
		"wsl.exe":       true,
		"wslhost.exe":   true,
		"vmcompute.exe": true,

		// Audio and system utilities
		"audiodg.exe":  true,
		"tasklist.exe": true,
		"WMIC.exe":     true,
	}
}

// Activate window by name on Windows
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

// Activate application by name on Windows
func ActivateApplicationByName(appName string) error {
	script := fmt.Sprintf(`
        $processes = Get-Process | Where-Object { $_.ProcessName -eq "%s" -or $_.ProcessName -eq "%s" }
        if ($processes.Count -gt 0) {
            $process = $processes[0]
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
                [WindowActivator]::ShowWindow($hwnd, 9)
                [WindowActivator]::SetForegroundWindow($hwnd)
                Write-Output "Success:$($process.Id)"
                exit
            }
            
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
                    [WindowEnumerator]::ShowWindow($hwnd, 9)
                    [WindowEnumerator]::SetForegroundWindow($hwnd)
                    $script:found = $true
                    return $false
                }
                
                return $true
            }
            
            [WindowEnumerator]::EnumWindows($callBack, [IntPtr]::Zero)
            
            if ($script:found) {
                Write-Output "Success:$processId"
            } else {
                Start-Process $process.Path
                Write-Output "Started:$($process.Path)"
            }
        } else {
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

// Activate application by bundle ID on macOS
func ActivateApplicationByBundleID(bundleID string) error {
	if bundleID == "" {
		return fmt.Errorf(i18n.T("bundle_id_empty"))
	}

	script := fmt.Sprintf(`
		set appWasRunning to false
		
		tell application "System Events"
			set appList to application processes whose bundle identifier is "%s"
			if (count of appList) > 0 then
				set appWasRunning to true
			end if
		end tell
		
		if appWasRunning then
			tell application "System Events"
				set frontmost of first application process whose bundle identifier is "%s" to true
			end tell
		else
			try
				tell application id "%s" to activate
			on error errMsg
				try
					set appPath to do shell script "mdfind 'kMDItemCFBundleIdentifier == " & quoted form of "%s" & "' | head -1"
					if appPath is not "" then
						tell application appPath to activate
					else
						return "Error: Application path not found"
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
