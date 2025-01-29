package ownerhandlers

import (
	"basketball-league/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func (h *Handler) RemovePlayerFromTeamByPlayerNum(chatId int64) {
  var player models.Player
  
  if err := h.DB.Where("chat_id = ?", chatId).First(&player).Error; err != nil {
    h.Bot.Send(tgbotapi.NewMessage(chatId, "Произошла ошибка, попробуйте позже"))  
  }

  if h.IsOwner(chatId, player.Team.Name) {
    player.TeamID = nil
    h.DB.Save(&player)
    return
  } else {
    h.Bot.Send(tgbotapi.NewMessage(chatId, "Вы не являетесь владельцем команды"))
    return
  }
}

func (h *Handler) AddPlayerToTeam(chatId int64) {
  var player models.Player

  if err := h.DB.Where("chat_id = ?", chatId).First(&player).Error; err != nil {
    h.Bot.Send(tgbotapi.NewMessage(chatId, "Произошла ошибка, попробуйте позже"))  
  }
}

