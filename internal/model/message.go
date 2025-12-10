package model

import (
	"agent/internal/errors"
	"time"
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func NewMessage(role, content string) (*Message, error) {
	if content == "" {
		return nil, errors.ErrEmptyContent
	}

	return &Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}, nil
}

func (m *Message) IsUser() bool {
	return m.Role == RoleUser
}

func (m *Message) isAssistant() bool {
	return m.Role == RoleAssistant
}
