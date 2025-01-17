package models

import "time"

type Team struct {
	ID        int      `gorm:"primaryKey"`
	Name      string   `gorm:"not null"`
	OwnerID   int      `gorm:"not null"`           // Внешний ключ
	Owner     *Player  `gorm:"foreignKey:OwnerID"` // Связь belongs-to
	Players   []Player `gorm:"foreignKey:TeamID"`  // Связь has-many
	IsActive  bool     `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
