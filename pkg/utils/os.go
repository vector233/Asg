package utils

import (
	"runtime"
	"strings"
)

// GetCurrentOS returns the current operating system type
func GetCurrentOS() string {
	os := runtime.GOOS
	if strings.Contains(os, "darwin") {
		return "macos"
	} else if strings.Contains(os, "windows") {
		return "windows"
	} else if strings.Contains(os, "linux") {
		return "linux"
	}
	return "unknown"
}