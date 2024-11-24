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

type Player struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:255;not null"`
	Height    int    `gorm:"not null"`
	Weight    int    `gorm:"not null"`
	Position  string `gorm:"size:50;not null"`
	ChatID    int64  `json:"chat_id"`
	TeamID    *uint  `json:"team_id"`
	Contact   string
	CreatedAt time.Time
}
