package chat

import (
	"agent/internal/config"
	"agent/internal/errors"
	"agent/internal/model"
	"agent/internal/session"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ollama/ollama/api"
)

type mockAIClient struct {
	generateFunc func(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error
}

func (m *mockAIClient) Generate(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, req, fn)
	}
	return nil
}

func newTestChat(client AIClient, cfg *config.Config) *Chat {
	return &Chat{
		client: client,
		cfg:    cfg,
		session: &session.ChatSession{
			UserName: "testuser",
			Messages: []model.Message{},
			Created:  time.Now(),
			Updated:  time.Now(),
			Cfg:      cfg,
		},
	}
}

func TestChat_isExitCommand(t *testing.T) {
	c := &Chat{}

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"exit command", "exit", true},
		{"quit command", "quit", true},
		{"empty string", "", true},
		{"regular text", "hello", false},
		{"exit with space", " exit", false},
		{"EXIT uppercase", "EXIT", false},
		{"quit with trailing space", "quit ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.isExitCommand(tt.input); got != tt.want {
				t.Errorf("isExitCommand(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestChat_calculateStartIndex(t *testing.T) {
	c := &Chat{}

	tests := []struct {
		name          string
		totalMessages int
		count         int
		want          int
	}{
		{"more messages than count", 10, 4, 6},
		{"equal messages and count", 4, 4, 0},
		{"less messages than count", 2, 4, 0},
		{"zero messages", 0, 4, 0},
		{"zero count", 10, 0, 10},
		{"single message", 1, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.calculateStartIndex(tt.totalMessages, tt.count); got != tt.want {
				t.Errorf("calculateStartIndex(%d, %d) = %v, want %v",
					tt.totalMessages, tt.count, got, tt.want)
			}
		})
	}
}

func TestChat_truncateContent(t *testing.T) {
	c := &Chat{}

	tests := []struct {
		name      string
		content   string
		maxLength int
		want      string
	}{
		{
			name:      "content shorter than max",
			content:   "Hello",
			maxLength: 10,
			want:      "Hello",
		},
		{
			name:      "content equal to max",
			content:   "Hello",
			maxLength: 5,
			want:      "Hello",
		},
		{
			name:      "content longer than max",
			content:   "Hello, World!",
			maxLength: 5,
			want:      "Hello...",
		},
		{
			name:      "empty content",
			content:   "",
			maxLength: 10,
			want:      "",
		},
		{
			name:      "zero max length",
			content:   "Hello",
			maxLength: 0,
			want:      "...",
		},
		{
			name:      "UTF-8 cyrillic text truncation",
			content:   "Привет, мир!",
			maxLength: 6,
			want:      "Привет...",
		},
		{
			name:      "UTF-8 emoji truncation",
			content:   "Hello 🌍🌎🌏 World",
			maxLength: 8,
			want:      "Hello 🌍🌎...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.truncateContent(tt.content, tt.maxLength); got != tt.want {
				t.Errorf("truncateContent(%q, %d) = %q, want %q",
					tt.content, tt.maxLength, got, tt.want)
			}
		})
	}
}

func TestChat_buildContextPrompt(t *testing.T) {
	cfg := &config.Config{
		CtxSizeLimit: 10,
	}
	c := &Chat{cfg: cfg}

	tests := []struct {
		name     string
		messages []model.Message
		contains []string
		excludes []string
	}{
		{
			name:     "empty messages",
			messages: []model.Message{},
			contains: []string{},
		},
		{
			name: "single user message",
			messages: []model.Message{
				{Role: model.RoleUser, Content: "Hello", Timestamp: time.Now()},
			},
			contains: []string{"Текущий вопрос: Hello"},
		},
		{
			name: "conversation with history",
			messages: []model.Message{
				{Role: model.RoleUser, Content: "Hi", Timestamp: time.Now()},
				{Role: model.RoleAssistant, Content: "Hello!", Timestamp: time.Now()},
				{Role: model.RoleUser, Content: "How are you?", Timestamp: time.Now()},
			},
			contains: []string{
				"Предыдущий контекст беседы:",
				"Пользователь: Hi",
				"Ассистент: Hello!",
				"Текущий вопрос: How are you?",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.buildContextPrompt(tt.messages)

			for _, s := range tt.contains {
				if !containsString(got, s) {
					t.Errorf("buildContextPrompt() should contain %q, got %q", s, got)
				}
			}

			for _, s := range tt.excludes {
				if containsString(got, s) {
					t.Errorf("buildContextPrompt() should NOT contain %q, got %q", s, got)
				}
			}
		})
	}
}

