package automation

import (
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
	"github.com/vector233/AsgGPT/internal/i18n"
)

// GetMouseClickPosition waits for user click and returns the click position
// timeout specifies the wait duration in seconds, 0 means wait indefinitely
func GetMouseClickPosition(timeout float64) (int, int, error) {
	clickChan := make(chan struct {
		X int
		Y int
	})

	stopChan := make(chan bool, 1)

	go func() {
		evChan := hook.Start()
		defer hook.End()

		for {
			select {
			case ev := <-evChan:
				if ev.Kind == hook.MouseDown {
					x, y := robotgo.Location()
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

	var timeoutChan <-chan time.Time
	if timeout > 0 {
		timeoutChan = time.After(time.Duration(timeout * float64(time.Second)))
	}

	select {
	case pos := <-clickChan:
		stopChan <- true
		return pos.X, pos.Y, nil
	case <-timeoutChan:
		stopChan <- true
		return 0, 0, fmt.Errorf(i18n.T("click_timeout"))
	}
}
