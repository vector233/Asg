package automation

import (
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
	"github.com/vector233/AsgGPT/internal/i18n"
)

// GetMouseClickPosition 等待用户点击并返回点击位置
// timeout 是等待超时时间（秒），如果为0则无限等待
func GetMouseClickPosition(timeout float64) (int, int, error) {
	clickChan := make(chan struct {
		X int
		Y int
	})

	// 创建一个停止信号通道，使用带缓冲的通道避免阻塞
	stopChan := make(chan bool, 1)

	// 在后台启动鼠标监听
	go func() {
		evChan := hook.Start()
		defer hook.End()

		for {
			select {
			case ev := <-evChan:
				// 只处理鼠标按下事件
				if ev.Kind == hook.MouseDown {
					x, y := robotgo.Location() // 使用GetMousePos替代Location
					clickChan <- struct {
						X int
						Y int
					}{X: x, Y: y}
				}
			case <-stopChan:
				return
			}
		}
	}()

	fmt.Println(i18n.T("click_to_get_coordinates"))

	// 设置超时
	var timeoutChan <-chan time.Time
	if timeout > 0 {
		timeoutChan = time.After(time.Duration(timeout * float64(time.Second)))
	}

	// 等待点击或超时
	select {
	case pos := <-clickChan:
		stopChan <- true
		return pos.X, pos.Y, nil
	case <-timeoutChan:
		stopChan <- true
		return 0, 0, fmt.Errorf(i18n.T("click_timeout"))
	}
}
