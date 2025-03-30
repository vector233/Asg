package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vector233/Asg/internal/automation"
	"github.com/vector233/Asg/internal/ui"
)

func main() {
	// Define command line arguments
	configFile := flag.String("config", "", "Path to configuration file")

	flag.Parse()

	// If no configuration file is specified, start GUI mode by default
	if *configFile == "" {
		ui.RunGUI()
		return
	}

	// Execute configuration file
	err := automation.ExecuteConfigFile(*configFile)
	if err != nil {
		fmt.Printf("Failed to execute configuration: %v\n", err)
		os.Exit(1)
	}
}
