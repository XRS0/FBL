package ownerhandlers

import (
	"basketball-league/internal/models"
	"errors"
	"fmt"
	"log"
	"strings"
  "strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"gorm.io/gorm"
)

type Handler struct {
  models.Handler
}

func (h *Handler) IsOwner(chatId int64, teamName string) bool {
  var player models.Player
  
  if err := h.DB.Where("chat_id = ?", chatId).First(&player).Error; err != nil {
    h.Bot.Send(tgbotapi.NewMessage(chatId, "Произошла ошибка, попробуйте позже"))  
  }

  var team models.Team

  if err := h.DB.Where("name = ?", teamName).First(&team).Error; err != nil {
    h.Bot.Send(tgbotapi.NewMessage(chatId, "Произошла ошибка, попробуйте позже"))  
  }

  return team.OwnerID == player.ID
}

func (h *Handler) RemovePlayerFromTeam(chatId int64) {
  var player models.Player
  
  if err := h.DB.Where("chat_id = ?", chatId).First(&player).Error; err != nil {
    h.Bot.Send(tgbotapi.NewMessage(chatId, "Произошла ошибка, попробуйте позже"))  
  }

  if h.IsOwner(chatId, player.Team.Name) {
    h.RemovePlayerFromTeam(chatId)
  } 
}
