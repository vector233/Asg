package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetConfigDir 获取适合当前操作系统的配置目录
func GetConfigDir() string {
	// 首先尝试获取应用程序数据目录
	var appDataDir string

	// 根据不同操作系统获取适当的配置目录
	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\AsgGPT\configs
		appData := os.Getenv("APPDATA")
		if appData != "" {
			appDataDir = filepath.Join(appData, "AsgGPT", "configs")
		}
	case "darwin":
		// macOS: ~/Library/Application Support/AsgGPT/configs
		homeDir, err := os.UserHomeDir()
		if err == nil {
			appDataDir = filepath.Join(homeDir, "Library", "Application Support", "AsgGPT", "configs")
		}
	}

	// 如果无法确定应用数据目录，则使用用户主目录下的 .AsgGPT 目录
	if appDataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			appDataDir = filepath.Join(homeDir, ".AsgGPT", "configs")
		} else {
			// 最后的后备方案：使用当前目录
			appDataDir = filepath.Join(".", "configs")
		}
	}

	return appDataDir
}

// GetExamplesDir 获取示例配置目录
func GetExamplesDir() string {
	// 获取基础配置目录的父目录
	baseDir := filepath.Dir(GetConfigDir())
	// 返回 examples 子目录
	return filepath.Join(baseDir, "examples")
}