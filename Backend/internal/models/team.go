package models

import (
	"errors"
	"time"
) 

type Team struct {
	ID        int      `gorm:"primaryKey"`
	Name      string   `gorm:"not null"`
	OwnerID   int      `gorm:"not null"`           // Внешний ключ
	Owner     *Player  `gorm:"foreignKey:OwnerID"` // Связь belongs-to
	Players   []Player `gorm:"foreignKey:TeamID"`  // Связь has-many
	IsActive  bool     `gorm:"default:true"`
  PathToLogo string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *Team) RemovePlayerFromTeam(playerNumber uint8) error {
    for i, player := range t.Players {
        if player.Number == playerNumber {
            t.Players = append(t.Players[:i], t.Players[i+1:]...)
            player.TeamID = nil
            return nil
        }
    }
    return errors.New("игрок с указанным номером не найден")
}

func (t *Team) IsOwner(ownerID int) bool {
    return int64(t.OwnerID) == int64(ownerID)
}
