package chat

import (
	"agent/internal/config"
	"agent/internal/errors"
	"agent/internal/model"
	"agent/internal/session"
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

type AIClient interface {
	Generate(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error
}

type Chat struct {
	client  AIClient
	cfg     *config.Config
	session *session.ChatSession
}

func NewChat(userName string, cfg *config.Config) (*Chat, error) {
	if userName == "" {
		return nil, errors.ErrEmptyInput
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrClientInit, err)
	}

	chatSession, err := session.NewChatSession(userName, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrSessionInit, err)
	}

	return &Chat{
		client:  client,
		cfg:     cfg,
		session: chatSession,
	}, nil
}

func (c *Chat) StartChat() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Вы: ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if c.isExitCommand(input) {
			fmt.Println("До свидания! 👋")
			break
		}

		if err := c.processUserInput(input); err != nil {
			fmt.Printf("Ошибка: %v\n", err)
		}
		fmt.Println()
	}
}

const (
	colorGray  = "\033[90m"
	colorReset = "\033[0m"
)

func (c *Chat) sendMessage(message []model.Message) error {
	if len(message) == 0 {
		return errors.ErrNoMessages
	}

	prompt := c.buildContextPrompt(message)

	if c.cfg.UseAssistantPrefill {
		prompt += "\n\nНачни свой ответ с фразы: " + c.cfg.AssistantPrefill
	}

	req := &api.GenerateRequest{
		Think:  c.cfg.ThinkValue,
		Model:  c.cfg.ModelName,
		Prompt: prompt,
		Stream: &[]bool{true}[0],
		System: c.cfg.SystemPrompt,
		Options: map[string]interface{}{
			"temperature": c.cfg.Temperature,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	var response strings.Builder
	var thinkingStarted bool

	err := c.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		if resp.Thinking != "" {
			if !thinkingStarted {
				fmt.Print(colorGray + "💭 ")
				thinkingStarted = true
			}
			fmt.Print(colorGray + resp.Thinking + colorReset)
		}
		if resp.Response != "" {
			fmt.Print(resp.Response)
			response.WriteString(resp.Response)
		}
		return nil
	})

	if thinkingStarted {
		fmt.Print(colorReset + "\n\n")
	}

	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrMessageSend, err)
	}

	c.addAIResponse(response.String())
	c.autoSave()
	return nil
}

func (c *Chat) isExitCommand(input string) bool {
	return input == "exit" || input == "quit" || input == ""
}

func (c *Chat) processUserInput(input string) error {
	userMessage := model.Message{
		Role:      model.RoleUser,
		Content:   input,
		Timestamp: time.Now(),
	}

	c.session.Messages = append(c.session.Messages, userMessage)
	c.session.Updated = time.Now()

	fmt.Print("AI: ")

	err := c.sendMessage(c.session.Messages)
	if err != nil {
		return err
	}
	return nil
}

func (c *Chat) GetMessages() []model.Message {
	return c.session.Messages
}

func (c *Chat) GetSession() *session.ChatSession {
	return c.session
}

func (c *Chat) DisplayRecentMessages(messages []model.Message, count int) {
	start := c.calculateStartIndex(len(messages), count)

	for i := start; i < len(messages); i++ {
		c.displayMessage(messages[i])
	}
	fmt.Println()
}

func (c *Chat) addAIResponse(response string) {
	aiMessage := model.Message{
		Role:      model.RoleAssistant,
		Content:   response,
		Timestamp: time.Now(),
	}
	c.session.Messages = append(c.session.Messages, aiMessage)
	c.session.Updated = time.Now()
}

func (c *Chat) autoSave() {
	msgCount := len(c.session.Messages)
	if msgCount == 2 || msgCount%4 == 0 {
		fmt.Println("\n💾 Автосохранение сессии...")
		if err := c.session.SaveSession(c.session); err != nil {
			fmt.Printf("⚠️  Ошибка автосохранения: %v\n", err)
		}
	}
}

func (c *Chat) calculateStartIndex(totalMessages, count int) int {
	start := totalMessages - count
	if start < 0 {
		start = 0
	}
	return start
}

func (c *Chat) displayMessage(msg model.Message) {
	if msg.IsUser() {
		fmt.Printf("  👤 Вы: %s\n", msg.Content)
	} else {
		content := c.truncateContent(msg.Content, 1000)
		fmt.Printf("  🤖 AI: %s\n", content)
	}
}

func (c *Chat) truncateContent(content string, maxLength int) string {
	if len(content) > maxLength {
		return content[:maxLength] + "..."
	}
	return content
}

func (c *Chat) buildContextPrompt(messages []model.Message) string {
	if len(messages) == 0 {
		return ""
	}

	var builder strings.Builder

	start := c.calculateStartIndex(len(messages), c.cfg.CtxSizeLimit)

	builder.WriteString("Предыдущий контекст беседы:\n")
	for i := start; i < len(messages)-1; i++ { // -1 чтобы исключить текущее сообщение
		msg := messages[i]
		if msg.IsUser() {
			builder.WriteString(fmt.Sprintf("Пользователь: %s\n", msg.Content))
		} else {
			builder.WriteString(fmt.Sprintf("Ассистент: %s\n", msg.Content))
		}
	}

	currentMessage := messages[len(messages)-1]
	builder.WriteString(fmt.Sprintf("\nТекущий вопрос: %s", currentMessage.Content))

	return builder.String()
}
