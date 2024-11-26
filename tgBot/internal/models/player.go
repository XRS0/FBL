package models

import "time"

// type Player struct {
// 	ID        uint   `gorm:"primaryKey"`
// 	Name      string `gorm:"size:255;not null"`
// 	Height    int    `gorm:"not null"`
// 	Weight    int    `gorm:"not null"`
// 	Position  string `gorm:"size:50;not null"`
// 	Contact   string
// 	TeamID    *uint
// 	CreatedAt time.Time
// }

// type Player struct {
// 	ID        uint   `gorm:"primaryKey"`
// 	Name      string `gorm:"size:255;not null"`
// 	Height    int    `gorm:"not null"`
// 	Weight    int    `gorm:"not null"`
// 	Position  string `gorm:"size:50;not null"`
// 	ChatID    int64  `json:"chat_id"`
// 	TeamID    *uint  `json:"team_id"`
// 	Contact   string
// 	CreatedAt time.Time
// }

type Player struct {
	ID        int    `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Height    int    `gorm:"not null"`
	Weight    int    `gorm:"not null"`
	Position  string `gorm:"not null"`
	ChatID    int64  `gorm:"unique;not null"` // Telegram ID игрока
	Contact   string
	TeamID    *int  `gorm:"index"`                                          // Внешний ключ (может быть NULL)
	Team      *Team `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` // Связь "belongs-to"
	CreatedAt time.Time
	UpdatedAt time.Time
}
