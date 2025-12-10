package errors

import "errors"

var (
	ErrNoMessages     = errors.New("нет сообщений для отправки")
	ErrEmptyInput     = errors.New("пустой ввод")
	ErrClientInit     = errors.New("ошибка инициализации клиента")
	ErrEmptyUserName  = errors.New("имя пользователя не может быть пустым")
	ErrInvalidRole    = errors.New("недопустимая роль")
	ErrEmptyContent   = errors.New("содержимое сообщения не может быть пустым")
	ErrInvalidMessage = errors.New("недопустимое сообщение")
	ErrMessageSend    = errors.New("ошибка при отправке сообщения")
	ErrFileRead       = errors.New("ошибка чтения файла сессии")
	ErrFileParse      = errors.New("ошибка парсинга файла сессии")
	ErrFileSave       = errors.New("ошибка сохранения файла сессии")
	ErrSessionInit    = errors.New("ошибка инициализации сессии")
)
