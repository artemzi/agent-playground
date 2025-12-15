package main

import (
	"agent/internal/chat"
	"agent/internal/config"
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	cfg := config.NewConfig()
	if cfg == nil {
		log.Fatal("Ошибка инициализации конфигурации")
	}

	cfg.DisplayConfig()

	userName := getUserName()

	curChat, err := chat.NewChat(userName, cfg)
	if err != nil {
		log.Fatal("Ошибка создания сессии чата:", err)
	}

	fmt.Printf("🤖 Добро пожаловать, %s!\n", userName)

	if len(curChat.GetMessages()) > 0 {
		fmt.Printf("📚 Продолжаем существующий чат (%d сообщений в истории)\n", len(curChat.GetMessages()))
		fmt.Println("\n📜 Последние сообщения:")
		curChat.DisplayRecentMessages(curChat.GetMessages(), 4)
	} else {
		fmt.Println("🆕 Начинаем новый чат")
	}

	fmt.Println("Введите 'exit' или 'quit' для выхода")
	fmt.Println("----------------------------------")

	curChat.StartChat()
}

func getUserName() string {
	fmt.Print("👤 Введите ваше имя: ")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		if scanner.Scan() {
			name := strings.TrimSpace(scanner.Text())

			if name != "" {
				return name
			}
		}
		fmt.Print("❌ Имя не может быть пустым. Попробуйте еще раз: ")
	}
}
