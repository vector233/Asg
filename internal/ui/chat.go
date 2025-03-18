package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/vector233/AsgGPT/internal/i18n"
)

// initChatInterface 初始化聊天界面
func (g *GUI) initChatInterface() {
	// 添加欢迎消息
	g.chatMessages = append(g.chatMessages, ChatMessage{
		Content: i18n.T("welcome_message"),
		IsUser:  false,
		Time:    time.Now(),
	})

	// 创建聊天显示区域
	g.chatDisplay = widget.NewRichText()
	g.chatDisplay.Wrapping = fyne.TextWrapWord

	// 状态标签
	g.statusLabel = widget.NewLabel("")

	// 创建消息输入框
	g.messageInput = widget.NewMultiLineEntry()
	g.messageInput.SetPlaceHolder(i18n.T("input_placeholder"))
	g.messageInput.Wrapping = fyne.TextWrapWord
	g.messageInput.SetMinRowsVisible(3)

	// 更新聊天显示
	g.updateChatDisplay()
}

// createChatContainer 创建聊天容器
func (g *GUI) createChatContainer() fyne.CanvasObject {
	// 创建复制按钮
	copyButton := widget.NewButtonWithIcon(i18n.T("copy_last_message"), theme.ContentCopyIcon(), func() {
		g.copyLastMessage()
	})

	// 创建发送按钮
	sendButton := widget.NewButtonWithIcon(i18n.T("send"), theme.MailSendIcon(), func() {
		g.sendMessage()
	})

	// 创建可滚动的聊天显示容器
	g.chatScrollContainer = container.NewScroll(g.chatDisplay)

	// 创建输入容器
	inputContainer := container.NewBorder(
		nil, nil, nil, sendButton,
		g.messageInput,
	)

	// 创建聊天容器
	return container.NewBorder(
		container.NewBorder(nil, nil, widget.NewLabel(i18n.T("chat_area")), copyButton, nil),
		inputContainer,
		nil, nil,
		g.chatScrollContainer,
	)
}

// copyLastMessage 复制最后一条消息
func (g *GUI) copyLastMessage() {
	if len(g.chatMessages) > 0 {
		lastMsg := g.chatMessages[len(g.chatMessages)-1]
		var content string
		if lastMsg.IsUser {
			content = fmt.Sprintf("You: %s", lastMsg.Content)
		} else {
			content = fmt.Sprintf("AI: %s", lastMsg.Content)
		}

		// 复制到剪贴板
		g.window.Clipboard().SetContent(content)
		g.statusLabel.SetText(i18n.T("last_message_copied"))
	} else {
		g.statusLabel.SetText(i18n.T("no_copyable_message"))
	}

	// 2秒后清除状态消息
	go func() {
		time.Sleep(2 * time.Second)
		g.statusLabel.SetText("")
	}()
}

// sendMessage 发送消息
func (g *GUI) sendMessage() {
	userMessage := g.messageInput.Text
	if userMessage == "" {
		return
	}

	// 添加用户消息到聊天
	g.chatMessages = append(g.chatMessages, ChatMessage{
		Content: userMessage,
		IsUser:  true,
		Time:    time.Now(),
	})

	// 清空输入框
	g.messageInput.SetText("")

	// 添加"思考中"消息
	g.chatMessages = append(g.chatMessages, ChatMessage{
		Content: i18n.T("thinking"),
		IsUser:  false,
		Time:    time.Now(),
	})

	g.updateChatDisplay()

	// 在后台生成响应
	go func() {
		// 生成JSON配置
		jsonStr, err := g.client.GenerateJSON(userMessage)

		// 更新"思考中"消息
		lastIndex := len(g.chatMessages) - 1
		if err != nil {
			g.chatMessages[lastIndex] = ChatMessage{
				Content: i18n.Tf("generate_config_failed", err),
				IsUser:  false,
				Time:    time.Now(),
			}
		} else {
			// 格式化JSON以便显示
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, []byte(jsonStr), "", "  "); err != nil {
				jsonStr = jsonStr // 如果格式化失败，使用原始JSON
			} else {
				jsonStr = prettyJSON.String()
			}

			g.chatMessages[lastIndex] = ChatMessage{
				Content: i18n.T("config_generated"),
				IsUser:  false,
				Time:    time.Now(),
			}

			// 更新JSON编辑器
			g.jsonEditor.SetText(jsonStr)
		}

		g.updateChatDisplay()
	}()
}

// updateChatDisplay 更新聊天显示
func (g *GUI) updateChatDisplay() {
	g.chatDisplay.Segments = []widget.RichTextSegment{}
	for _, msg := range g.chatMessages {
		if msg.IsUser {
			segment := &widget.TextSegment{
				Text: fmt.Sprintf("%s: %s\n\n", i18n.T("you"), msg.Content),
				Style: widget.RichTextStyle{
					ColorName: theme.ColorNamePrimary,
					TextStyle: fyne.TextStyle{
						Bold: true,
					},
				},
			}
			g.chatDisplay.Segments = append(g.chatDisplay.Segments, segment)
		} else {
			segment := &widget.TextSegment{
				Text: fmt.Sprintf("%s: %s\n\n", i18n.T("ai"), msg.Content),
				Style: widget.RichTextStyle{
					ColorName: theme.ColorNameForeground,
					TextStyle: fyne.TextStyle{
						Bold: true,
					},
				},
			}
			g.chatDisplay.Segments = append(g.chatDisplay.Segments, segment)
		}
	}

	g.chatDisplay.Refresh()

	// 滚动到底部 - 延迟执行以确保内容已更新
	go func() {
		time.Sleep(100 * time.Millisecond)
		if g.chatScrollContainer != nil {
			g.chatScrollContainer.ScrollToBottom()
		}
	}()
}
