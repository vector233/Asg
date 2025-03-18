package automation

// Config 表示自动化配置
type Config struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Actions     []Action `json:"actions"`
}