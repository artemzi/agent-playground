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
	Chat(ctx context.Context, req *api.ChatRequest, fn api.ChatResponseFunc) error
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

	var chatMessages []api.Message

	// Добавляем системный промпт
	if c.cfg.SystemPrompt != "" {
		chatMessages = append(chatMessages, api.Message{
			Role:    "system",
			Content: c.cfg.SystemPrompt,
		})
	}

	// Добавляем историю сообщений с учетом лимита
	start := c.calculateStartIndex(len(message), c.cfg.CtxSizeLimit)
	for i := start; i < len(message); i++ {
		msg := message[i]
		chatMessages = append(chatMessages, api.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	req := &api.ChatRequest{
		Model:    c.cfg.ModelName,
		Messages: chatMessages,
		Stream:   &[]bool{true}[0],
		Options: map[string]interface{}{
			"temperature": c.cfg.Temperature,
		},
	}

	// Увеличили таймаут до 5 минут
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	var response strings.Builder

	err := c.client.Chat(ctx, req, func(resp api.ChatResponse) error {
		content := resp.Message.Content
		if content != "" {
			fmt.Print(content)
			response.WriteString(content)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrMessageSend, err)
	}
	fmt.Println()

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
