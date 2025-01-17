package capitanhandlers

import (
  "log"
  "fmt"
  "basketball-league/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	models.Handler
}

func (h *Handler) DeletePlayerFT(bot *tgbotapi.BotAPI, chatID int64) {
  var player models.Player
  if err := h.DB.Where("chat_id = ?", chatID).First(&player).Error; err != nil {
    sendMessage(bot, chatID, "Не удалось найти такого игрока")
  }

  player.TeamID = nil

  if err := h.DB.Save(&player); err != nil {
    sendMessage(bot, chatID, "Произошла ошибка, повторите попытку позже")
  }

  sendMessage(bot, chatID, fmt.Sprintf("Игрок %s исключен из комманды", player.Name))
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to chat %d: %v", chatID, err)
	}
}
