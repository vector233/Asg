package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetConfigDir returns the appropriate configuration directory for the current operating system
func GetConfigDir() string {
	var appDataDir string

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

	if appDataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			appDataDir = filepath.Join(homeDir, ".AsgGPT", "configs")
		} else {
			appDataDir = filepath.Join(".", "configs")
		}
	}

	return appDataDir
}

// GetExamplesDir returns the examples configuration directory
func GetExamplesDir() string {
	return filepath.Join(filepath.Dir(GetConfigDir()), "examples")
}