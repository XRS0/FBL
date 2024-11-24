package models

import "time"

type Match struct {
	ID        uint `gorm:"primaryKey"`
	Team1ID   uint
	Team2ID   uint
	Date      time.Time `gorm:"not null"`
	Location  string    `gorm:"size:255"`
	CreatedAt time.Time
}
