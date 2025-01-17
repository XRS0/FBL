package userhandlers

import (
	"basketball-league/internal/models"
	. "basketball-league/internal/tempDataHandlers"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"gorm.io/gorm"
)

type Handler struct {
	models.Handler
}

func (h *Handler) UpdatePlayer(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, userStates map[int64]string, temporaryData map[int64]map[string]string) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	if _, exists := userStates[userID]; !exists {
		userStates[userID] = "update_name"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваше имя:"))
		return
	}

	var existingPlayer models.Player
	err := h.DB.Where("chat_id = ?", chatID).First(&existingPlayer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			bot.Send(tgbotapi.NewMessage(chatID, "Вы не зарегистрированы в системе. Сначала пройдите регистрацию."))
			delete(userStates, userID)
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при поиске ваших данных. Попробуйте позже."))
			log.Printf("Ошибка при поиске игрока: %v", err)
		}
		return
	}

	state := userStates[userID]
	switch state {
	case "update_name":
		if len(msg.Text) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Имя должно быть длиннее 1 символа. Попробуйте снова."))
			return
		}
		SetTemporaryData(userID, "name", msg.Text, temporaryData)
		userStates[userID] = "update_patronymic"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваше отчество:"))

	case "update_patronymic":
		if len(msg.Text) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Отчество должно быть длиннее 1 символа. Попробуйте снова."))
			return
		}
		SetTemporaryData(userID, "patronymic", msg.Text, temporaryData)
		userStates[userID] = "update_surname"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите вашу фамилию:"))

	case "update_surname":
		if len(msg.Text) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Фамилия должна быть длиннее 1 символа. Попробуйте снова."))
			return
		}
		SetTemporaryData(userID, "surname", msg.Text, temporaryData)
		userStates[userID] = "update_height"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваш новый рост (см):"))

	case "update_height":
		height, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || height < 100 || height > 250 {
			bot.Send(tgbotapi.NewMessage(chatID, "Укажите корректный рост в сантиметрах (от 100 до 250)."))
			return
		}
		SetTemporaryData(userID, "height", strconv.Itoa(height), temporaryData)
		userStates[userID] = "update_weight"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваш новый вес (кг):"))

	case "update_weight":
		weight, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || weight < 30 || weight > 200 {
			bot.Send(tgbotapi.NewMessage(chatID, "Укажите корректный вес в килограммах (от 30 до 200)."))
			return
		}
		SetTemporaryData(userID, "weight", strconv.Itoa(weight), temporaryData)
		userStates[userID] = "update_position"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите вашу новую игровую позицию:"))

	case "update_position":
		if len(msg.Text) < 3 {
			bot.Send(tgbotapi.NewMessage(chatID, "Позиция должна содержать хотя бы 3 символа. Попробуйте снова."))
			return
		}
		SetTemporaryData(userID, "position", msg.Text, temporaryData)

		tempData := GetTemporaryData(userID, temporaryData)
		fullName := fmt.Sprintf("%s %s %s", tempData["name"], tempData["patronymic"], tempData["surname"])
		height, _ := strconv.Atoi(tempData["height"])
		weight, _ := strconv.Atoi(tempData["weight"])

		updatedPlayer := models.Player{
			Name:     fullName,
			Height:   height,
			Weight:   weight,
			Position: tempData["position"],
		}

		err := h.DB.Model(&existingPlayer).Updates(updatedPlayer).Error
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при обновлении данных. Попробуйте снова позже."))
			log.Printf("Ошибка при обновлении игрока: %v", err)
			return
		}

		DeleteTemporaryData(userID, temporaryData)
		delete(userStates, userID)

		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Ваши данные успешно обновлены, %s!", fullName)))

	default:
		userStates[userID] = "update_name"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваше новое имя:"))
	}
}

