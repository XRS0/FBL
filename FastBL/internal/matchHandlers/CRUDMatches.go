package matchhandlers

import (
	"basketball-league/internal/models"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

// Просмотр списка матчей
func ListMatches(bot *tgbotapi.BotAPI, chatID int64, DB *gorm.DB) {
	var matches []models.Match

	err := DB.Find(&matches).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Не удалось получить список матчей. Попробуйте позже."))
		return
	}

	if len(matches) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Предстоящих матчей нет."))
		return
	}

	message := "Список предстоящих матчей:\n"
	for _, match := range matches {
		message += fmt.Sprintf("- Матч #%d: %v vs %v (%s)\n", match.ID, match.Team1ID, match.Team2ID, match.Location)
	}
	bot.Send(tgbotapi.NewMessage(chatID, message))
}

// Запись на матч
func JoinMatch(bot *tgbotapi.BotAPI, chatID int64, text string, DB *gorm.DB) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		bot.Send(tgbotapi.NewMessage(chatID, "Используйте формат: /join_match ID_матча"))
		return
	}

	matchID, err := strconv.Atoi(parts[1])
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "ID матча должно быть числом."))
		return
	}

	var match models.Match
	if err := DB.First(&match, matchID).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Матч не найден. Проверьте ID матча."))
		return
	}

	var player models.Player
	if err := DB.Where("chat_id = ?", chatID).First(&player).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы не зарегистрированы как игрок. Сначала используйте /register."))
		return
	}

	// Привязка игрока к матчу
	// match. = append(match.Players, player)
	// if err := DB.Save(&match).Error; err != nil {
	// 	bot.Send(tgbotapi.NewMessage(chatID, "Ошибка записи на матч. Попробуйте позже."))
	// 	return
	// }

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы успешно записались на матч #%d!", match.ID)))
}

// Добавление новой статистики матча
func CreateMatchStatistic(db *gorm.DB, stat models.MatchStatistics) error {
	if err := db.Create(&stat).Error; err != nil {
		return fmt.Errorf("ошибка при создании записи: %w", err)
	}
	return nil
}

// Получение статистики матча по ID
func GetMatchStatisticByID(db *gorm.DB, id uint) (models.MatchStatistics, error) {
	var stat models.MatchStatistics
	if err := db.First(&stat, id).Error; err != nil {
		return stat, fmt.Errorf("запись с ID %d не найдена: %w", id, err)
	}
	return stat, nil
}

// Обновление статистики матча
func UpdateMatchStatistic(db *gorm.DB, id uint, updatedStat models.MatchStatistics) error {
	var stat models.MatchStatistics
	if err := db.First(&stat, id).Error; err != nil {
		return fmt.Errorf("запись с ID %d не найдена: %w", id, err)
	}

	stat.MatchID = updatedStat.MatchID
	stat.TeamID1 = updatedStat.TeamID1
	stat.TeamID2 = updatedStat.TeamID2
	stat.Team1Score = updatedStat.Team1Score
	stat.Team2Score = updatedStat.Team2Score

	if err := db.Save(&stat).Error; err != nil {
		return fmt.Errorf("ошибка при обновлении записи: %w", err)
	}
	return nil
}

// Удаление статистики матча по ID
func DeleteMatchStatistic(db *gorm.DB, id uint) error {
	if err := db.Delete(&models.MatchStatistics{}, id).Error; err != nil {
		return fmt.Errorf("ошибка при удалении записи с ID %d: %w", id, err)
	}
	return nil
}
