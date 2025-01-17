package models

import "time"

type Player struct {
	ID        int    `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Height    int    `gorm:"not null"`
	Weight    int    `gorm:"not null"`
	Position  string `gorm:"not null"`
	ChatID    int64  `gorm:"unique;not null"` // Telegram ID игрока
	Contact   string
  Number    uint8 
	TeamID    *int  `gorm:"index"`                                          // ID команды (внешний ключ)
	Team      *Team `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` // Связь belongs-to
	CreatedAt time.Time
	UpdatedAt time.Time
}
