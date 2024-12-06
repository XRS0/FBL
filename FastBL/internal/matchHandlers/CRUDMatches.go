package matchhandlers

import (
	"basketball-league/internal/models"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// GetMatchByID получает информацию о матче по ID
func GetMatchByID(db *gorm.DB, matchID int) *models.Match {
	var match models.Match
	err := db.Preload("Team1").Preload("Team2").First(&match, matchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	} else if err != nil {
		fmt.Println("матч не найден")
		return nil
	}
	return &match
}

// GetAllMatches возвращает список всех матчей
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

	match.Team1ID = team1ID
	match.Team2ID = team2ID
	match.Date = date
	match.Location = location

	return db.Save(&match).Error
}

// CreateMatch создает новый матч
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

// DeleteMatch удаляет матч по ID
func DeleteMatch(db *gorm.DB, matchID int) error {
	var match models.Match
	err := db.First(&match, matchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("матч не найден")
	} else if err != nil {
		return err
	}

	return db.Delete(&match).Error
}

// Добавление новой статистики матча
func CreateMatchStatistics(db *gorm.DB, matchID uint, teamID1, teamID2 uint, team1Score, team2Score int) (*models.MatchStatistics, error) {
	// Проверяем, существует ли матч с таким ID
	var match models.Match
	if err := db.First(&match, matchID).Error; err != nil {
		return nil, errors.New("матч с таким ID не найден")
	}

	// Создаем статистику матча с таким же ID
	stats := models.MatchStatistics{
		ID:         matchID, // ID статистики совпадает с ID матча
		MatchID:    matchID,
		TeamID1:    teamID1,
		TeamID2:    teamID2,
		Team1Score: team1Score,
		Team2Score: team2Score,
	}

	if err := db.Create(&stats).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// Получение записи статистики по ID
func GetStatisticsByMatchID(db *gorm.DB, matchID uint) (*models.MatchStatistics, error) {
	var stats models.MatchStatistics
	err := db.First(&stats, "match_id = ?", matchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("статистика для матча не найдена")
	} else if err != nil {
		return nil, err
	}
	return &stats, nil
}

// Удаление записи статистики по ID
func DeleteMatchStatistic(db *gorm.DB, id string) string {
	statID, err := strconv.Atoi(id)
	if err != nil {
		return "Некорректный ID."
	}

	if err := db.Delete(&models.MatchStatistics{}, statID).Error; err != nil {
		return "Ошибка удаления статистики матча: " + err.Error()
	}

	return fmt.Sprintf("Статистика матча с ID=%d успешно удалена.", statID)
}
