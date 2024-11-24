package models

import "time"

type MatchStatistics struct {
	ID        uint `gorm:"primaryKey"`
	MatchID   uint `gorm:"not null"`
	TeamID    uint `gorm:"not null"`
	Points    int  `gorm:"default:0"`
	Assists   int  `gorm:"default:0"`
	Rebounds  int  `gorm:"default:0"`
	CreatedAt time.Time
}
