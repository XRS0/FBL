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

func UpdatePlayer(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, DB *gorm.DB, userStates map[int64]string, temporaryData map[int64]map[string]string) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	if _, exists := userStates[userID]; !exists {
		userStates[userID] = "update_name"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваше имя:"))
		return
	}

	var existingPlayer models.Player
	err := DB.Where("chat_id = ?", chatID).First(&existingPlayer).Error
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

	// Состояние обновления данных
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

		// Сбор данных для обновления
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

		// Обновление полей игрока, кроме TeamID
		err := DB.Model(&existingPlayer).Updates(updatedPlayer).Error
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при обновлении данных. Попробуйте снова позже."))
			log.Printf("Ошибка при обновлении игрока: %v", err)
			return
		}

		// Сброс временных данных и состояния
		DeleteTemporaryData(userID, temporaryData)
		delete(userStates, userID)

		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Ваши данные успешно обновлены, %s!", fullName)))

	default:
		userStates[userID] = "update_name"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваше новое имя:"))
	}
}
