package matchhandlers

import (
	"basketball-league/internal/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

func GetMatchByID(db *gorm.DB, matchID int) (*models.Match, error) {
	var match models.Match
	err := db.Preload("Team1.Players").Preload("Team2.Players").First(&match, matchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("матч не найден")
	} else if err != nil {
		return nil, err
	}
	return &match, nil
}

// GetAllMatches получает список всех матчей
func GetAllMatches(db *gorm.DB) ([]models.Match, error) {
	var matches []models.Match
	err := db.Preload("Team1").Preload("Team2").Find(&matches).Error
	if err != nil {
		return nil, err
	}
	return matches, nil
}

// UpdateMatch обновляет информацию о матче
func UpdateMatch(db *gorm.DB, matchID int, team1ID, team2ID uint, date time.Time, location string) error {
	var match models.Match
	err := db.First(&match, matchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("матч не найден")
	} else if err != nil {
		return err
	}

	// Обновление данных
	match.Team1ID = team1ID
	match.Team2ID = team2ID
	match.Date = date
	match.Location = location

	return db.Save(&match).Error
}

func CreateMatch(db *gorm.DB, team1ID, team2ID uint, date time.Time, location string) (*models.Match, error) {
	if team1ID == team2ID {
		return nil, errors.New("команды не могут быть одинаковыми")
	}

	match := models.Match{
		Team1ID:  team1ID,
		Team2ID:  team2ID,
		Date:     date,
		Location: location,
	}

	err := db.Create(&match).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}
