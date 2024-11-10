package main

import (
	"fmt"
	"log"

	. "tgbot/internal"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	bot         *tgbotapi.BotAPI
	authManager *AuthManager
}

func NewTelegramBot(token string) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TelegramBot{
		bot:         bot,
		authManager: NewAuthManager(),
	}, nil
}

func (tb *TelegramBot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tb.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			switch update.Message.Command() {
			case "start":
				tb.authenticateUser(update.Message)
			default:
				tb.handleMessage(update.Message)
			}
		}
	}
}

func (tb *TelegramBot) authenticateUser(message *tgbotapi.Message) {
	telegramID := message.Chat.ID
	name := message.From.FirstName

	if tb.authManager.Authenticate(telegramID) {
		tb.sendWelcomeMessage(message)
	} else {
		tb.authManager.RegisterUser(telegramID, name)
		tb.sendRegistrationMessage(message)
	}
}

func (tb *TelegramBot) sendWelcomeMessage(message *tgbotapi.Message) {
	telegramID := message.Chat.ID
	player := tb.authManager.GetPlayer(telegramID)

	if player != nil {
		msg := tgbotapi.NewMessage(telegramID, "Привет, "+player.Name+"!")
		tb.bot.Send(msg)
	}
}

func (tb *TelegramBot) sendRegistrationMessage(message *tgbotapi.Message) {
	telegramID := message.Chat.ID
	msg := tgbotapi.NewMessage(telegramID, "Добро пожаловать, "+message.From.FirstName+"! Пожалуйста, зарегистрируйтесь.")
	tb.bot.Send(msg)
}

func (tb *TelegramBot) handleMessage(message *tgbotapi.Message) {
	fmt.Printf("Получено сообщение: %s\n", message.Text)
}

func main() {
	bot, err := NewTelegramBot("7945815181:AAHAzN3QI5dUtq7iSmw9if2rrA5Rzi2j3bY")
	if err != nil {
		log.Fatalf("Ошибка при создании бота: %v", err)
	}

	fmt.Println("Бот запущен...")
	bot.Start()
}