func (h *Handler) RegisterPlayer(bot *tgbotapi.BotAPI, temporaryData map[int64]map[string]string, userStates map[int64]string, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	if _, exists := userStates[userID]; !exists {
		userStates[userID] = "register_name"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваше имя:"))
		return
	}

	var existingPlayer models.Player
	err := h.DB.Where("chat_id = ?", chatID).First(&existingPlayer).Error
	if err == nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы уже зарегистрированы в системе!"))
		userStates[userID] = ""
		return
	}

	// Состояние регистрации
	state := userStates[userID]
	switch state {
	case "register_name":
		if len(msg.Text) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Имя должно быть длиннее 1 символа. Попробуйте снова."))
			return
		}
		SetTemporaryData(userID, "first_name", msg.Text, temporaryData)
		userStates[userID] = "register_patronymic"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваше отчество:"))

	case "register_patronymic":
		if len(msg.Text) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Отчество должно быть длиннее 1 символа. Попробуйте снова."))
			return
		}
		SetTemporaryData(userID, "patronymic", msg.Text, temporaryData)
		userStates[userID] = "register_last_name"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите вашу фамилию:"))

	case "register_last_name":
		if len(msg.Text) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Фамилия должна быть длиннее 1 символа. Попробуйте снова."))
			return
		}
		SetTemporaryData(userID, "last_name", msg.Text, temporaryData)

		tempData := GetTemporaryData(userID, temporaryData)
		fullName := fmt.Sprintf("%s %s %s", tempData["first_name"], tempData["patronymic"], tempData["last_name"])
		SetTemporaryData(userID, "name", fullName, temporaryData)

		userStates[userID] = "register_height"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваш рост (см):"))

	case "register_height":
		height, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || height < 100 || height > 250 {
			bot.Send(tgbotapi.NewMessage(chatID, "Укажите корректный рост в сантиметрах (от 100 до 250)."))
			return
		}
		SetTemporaryData(userID, "height", strconv.Itoa(height), temporaryData)
		userStates[userID] = "register_weight"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваш вес (кг):"))

	case "register_weight":
		weight, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || weight < 30 || weight > 200 {
			bot.Send(tgbotapi.NewMessage(chatID, "Укажите корректный вес в килограммах (от 30 до 200)."))
			return
		}
		SetTemporaryData(userID, "weight", strconv.Itoa(weight), temporaryData)
		userStates[userID] = "register_position"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите вашу игровую позицию (например, Центровой, Разыгрывающий):"))

	case "register_position":
		if len(msg.Text) < 3 {
			bot.Send(tgbotapi.NewMessage(chatID, "Позиция должна содержать хотя бы 3 символа. Попробуйте снова."))
			return
		}
		SetTemporaryData(userID, "position", msg.Text, temporaryData)
		userStates[userID] = "register_contact"

		// Клавиатура для отправки контакта
		contactButton := tgbotapi.NewKeyboardButtonContact("Поделиться контактом")
		replyMarkup := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(contactButton),
		)
		message := tgbotapi.NewMessage(chatID, "Поделитесь вашим контактом через кнопку ниже:")
		message.ReplyMarkup = replyMarkup
		bot.Send(message)

	case "register_contact":
		if msg.Contact == nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, используйте кнопку 'Поделиться контактом'."))
			return
		}

		SetTemporaryData(userID, "contact", fmt.Sprintf("Номер телефона - %s\nTgID - @%s", msg.Contact.PhoneNumber, msg.From.UserName), temporaryData)

		tempData := GetTemporaryData(userID, temporaryData)
		height, _ := strconv.Atoi(tempData["height"])
		weight, _ := strconv.Atoi(tempData["weight"])

		player := models.Player{
			Name:     tempData["name"],
			Height:   height,
			Weight:   weight,
			Position: tempData["position"],
			Contact:  tempData["contact"],
			ChatID:   chatID,
		}

		err := h.DB.Create(&player).Error
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при регистрации. Попробуйте снова."))
			log.Printf("Ошибка при регистрации игрока: %v", err)
			return
		}

		DeleteTemporaryData(userID, temporaryData)
		delete(userStates, userID)

		msgClearKeyboard := tgbotapi.NewMessage(chatID, fmt.Sprintf("Игрок %s успешно зарегистрирован!", player.Name))
		msgClearKeyboard.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		bot.Send(msgClearKeyboard)

	default:
		userStates[userID] = "register_name"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваше имя:"))
	}
}

func (h *Handler) ListProfile(bot *tgbotapi.BotAPI, chatID int64) {
	var player models.Player

	err := h.DB.Where("chat_id = ?", chatID).First(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы не зарегистрированы. Используйте команду /register."))
		return
	}

  message := fmt.Sprintf("Имя: %s\nРост: %d см\nВес: %d кг\nПозиция: %s\nКонтакты: %s\nНомер игрока: %d",
		player.Name, player.Height, player.Weight, player.Position, player.Contact, player.Number)
	bot.Send(tgbotapi.NewMessage(chatID, message))
}

func (h *Handler) Logout(bot *tgbotapi.BotAPI, temporaryData map[int64]map[string]string, userStates map[int64]string, chatID int64, userID int64) {
	err := h.DB.Where("chat_id = ?", chatID).Delete(&models.Player{}).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при выходе из аккаунта. Попробуйте снова."))
		log.Printf("Ошибка при удалении аккаунта: %v", err)
		return
	}

	delete(userStates, userID)
	DeleteTemporaryData(userID, temporaryData)

	bot.Send(tgbotapi.NewMessage(chatID, "Вы успешно вышли из аккаунта. Для повторной регистрации используйте команду /register."))
}
