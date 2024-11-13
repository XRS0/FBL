// main.go
package main

import (
	"fmt"
	"log"
	"strings"

	. "tgbot/internal"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("7945815181:AAHAzN3QI5dUtq7iSmw9if2rrA5Rzi2j3bY")
	if err != nil {
		log.Panic(err)
	}

	authManager := NewAuthManager()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		switch update.Message.Command() {
		case "start":
			handleStartCommand(bot, update.Message, authManager)
		case "register":
			handleRegisterCommand(bot, update.Message, authManager)
		case "login":
			handleLoginCommand(bot, update.Message, authManager)
		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /start для начала."))
		}
	}
}

func handleStartCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, authManager *AuthManager) {
	telegramID := message.From.ID

	if authManager.IsRegistered(telegramID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Вы уже зарегистрированы. Используйте /login для входа.")
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Привет! Чтобы зарегистрироваться, используйте /register <ваше имя> <пароль>.")
		bot.Send(msg)
	}
}

func handleRegisterCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, authManager *AuthManager) {
	args := strings.Split(message.CommandArguments(), " ")
	if len(args) < 2 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Использование: /register <ваше имя> <пароль>")
		bot.Send(msg)
		return
	}

	name := args[0]
	password := args[1]
	telegramID := message.From.ID

	if authManager.IsRegistered(telegramID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Вы уже зарегистрированы. Используйте /login для входа.")
		bot.Send(msg)
	} else {
		authManager.RegisterUser(telegramID, name, password)
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Регистрация успешна! Добро пожаловать, %s!", name))
		bot.Send(msg)
	}
}

func handleLoginCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, authManager *AuthManager) {
	args := strings.Split(message.CommandArguments(), " ")
	if len(args) < 1 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Использование: /login <пароль>")
		bot.Send(msg)
		return
	}

	name := args[0]
	password := args[1]
	telegramID := message.From.ID

	if authManager.Authenticate(telegramID, name, password) {
		player := authManager.GetPlayer(telegramID)
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Добро пожаловать, %s! Вы вошли в систему.", player.Name))
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка: Неверный пароль или пользователь не зарегистрирован.")
		bot.Send(msg)
	}
}
