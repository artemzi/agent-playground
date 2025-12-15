package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ollama/ollama/api"
)

type Config struct {
	ModelName           string
	Temperature         float64
	ThinkValue          *api.ThinkValue
	CtxDir              string
	CtxSizeLimit        int
	CtxFileExt          string
	SystemPrompt        string
	AssistantPrefill    string
	UseAssistantPrefill bool
	StopSequences       []string
	MaxResponseSize     int
}

func NewConfig() *Config {
	loadEnvFile(".env")

	config := &Config{
		ModelName:           getEnvString("MODEL_NAME", "deepseek-r1:8b"),
		Temperature:         getEnvFloat("TEMPERATURE", 0.1), // 0 для детерминированных ответов
		ThinkValue:          &api.ThinkValue{Value: getEnvThinkValue("MODEL_THINK_VALUE", false)},
		CtxDir:              getEnvString("CTX_DIR", "chats"),
		CtxSizeLimit:        getEnvInt("CTX_SIZE_LIMIT", 10000),
		CtxFileExt:          getEnvString("CTX_FILE_EXT", ".json"),
		SystemPrompt:        getEnvString("SYSTEM_PROMPT", "Ты - умный помощник, который помогает пользователю в его задачах."),
		AssistantPrefill:    getEnvString("ASSISTANT_PREFILL", "Хорошо, давайте разберем ваш вопрос. "),
		UseAssistantPrefill: getEnvBool("USE_ASSISTANT_PREFILL", true),
		StopSequences:       getEnvStringArray("STOP_SEQUENCES", []string{"Human:", "User:", "Пользователь:"}),
		MaxResponseSize:     getEnvInt("MAX_RESPONSE_SIZE", 0),
	}

	return config
}

func (c *Config) DisplayConfig() {
	fmt.Println("📋 Текущие настройки:")
	fmt.Printf("  🤖 Модель: %s\n", c.ModelName)
	fmt.Printf("  🌡️  Температура: %.1f\n", c.Temperature)
	fmt.Printf("  📁 Директория чатов: %s\n", c.CtxDir)
	fmt.Printf("  📏 Лимит контекста: %d символов\n", c.CtxSizeLimit)
	if c.MaxResponseSize > 0 {
		fmt.Printf("  📐 Лимит ответа: %d символов\n", c.MaxResponseSize)
	} else {
		fmt.Printf("  📐 Лимит ответа: без ограничений\n")
	}
	fmt.Printf("  📄 Расширение файлов: %s\n", c.CtxFileExt)
	fmt.Printf("  🎯 Использовать префилл: %t\n", c.UseAssistantPrefill)
	if c.UseAssistantPrefill {
		fmt.Printf("  💬 Префилл: %s\n", c.AssistantPrefill)
	}
	fmt.Printf("  🛑 Стоп-последовательности: %v\n", c.StopSequences)
	fmt.Println()
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	fmt.Printf("Переменная окружения %s не установлена, используем значение по умолчанию: %s\n", key, defaultValue)
	return defaultValue
}

func getEnvStringArray(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		value = strings.Trim(value, "\"")

		var result []string
		if err := json.Unmarshal([]byte(value), &result); err == nil {
			return result
		}
		fmt.Printf("Переменная окружения %s имеет некорректный JSON формат, используем значение по умолчанию\n", key)
	}
	fmt.Printf("Переменная окружения %s не установлена, используем значение по умолчанию\n", key)
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

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	fmt.Printf("Переменная окружения %s не установлена или некорректна, используем значение по умолчанию: %t\n", key, defaultValue)
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
