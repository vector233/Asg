package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vector233/AsgGPT/internal/automation"
	"github.com/vector233/AsgGPT/internal/ui"
)

func main() {
	// 定义命令行参数
	configFile := flag.String("config", "", "配置文件路径")

	flag.Parse()

	// 如果没有指定配置文件，默认启用 GUI 模式
	if *configFile == "" {
		ui.RunGUI()
		return
	}

	// 执行配置文件
	err := automation.ExecuteConfigFile(*configFile)
	if err != nil {
		fmt.Printf("执行配置失败: %v\n", err)
		os.Exit(1)
	}
}
