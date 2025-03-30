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
	"github.com/vector233/Asg/internal/i18n"
)

// initChatInterface initializes the chat interface
func (g *GUI) initChatInterface() {
	// Add welcome message
	g.chatMessages = append(g.chatMessages, ChatMessage{
		Content: i18n.T("welcome_message"),
		IsUser:  false,
		Time:    time.Now(),
	})

	// Create chat display area
	g.chatDisplay = widget.NewRichText()
	g.chatDisplay.Wrapping = fyne.TextWrapWord

	// Status label
	g.statusLabel = widget.NewLabel("")

	// Create message input box
	g.messageInput = widget.NewMultiLineEntry()
	g.messageInput.SetPlaceHolder(i18n.T("input_placeholder"))
	g.messageInput.Wrapping = fyne.TextWrapWord
	g.messageInput.SetMinRowsVisible(3)

	// Update chat display
	g.updateChatDisplay()
}

// createChatContainer creates the chat container
func (g *GUI) createChatContainer() fyne.CanvasObject {
	// Create copy button
	copyButton := widget.NewButtonWithIcon(i18n.T("copy_last_message"), theme.ContentCopyIcon(), func() {
		g.copyLastMessage()
	})

	// Create send button
	sendButton := widget.NewButtonWithIcon(i18n.T("send"), theme.MailSendIcon(), func() {
		g.sendMessage()
	})

	// Create scrollable chat display container
	g.chatScrollContainer = container.NewScroll(g.chatDisplay)

	// Create input container
	inputContainer := container.NewBorder(
		nil, nil, nil, sendButton,
		g.messageInput,
	)

	// Create chat container
	return container.NewBorder(
		container.NewBorder(nil, nil, widget.NewLabel(i18n.T("chat_area")), copyButton, nil),
		inputContainer,
		nil, nil,
		g.chatScrollContainer,
	)
}

// copyLastMessage copies the last message
func (g *GUI) copyLastMessage() {
	if len(g.chatMessages) > 0 {
		lastMsg := g.chatMessages[len(g.chatMessages)-1]
		var content string
		if lastMsg.IsUser {
			content = fmt.Sprintf("You: %s", lastMsg.Content)
		} else {
			content = fmt.Sprintf("AI: %s", lastMsg.Content)
		}

		// Copy to clipboard
		g.window.Clipboard().SetContent(content)
		g.statusLabel.SetText(i18n.T("last_message_copied"))
	} else {
		g.statusLabel.SetText(i18n.T("no_copyable_message"))
	}

	// Clear status message after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		g.statusLabel.SetText("")
	}()
}

// sendMessage sends a message
func (g *GUI) sendMessage() {
	userMessage := g.messageInput.Text
	if userMessage == "" {
		return
	}

	// Add user message to chat
	g.chatMessages = append(g.chatMessages, ChatMessage{
		Content: userMessage,
		IsUser:  true,
		Time:    time.Now(),
	})

	// Clear input box
	g.messageInput.SetText("")

	// Add "thinking" message
	g.chatMessages = append(g.chatMessages, ChatMessage{
		Content: i18n.T("thinking"),
		IsUser:  false,
		Time:    time.Now(),
	})

	g.updateChatDisplay()

	// Generate response in background
	go func() {
		// Generate JSON configuration
		jsonStr, err := g.client.GenerateJSON(userMessage)

		// Update "thinking" message
		lastIndex := len(g.chatMessages) - 1
		if err != nil {
			g.chatMessages[lastIndex] = ChatMessage{
				Content: i18n.Tf("generate_config_failed", err),
				IsUser:  false,
				Time:    time.Now(),
			}
		} else {
			// Format JSON for display
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, []byte(jsonStr), "", "  "); err != nil {
				jsonStr = jsonStr // If formatting fails, use original JSON
			} else {
				jsonStr = prettyJSON.String()
			}

			g.chatMessages[lastIndex] = ChatMessage{
				Content: i18n.T("config_generated"),
				IsUser:  false,
				Time:    time.Now(),
			}

			// Update JSON editor
			g.jsonEditor.SetText(jsonStr)
		}

		g.updateChatDisplay()
	}()
}

// updateChatDisplay updates the chat display
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

	// Scroll to bottom - delay execution to ensure content has been updated
	go func() {
		time.Sleep(100 * time.Millisecond)
		if g.chatScrollContainer != nil {
			g.chatScrollContainer.ScrollToBottom()
		}
	}()
}
