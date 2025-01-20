package models

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
  "gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
  Bot *tgbotapi.BotAPI
}
