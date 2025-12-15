package session

import (
	"agent/internal/config"
	"agent/internal/model"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSanitizeUserName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "john", "john"},
		{"name with space", "john doe", "john_doe"},
		{"name with slash", "john/doe", "john_doe"},
		{"name with backslash", "john\\doe", "john_doe"},
		{"name with colon", "john:doe", "john_doe"},
		{"name with asterisk", "john*doe", "john_doe"},
		{"name with question mark", "john?doe", "john_doe"},
		{"name with quotes", "john\"doe", "john_doe"},
		{"name with angle brackets", "john<doe>", "john_doe_"},
		{"name with pipe", "john|doe", "john_doe"},
		{"multiple special chars", "a/b\\c:d*e?f\"g<h>i|j", "a_b_c_d_e_f_g_h_i_j"},
		{"empty string", "", ""},
		{"unicode name", "Иван", "Иван"},
		{"unicode with space", "Иван Иванов", "Иван_Иванов"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeUserName(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeUserName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetSessionFilePath(t *testing.T) {
	cfg := &config.Config{
		CtxDir:     "chats",
		CtxFileExt: ".json",
	}

	tests := []struct {
		name     string
		userName string
		expected string
	}{
		{
			name:     "simple name",
			userName: "john",
			expected: filepath.Join("chats", "john.json"),
		},
		{
			name:     "name with space",
			userName: "john doe",
			expected: filepath.Join("chats", "john_doe.json"),
		},
		{
			name:     "name with special chars",
			userName: "john/doe",
			expected: filepath.Join("chats", "john_doe.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSessionFilePath(tt.userName, cfg)
			if got != tt.expected {
				t.Errorf("getSessionFilePath(%q) = %q, want %q", tt.userName, got, tt.expected)
			}
		})
	}
}

func TestNewChatSession_CreatesNew(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()

	cfg := &config.Config{
		CtxDir:     tempDir,
		CtxFileExt: ".json",
	}

	session, err := NewChatSession("testuser", cfg)
	if err != nil {
		t.Fatalf("NewChatSession() error = %v", err)
	}

	if session == nil {
		t.Fatal("NewChatSession() returned nil")
	}

	if session.UserName != "testuser" {
		t.Errorf("session.UserName = %q, want %q", session.UserName, "testuser")
	}

	if len(session.Messages) != 0 {
		t.Errorf("session.Messages should be empty, got %d", len(session.Messages))
	}

	if session.Created.IsZero() {
		t.Error("session.Created should not be zero")
	}

	if session.Cfg != cfg {
		t.Error("session.Cfg should reference the config")
	}
}

func TestChatSession_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		CtxDir:     tempDir,
		CtxFileExt: ".json",
	}

	// Create and save session
	original, err := NewChatSession("testuser", cfg)
	if err != nil {
		t.Fatalf("NewChatSession() error = %v", err)
	}

	original.Messages = []model.Message{
		{Role: model.RoleUser, Content: "Hello", Timestamp: time.Now()},
		{Role: model.RoleAssistant, Content: "Hi there!", Timestamp: time.Now()},
	}
	original.Updated = time.Now()

	err = original.SaveSession(original)
	if err != nil {
		t.Fatalf("SaveSession() error = %v", err)
	}

	// Verify file exists
	filePath := getSessionFilePath("testuser", cfg)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("SaveSession() did not create file")
	}

	// Load session
	loaded, err := NewChatSession("testuser", cfg)
	if err != nil {
		t.Fatalf("Loading session error = %v", err)
	}

	if loaded.UserName != original.UserName {
		t.Errorf("loaded.UserName = %q, want %q", loaded.UserName, original.UserName)
	}

	if len(loaded.Messages) != len(original.Messages) {
		t.Errorf("loaded.Messages length = %d, want %d",
			len(loaded.Messages), len(original.Messages))
	}

	if len(loaded.Messages) > 0 {
		if loaded.Messages[0].Content != original.Messages[0].Content {
			t.Errorf("loaded message content = %q, want %q",
				loaded.Messages[0].Content, original.Messages[0].Content)
		}
	}
}

func TestEnsureChatsDir(t *testing.T) {
	tempDir := t.TempDir()
	newDir := filepath.Join(tempDir, "nested", "chats")

	cfg := &config.Config{
		CtxDir: newDir,
	}

	err := ensureChatsDir(cfg)
	if err != nil {
		t.Fatalf("ensureChatsDir() error = %v", err)
	}

	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Error("ensureChatsDir() did not create directory")
	}
}
