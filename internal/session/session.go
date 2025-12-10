package session

import (
	"agent/internal/config"
	"agent/internal/errors"
	"agent/internal/model"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ChatSession struct {
	UserName string          `json:"username"`
	Messages []model.Message `json:"messages"`
	Created  time.Time       `json:"created"`
	Updated  time.Time       `json:"updated"`
	Cfg      *config.Config  `json:"-"`
}

func NewChatSession(userName string, cfg *config.Config) (*ChatSession, error) {
	if err := ensureChatsDir(cfg); err != nil {
		return nil, fmt.Errorf("создание директории чатов: %w", err)
	}

	return loadOrCreateSession(userName, cfg)
}

func (c *ChatSession) SaveSession(session *ChatSession) error {
	filePath := getSessionFilePath(session.UserName, c.Cfg)

	data, err := json.MarshalIndent(session, "", " ")
	if err != nil {
		return fmt.Errorf("%w: ошибка сериализации: %v", errors.ErrFileSave, err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("%w: ошибка записи: %v", errors.ErrFileSave, err)
	}

	return nil
}

func ensureChatsDir(cfg *config.Config) error {
	return os.MkdirAll(cfg.CtxDir, os.ModePerm)
}

func loadOrCreateSession(userName string, cfg *config.Config) (*ChatSession, error) {
	filePath := getSessionFilePath(userName, cfg)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &ChatSession{
			UserName: userName,
			Messages: make([]model.Message, 0),
			Created:  time.Now(),
			Updated:  time.Now(),
			Cfg:      cfg,
		}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrFileRead, err)
	}

	var session ChatSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrFileParse, err)
	}

	session.Cfg = cfg
	return &session, nil
}

func getSessionFilePath(userName string, cfg *config.Config) string {
	safeUserName := sanitizeUserName(userName)
	return filepath.Join(cfg.CtxDir, fmt.Sprintf("%s%s", safeUserName, cfg.CtxFileExt))
}

func sanitizeUserName(userName string) string {
	safeUserName := strings.ReplaceAll(userName, " ", "_")
	safeUserName = strings.ReplaceAll(safeUserName, "/", "_")
	safeUserName = strings.ReplaceAll(safeUserName, "\\", "_")
	safeUserName = strings.ReplaceAll(safeUserName, ":", "_")
	safeUserName = strings.ReplaceAll(safeUserName, "*", "_")
	safeUserName = strings.ReplaceAll(safeUserName, "?", "_")
	safeUserName = strings.ReplaceAll(safeUserName, "\"", "_")
	safeUserName = strings.ReplaceAll(safeUserName, "<", "_")
	safeUserName = strings.ReplaceAll(safeUserName, ">", "_")
	safeUserName = strings.ReplaceAll(safeUserName, "|", "_")
	return safeUserName
}
