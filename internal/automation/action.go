package automation

// Action 表示一个自动化操作
type Action struct {
	Type         string   `json:"type"`                    // 操作类型：move, click, type, key, sleep, activate, if, for
	X            int      `json:"x,omitempty"`             // 鼠标X坐标
	Y            int      `json:"y,omitempty"`             // 鼠标Y坐标
	Button       string   `json:"button,omitempty"`        // 鼠标按钮：left, right, center
	Text         string   `json:"text,omitempty"`          // 要输入的文本
	Key          string   `json:"key,omitempty"`           // 要按下的键
	Modifiers    []string `json:"modifiers,omitempty"`     // 修饰键：control, shift, alt, command
	Duration     float64  `json:"duration,omitempty"`      // 等待时间（秒）
	ProcessName  string   `json:"process_name,omitempty"`  // 进程名称
	BundleID     string   `json:"bundle_id,omitempty"`     // 应用程序包标识符
	AppPath      string   `json:"app_path,omitempty"`      // 应用程序路径
	WindowHandle int64    `json:"window_handle,omitempty"` // 窗口句柄，用于精确激活窗口

	// 条件判断相关字段
	Condition   string   `json:"condition,omitempty"`    // 条件表达式
	ThenActions []Action `json:"then_actions,omitempty"` // 条件为真时执行的操作
	ElseActions []Action `json:"else_actions,omitempty"` // 条件为假时执行的操作

	// 循环相关字段
	Count       int      `json:"count,omitempty"`        // 循环次数
	LoopActions []Action `json:"loop_actions,omitempty"` // 循环体内的操作
}
