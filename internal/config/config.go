package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ollama/ollama/api"
)

type Config struct {
	ModelName    string
	Temperature  float64
	ThinkValue   *api.ThinkValue
	CtxDir       string
	CtxSizeLimit int
	CtxFileExt   string
	SystemPrompt string
}

func NewConfig() *Config {
	loadEnvFile(".env")

	config := &Config{
		ModelName:    getEnvString("MODEL_NAME", "deepseek-r1:8b"),
		Temperature:  getEnvFloat("TEMPERATURE", 0.1), // 0 для детерминированных ответов
		ThinkValue:   &api.ThinkValue{Value: getEnvThinkValue("MODEL_THINK_VALUE", false)},
		CtxDir:       getEnvString("CTX_DIR", "chats"),
		CtxSizeLimit: getEnvInt("CTX_SIZE_LIMIT", 10000),
		CtxFileExt:   getEnvString("CTX_FILE_EXT", ".json"),
		SystemPrompt: getEnvString("SYSTEM_PROMPT", "Ты - умный помощник, который помогает пользователю в его задачах."),
	}

	return config
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	fmt.Printf("Переменная окружения %s не установлена, используем значение по умолчанию: %s\n", key, defaultValue)
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}

	fmt.Printf("Переменная окружения %s не установлена или некорректна, используем значение по умолчанию: %.2f\n", key, defaultValue)
	return defaultValue
}

func getEnvThinkValue(key string, defaultValue any) any {
	value := os.Getenv(key)
	if value == "" {
		fmt.Printf("Переменная окружения %s не установлена, используем значение по умолчанию: %v\n", key, defaultValue)
		return defaultValue
	}

	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}

	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}

	fmt.Printf("Переменная окружения %s некорректна (%q), используем значение по умолчанию: %d\n", key, value, defaultValue)
	return defaultValue
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if key, value, ok := strings.Cut(line, "="); ok {
			os.Setenv(strings.TrimSpace(key), strings.TrimSpace(value))
		}
	}
}