func TestChat_buildContextPrompt_respectsLimit(t *testing.T) {
	cfg := &config.Config{
		CtxSizeLimit: 2, // Only last 2 messages in context
	}
	c := &Chat{cfg: cfg}

	messages := []model.Message{
		{Role: model.RoleUser, Content: "First message", Timestamp: time.Now()},
		{Role: model.RoleAssistant, Content: "First reply", Timestamp: time.Now()},
		{Role: model.RoleUser, Content: "Second message", Timestamp: time.Now()},
		{Role: model.RoleAssistant, Content: "Second reply", Timestamp: time.Now()},
		{Role: model.RoleUser, Content: "Current question", Timestamp: time.Now()},
	}

	got := c.buildContextPrompt(messages)

	// Should NOT contain first messages (outside limit)
	if containsString(got, "First message") {
		t.Error("buildContextPrompt() should NOT contain messages outside limit")
	}

	// Should contain recent messages
	if !containsString(got, "Current question") {
		t.Error("buildContextPrompt() should contain current message")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ==================== sendMessage tests ====================

func TestChat_sendMessage_emptyMessages(t *testing.T) {
	cfg := &config.Config{CtxSizeLimit: 10}
	client := &mockAIClient{}
	chat := newTestChat(client, cfg)

	err := chat.sendMessage([]model.Message{})

	if err != errors.ErrNoMessages {
		t.Errorf("sendMessage() with empty messages should return ErrNoMessages, got %v", err)
	}
}

func TestChat_sendMessage_success(t *testing.T) {
	cfg := &config.Config{
		CtxSizeLimit:        10,
		ModelName:           "test-model",
		Temperature:         0.7,
		UseAssistantPrefill: false,
	}

	var capturedReq *api.GenerateRequest
	client := &mockAIClient{
		generateFunc: func(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error {
			capturedReq = req

			// Симулируем стриминг ответа
			fn(api.GenerateResponse{Response: "Hello, "})
			fn(api.GenerateResponse{Response: "world!"})
			return nil
		},
	}

	chat := newTestChat(client, cfg)
	messages := []model.Message{
		{Role: model.RoleUser, Content: "Hi there", Timestamp: time.Now()},
	}

	err := chat.sendMessage(messages)

	if err != nil {
		t.Fatalf("sendMessage() unexpected error: %v", err)
	}

	// Проверяем что запрос сформирован правильно
	if capturedReq == nil {
		t.Fatal("Generate was not called")
	}

	if capturedReq.Model != "test-model" {
		t.Errorf("Request model = %q, want %q", capturedReq.Model, "test-model")
	}

	if !containsString(capturedReq.Prompt, "Hi there") {
		t.Errorf("Request prompt should contain user message, got %q", capturedReq.Prompt)
	}

	// Проверяем что ответ сохранён в сессию
	if len(chat.session.Messages) != 1 {
		t.Fatalf("Expected 1 message in session, got %d", len(chat.session.Messages))
	}

	if chat.session.Messages[0].Content != "Hello, world!" {
		t.Errorf("Saved response = %q, want %q", chat.session.Messages[0].Content, "Hello, world!")
	}

	if chat.session.Messages[0].Role != model.RoleAssistant {
		t.Errorf("Saved message role = %q, want %q", chat.session.Messages[0].Role, model.RoleAssistant)
	}
}

func TestChat_sendMessage_withThinking(t *testing.T) {
	cfg := &config.Config{
		CtxSizeLimit:        10,
		ModelName:           "deepseek-r1:8b",
		UseAssistantPrefill: false,
	}

	client := &mockAIClient{
		generateFunc: func(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error {
			// Симулируем thinking + response
			fn(api.GenerateResponse{Thinking: "Let me think..."})
			fn(api.GenerateResponse{Thinking: " about this."})
			fn(api.GenerateResponse{Response: "Here is my answer."})
			return nil
		},
	}

	chat := newTestChat(client, cfg)
	messages := []model.Message{
		{Role: model.RoleUser, Content: "Complex question", Timestamp: time.Now()},
	}

	err := chat.sendMessage(messages)

	if err != nil {
		t.Fatalf("sendMessage() unexpected error: %v", err)
	}

	// Thinking не должен попасть в сохранённый ответ
	if len(chat.session.Messages) != 1 {
		t.Fatalf("Expected 1 message in session, got %d", len(chat.session.Messages))
	}

	savedContent := chat.session.Messages[0].Content
	if savedContent != "Here is my answer." {
		t.Errorf("Saved response = %q, want %q (thinking should not be included)", savedContent, "Here is my answer.")
	}
}

func TestChat_sendMessage_clientError(t *testing.T) {
	cfg := &config.Config{
		CtxSizeLimit:        10,
		UseAssistantPrefill: false,
	}

	expectedErr := fmt.Errorf("connection refused")
	client := &mockAIClient{
		generateFunc: func(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error {
			return expectedErr
		},
	}

	chat := newTestChat(client, cfg)
	messages := []model.Message{
		{Role: model.RoleUser, Content: "Hello", Timestamp: time.Now()},
	}

	err := chat.sendMessage(messages)

	if err == nil {
		t.Fatal("sendMessage() should return error when client fails")
	}

	// Проверяем что ошибка обёрнута правильно
	if !containsString(err.Error(), "connection refused") {
		t.Errorf("Error should contain original message, got %v", err)
	}

	// Проверяем что ничего не сохранено при ошибке
	if len(chat.session.Messages) != 0 {
		t.Errorf("No messages should be saved on error, got %d", len(chat.session.Messages))
	}
}

func TestChat_sendMessage_withPrefill(t *testing.T) {
	cfg := &config.Config{
		CtxSizeLimit:        10,
		ModelName:           "test-model",
		UseAssistantPrefill: true,
		AssistantPrefill:    "Давайте разберём",
	}

	var capturedPrompt string
	client := &mockAIClient{
		generateFunc: func(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error {
			capturedPrompt = req.Prompt
			fn(api.GenerateResponse{Response: "OK"})
			return nil
		},
	}

	chat := newTestChat(client, cfg)
	messages := []model.Message{
		{Role: model.RoleUser, Content: "Question", Timestamp: time.Now()},
	}

	err := chat.sendMessage(messages)

	if err != nil {
		t.Fatalf("sendMessage() unexpected error: %v", err)
	}

	if !containsString(capturedPrompt, "Давайте разберём") {
		t.Errorf("Prompt should contain prefill instruction, got %q", capturedPrompt)
	}
}

func TestChat_sendMessage_requestOptions(t *testing.T) {
	cfg := &config.Config{
		CtxSizeLimit:    10,
		ModelName:       "llama3",
		Temperature:     0.5,
		StopSequences:   []string{"Human:", "User:"},
		MaxResponseSize: 1024,
		SystemPrompt:    "You are helpful",
	}

	var capturedReq *api.GenerateRequest
	client := &mockAIClient{
		generateFunc: func(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error {
			capturedReq = req
			fn(api.GenerateResponse{Response: "Response"})
			return nil
		},
	}

	chat := newTestChat(client, cfg)
	messages := []model.Message{
		{Role: model.RoleUser, Content: "Test", Timestamp: time.Now()},
	}

	_ = chat.sendMessage(messages)

	if capturedReq.Model != "llama3" {
		t.Errorf("Model = %q, want %q", capturedReq.Model, "llama3")
	}

	if capturedReq.System != "You are helpful" {
		t.Errorf("System = %q, want %q", capturedReq.System, "You are helpful")
	}

	opts := capturedReq.Options
	if temp, ok := opts["temperature"].(float64); !ok || temp != 0.5 {
		t.Errorf("Temperature = %v, want 0.5", opts["temperature"])
	}

	if numPredict, ok := opts["num_predict"].(int); !ok || numPredict != 1024 {
		t.Errorf("num_predict = %v, want 1024", opts["num_predict"])
	}
}
