package matchhandlers

import (
	"basketball-league/internal/models"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type Handler struct {
	models.Handler
}

func (h *Handler) GetUpcomingMatches() []models.Match {
    var matches []models.Match
    now := time.Now()
    h.DB.Preload("Team1").Preload("Team2").
        Where("date > ?", now).
        Order("date ASC").
        Find(&matches)
    return matches
}

func (h *Handler) GetMatchByID(matchID int) *models.Match {
	var match models.Match
	err := h.DB.Preload("Team1").Preload("Team2").First(&match, matchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	} else if err != nil {
		fmt.Println("матч не найден")
		return nil
	}
	return &match
}

func (h *Handler) GetAllMatches() ([]models.Match, error) {
	var matches []models.Match
	err := h.DB.Preload("Team1").Preload("Team2").Find(&matches).Error
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func (h *Handler) UpdateMatch(matchID int, team1ID, team2ID uint, date time.Time, location string) error {
	var match models.Match
	err := h.DB.First(&match, matchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("матч не найден")
	} else if err != nil {
		return err
	}

	match.Team1ID = team1ID
	match.Team2ID = team2ID
	match.Date = date
	match.Location = location

	return h.DB.Save(&match).Error
}

func (h *Handler) CreateMatch(team1ID, team2ID uint, date time.Time, location string) (*models.Match, error) {
	if team1ID == team2ID {
		return nil, errors.New("команды не могут быть одинаковыми")
	}

	match := models.Match{
		Team1ID:  team1ID,
		Team2ID:  team2ID,
		Date:     date,
		Location: location,
	}

	err := h.DB.Create(&match).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (h *Handler) DeleteMatch(matchID int) error {
	var match models.Match
	err := h.DB.First(&match, matchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("матч не найден")
	} else if err != nil {
		return err
	}

	return h.DB.Delete(&match).Error
}

func (h *Handler) CreateMatchStatistics(matchID uint, teamID1, teamID2 uint, team1Score, team2Score int) (*models.MatchStatistics, error) {
	var match models.Match
	if err := h.DB.First(&match, matchID).Error; err != nil {
		return nil, errors.New("матч с таким ID не найден")
	}

	stats := models.MatchStatistics{
		ID:         matchID, // ID статистики совпадает с ID матча
		MatchID:    matchID,
		TeamID1:    teamID1,
		TeamID2:    teamID2,
		Team1Score: team1Score,
		Team2Score: team2Score,
	}

	if err := h.DB.Create(&stats).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

func (h *Handler) GetStatisticsByMatchID(matchID uint) (*models.MatchStatistics, error) {
	var stats models.MatchStatistics
	err := h.DB.First(&stats, "match_id = ?", matchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("статистика для матча не найдена")
	} else if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (h *Handler) DeleteMatchStatistic(id string) string {
	statID, err := strconv.Atoi(id)
	if err != nil {
		return "Некорректный ID."
	}

	if err := h.DB.Delete(&models.MatchStatistics{}, statID).Error; err != nil {
		return "Ошибка удаления статистики матча: " + err.Error()
	}

	return fmt.Sprintf("Статистика матча с ID=%d успешно удалена.", statID)
}

func (h *Handler) CreateMatchStat() {}
