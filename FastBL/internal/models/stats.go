package models

import "time"

type MatchStatistics struct {
	ID         uint `gorm:"primaryKey"`
	MatchID    uint `gorm:"not null"`
	TeamID1    uint `gorm:"not null"`
	TeamID2    uint `gorm:"not null"`
	Team1Score int  `gorm:"default:0"`
	Team2Score int  `gorm:"default:0"`
	CreatedAt  time.Time
}
