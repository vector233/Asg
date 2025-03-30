package utils

import (
	"runtime"
	"strings"
)

// GetCurrentOS 返回当前操作系统类型
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