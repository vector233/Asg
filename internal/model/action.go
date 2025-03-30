package model

// Action represents an automation operation
type Action struct {
	Type         string   `json:"type"`                    // Operation type: move, click, type, key, sleep, activate, if, for
	X            int      `json:"x,omitempty"`             // Mouse X coordinate
	Y            int      `json:"y,omitempty"`             // Mouse Y coordinate
	Button       string   `json:"button,omitempty"`        // Mouse button: left, right, center
	Text         string   `json:"text,omitempty"`          // Text to input
	Key          string   `json:"key,omitempty"`           // Key to press
	Modifiers    []string `json:"modifiers,omitempty"`     // Modifier keys: control, shift, alt, command
	Duration     float64  `json:"duration,omitempty"`      // Wait duration in seconds
	ProcessName  string   `json:"process_name,omitempty"`  // Process name
	BundleID     string   `json:"bundle_id,omitempty"`     // Application bundle identifier (macOS)
	AppPath      string   `json:"app_path,omitempty"`      // Application path
	WindowHandle int64    `json:"window_handle,omitempty"` // Window handle for precise activation

	// Conditional fields
	Condition   string   `json:"condition,omitempty"`    // Condition expression
	ThenActions []Action `json:"then_actions,omitempty"` // Actions to execute if condition is true
	ElseActions []Action `json:"else_actions,omitempty"` // Actions to execute if condition is false

	// Loop fields
	Count       int      `json:"count,omitempty"`        // Number of iterations
	LoopActions []Action `json:"loop_actions,omitempty"` // Actions to execute in loop
	ImagePath   string   `json:"imagePath,omitempty"`
}
