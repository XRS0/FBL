package models

import "time"

// type Match struct {
// 	ID        uint `gorm:"primaryKey"`
// 	Team1ID   uint
// 	Team2ID   uint
// 	Date      time.Time `gorm:"not null"`
// 	Location  string    `gorm:"size:255"`
// 	CreatedAt time.Time
// }

type Match struct {
	ID        uint      `gorm:"primaryKey"`
	Team1ID   uint      `gorm:"not null"`
	Team2ID   uint      `gorm:"not null"`
	Team1     Team      `gorm:"foreignKey:Team1ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Team2     Team      `gorm:"foreignKey:Team2ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Date      time.Time `gorm:"not null"`
	Location  string    `gorm:"size:255"`
	CreatedAt time.Time
}
