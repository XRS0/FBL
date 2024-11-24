package models

import "time"

type Team struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	OwnerID   uint   // ID владельца команды (ссылается на Player)
	Owner     Player
	IsActive  bool `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// type Team struct {
// 	ID           uint   `gorm:"primaryKey"`
// 	Name         string `gorm:"size:255;not null"`
// 	OwnerContact string `gorm:"size:255;not null"`
// 	IsActive     bool   `gorm:"default:true"`
// 	CreatedAt    time.Time
// 	Players      []Player `gorm:"foreignKey:TeamID"`
// }
