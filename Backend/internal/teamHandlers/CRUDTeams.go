package teamhandlers

import (
	"basketball-league/internal/models"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"gorm.io/gorm"
)

func ListTeams(bot *tgbotapi.BotAPI, chatID int64, DB *gorm.DB) {
	var teams []models.Team

	if DB == nil || bot == nil {
		fmt.Println("База данных или бот не инициализированы.")
		os.Exit(1)
	}

	// Загружаем команды вместе с владельцами
	err := DB.Preload("Owner").Where("is_active = ?", true).Find(&teams).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Не удалось получить список команд. Попробуйте позже."))
		return
	}

	if len(teams) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Доступных команд нет."))
		return
	}

	message := "Список команд:\n"
	for _, team := range teams {
		ownerName := "Неизвестен"
		if team.Owner != nil {
			ownerName = team.Owner.Name
		}
		message += fmt.Sprintf("- %s (Владелец: %s)\n", team.Name, ownerName)
	}

	bot.Send(tgbotapi.NewMessage(chatID, message))
}

func CreateTeamName(bot *tgbotapi.BotAPI, chatID int64, msg *tgbotapi.Message, userID int, userStates map[int64]string, DB *gorm.DB) {
	teamName := strings.TrimSpace(msg.Text)

	// Проверка, существует ли уже команда с таким названием
	var existingTeam models.Team
	err := DB.Where("name = ?", teamName).First(&existingTeam).Error
	if err == nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Команда с таким названием уже существует. Попробуйте другое."))
		return
	}

	// Создание команды
	var owner models.Player
	if err := DB.Where("chat_id = ?", chatID).First(&owner).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы не зарегистрированы как игрок. Сначала используйте /register."))
		return
	}

	team := models.Team{
		Name:     teamName,
		OwnerID:  owner.ID,
		Owner:    &owner,
		IsActive: true,
	}

	if err := DB.Create(&team).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка создания команды. Попробуйте позже."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Команда '%s' успешно создана!", team.Name)))
	userStates[int64(userID)] = ""
}

func JoinTeam(bot *tgbotapi.BotAPI, chatID int64, text string, DB *gorm.DB) {
	parts := strings.Split(text, " ")
	if len(parts) < 2 {
		bot.Send(tgbotapi.NewMessage(chatID, "Используйте формат: /join_team \"Имя команды\""))
		return
	}

	teamName := strings.Join(parts[1:], " ")

	var team models.Team
	if err := DB.Where("name = ?", teamName).First(&team).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Команда с указанным именем не найдена. Проверьте имя команды."))
		return
	}

	var player models.Player
	if err := DB.Where("chat_id = ?", chatID).First(&player).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы не зарегистрированы как игрок. Сначала используйте /register."))
		return
	}

	team.Players = append(team.Players, player)
	if err := DB.Save(&team).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка вступления в команду. Попробуйте позже."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы успешно вступили в команду '%s'!", team.Name)))
}

func CreateTeam(bot *tgbotapi.BotAPI, chatID int64, userID int, userStates map[int64]string, DB *gorm.DB) {

	var owner models.Player
	if err := DB.Where("chat_id = ?", chatID).First(&owner).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы не зарегистрированы как игрок. Сначала используйте /register."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, "Введите название для команды:"))

	userStates[int64(userID)] = "create_team_name"
}

func GetTeamByID(db *gorm.DB, teamID int) *models.Team {
	var team models.Team
	err := db.Preload("Players").First(&team, teamID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// fmt.Println("\n\nпопа\n\n\n")
		return nil
	} else if err != nil {
		return nil
	}
	return &team
}

func ListPlayersWithoutTeam(bot *tgbotapi.BotAPI, chatID int64, DB *gorm.DB) {
	if DB == nil || bot == nil {
		fmt.Println("База данных или бот не инициализированы.")
		os.Exit(1)
	}

	var players []models.Player
	err := DB.Where("team_id IS NULL").Find(&players).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка при получении списка игроков. Попробуйте позже."))
		log.Printf("Ошибка при получении списка игроков без команды: %v", err)
		return
	}

	if len(players) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Нет игроков без команды."))
		return
	}

	message := "Игроки без команды:\n"
	for _, player := range players {
		message += fmt.Sprintf("- %s (Рост: %d см, Вес: %d кг, Позиция: %s)\n",
			player.Name, player.Height, player.Weight, player.Position)
	}

	bot.Send(tgbotapi.NewMessage(chatID, message))
}

func ListPlayersByTeam(bot *tgbotapi.BotAPI, chatID int64, teamName string, DB *gorm.DB) {
	if DB == nil || bot == nil {
		fmt.Println("База данных или бот не инициализированы.")
		os.Exit(1)
	}

	var team models.Team
	err := DB.Preload("Players").Where("name = ?", teamName).First(&team).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Команда '%s' не найдена.", teamName)))
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка при получении команды. Попробуйте позже."))
			log.Printf("Ошибка при получении команды: %v", err)
		}
		return
	}

	if len(team.Players) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("В команде '%s' нет зарегистрированных игроков.", teamName)))
		return
	}

	message := fmt.Sprintf("Игроки команды '%s':\n", teamName)
	for _, player := range team.Players {
		message += fmt.Sprintf("- %s (Рост: %d см, Вес: %d кг, Позиция: %s)\n", player.Name, player.Height, player.Weight, player.Position)
	}

	bot.Send(tgbotapi.NewMessage(chatID, message))
}
