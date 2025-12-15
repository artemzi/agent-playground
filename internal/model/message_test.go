package model

import (
	"agent/internal/errors"
	"testing"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name        string
		role        string
		content     string
		wantErr     error
		wantRole    string
		wantContent string
	}{
		{
			name:        "valid user message",
			role:        RoleUser,
			content:     "Hello, world!",
			wantErr:     nil,
			wantRole:    RoleUser,
			wantContent: "Hello, world!",
		},
		{
			name:        "valid assistant message",
			role:        RoleAssistant,
			content:     "Hi there!",
			wantErr:     nil,
			wantRole:    RoleAssistant,
			wantContent: "Hi there!",
		},
		{
			name:    "empty content returns error",
			role:    RoleUser,
			content: "",
			wantErr: errors.ErrEmptyContent,
		},
		{
			name:        "unicode content",
			role:        RoleUser,
			content:     "Привет, мир! 🌍",
			wantErr:     nil,
			wantRole:    RoleUser,
			wantContent: "Привет, мир! 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := NewMessage(tt.role, tt.content)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewMessage() error = %v, wantErr %v", err, tt.wantErr)
				}
				if msg != nil {
					t.Errorf("NewMessage() should return nil on error, got %v", msg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewMessage() unexpected error = %v", err)
				return
			}

			if msg.Role != tt.wantRole {
				t.Errorf("NewMessage() role = %v, want %v", msg.Role, tt.wantRole)
			}

			if msg.Content != tt.wantContent {
				t.Errorf("NewMessage() content = %v, want %v", msg.Content, tt.wantContent)
			}

			if msg.Timestamp.IsZero() {
				t.Error("NewMessage() timestamp should not be zero")
			}
		})
	}
}

func TestMessage_IsUser(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{
			name: "user role returns true",
			role: RoleUser,
			want: true,
		},
		{
			name: "assistant role returns false",
			role: RoleAssistant,
			want: false,
		},
		{
			name: "unknown role returns false",
			role: "unknown",
			want: false,
		},
		{
			name: "empty role returns false",
			role: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{Role: tt.role}
			if got := msg.IsUser(); got != tt.want {
				t.Errorf("Message.IsUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_isAssistant(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{
			name: "assistant role returns true",
			role: RoleAssistant,
			want: true,
		},
		{
			name: "user role returns false",
			role: RoleUser,
			want: false,
		},
		{
			name: "unknown role returns false",
			role: "unknown",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{Role: tt.role}
			if got := msg.isAssistant(); got != tt.want {
				t.Errorf("Message.isAssistant() = %v, want %v", got, tt.want)
			}
		})
	}
}
