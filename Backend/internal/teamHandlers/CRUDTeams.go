package teamhandlers

import (
	"basketball-league/internal/models"
	"errors"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"gorm.io/gorm"
)

type Handler struct {
	models.Handler
}

func (h *Handler) ListTeams(bot *tgbotapi.BotAPI, chatID int64) {
	var teams []models.Team

	// Загружаем команды вместе с владельцами
	err := h.DB.Preload("Owner").Where("is_active = ?", true).Find(&teams).Error
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

func (h *Handler) CreateTeamName(bot *tgbotapi.BotAPI, chatID int64, msg *tgbotapi.Message, userID int, userStates map[int64]string) {
	teamName := strings.TrimSpace(msg.Text)

	// Проверка, существует ли уже команда с таким названием
	var existingTeam models.Team
	err := h.DB.Where("name = ?", teamName).First(&existingTeam).Error
	if err == nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Команда с таким названием уже существует. Попробуйте другое."))
		return
	}

	// Создание команды
	var owner models.Player
	if err := h.DB.Where("chat_id = ?", chatID).First(&owner).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы не зарегистрированы как игрок. Сначала используйте /register."))
		return
	}

	team := models.Team{
		Name:     teamName,
		OwnerID:  owner.ID,
		Owner:    &owner,
		IsActive: true,
	}

	if err := h.DB.Create(&team).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка создания команды. Попробуйте позже."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Команда '%s' успешно создана!", team.Name)))
	userStates[int64(userID)] = ""
}

func (h *Handler) JoinTeam(bot *tgbotapi.BotAPI, chatID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) < 2 {
		bot.Send(tgbotapi.NewMessage(chatID, "Используйте формат: /join_team \"Имя команды\""))
		return
	}

	teamName := strings.Join(parts[1:], " ")

	var team models.Team
	if err := h.DB.Where("name = ?", teamName).First(&team).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Команда с указанным именем не найдена. Проверьте имя команды."))
		return
	}

	var player models.Player
	if err := h.DB.Where("chat_id = ?", chatID).First(&player).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы не зарегистрированы как игрок. Сначала используйте /register."))
		return
	}

	team.Players = append(team.Players, player)
	if err := h.DB.Save(&team).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка вступления в команду. Попробуйте позже."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы успешно вступили в команду '%s'!", team.Name)))
}

func (h *Handler) CreateTeam(bot *tgbotapi.BotAPI, chatID int64, userID int, userStates map[int64]string) {

	var owner models.Player
	if err := h.DB.Where("chat_id = ?", chatID).First(&owner).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы не зарегистрированы как игрок. Сначала используйте /register."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, "Введите название для команды:"))

	userStates[int64(userID)] = "create_team_name"
}

func (h *Handler) GetTeamByID(teamID int) *models.Team {
	var team models.Team
	err := h.DB.Preload("Players").First(&team, teamID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// fmt.Println("\n\nпопа\n\n\n")
		return nil
	} else if err != nil {
		return nil
	}
	return &team
}

// RenameTeam позволяет владельцу команды переименовать команду
func (h *Handler) RenameTeam(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, userStates map[int64]string) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	// Проверка, находится ли пользователь в состоянии переименования
	if userStates[userID] != "rename_team" {
		userStates[userID] = "rename_team"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите новое название для вашей команды:"))
		return
	}

	// Проверка существования команды, где пользователь является владельцем
	var team models.Team
	err := h.DB.Where("owner_id = ?", userID).First(&team).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			bot.Send(tgbotapi.NewMessage(chatID, "Вы не являетесь владельцем команды или команда не найдена."))
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при поиске вашей команды. Попробуйте позже."))
			log.Printf("Ошибка при поиске команды: %v", err)
		}
		delete(userStates, userID)
		return
	}

	// Проверка длины нового названия
	newTeamName := msg.Text
	if len(newTeamName) < 3 {
		bot.Send(tgbotapi.NewMessage(chatID, "Название команды должно содержать хотя бы 3 символа. Попробуйте снова."))
		return
	}

	// Обновление имени команды в базе данных
	team.Name = newTeamName
	err = h.DB.Save(&team).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при обновлении названия команды. Попробуйте снова позже."))
		log.Printf("Ошибка при обновлении названия команды: %v", err)
		return
	}

	delete(userStates, userID)

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Название вашей команды успешно обновлено на: %s", newTeamName)))
}

func (h *Handler) ListPlayersWithoutTeam(bot *tgbotapi.BotAPI, chatID int64) {
	var players []models.Player

	err := h.DB.Where("team_id IS NULL").Find(&players).Error
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

func (h *Handler) ListPlayersByTeam(bot *tgbotapi.BotAPI, chatID int64, teamName string) {
	var team models.Team

	err := h.DB.Preload("Players").Where("name = ?", teamName).First(&team).Error
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

func (h *Handler) DeleteTeamByName(bot *tgbotapi.BotAPI, chatID int64, teamName string) {
	var team models.Team

	err := h.DB.Where("name = ?", teamName).First(&team).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Команда с именем '%s' не найдена.", teamName)))
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка при поиске команды. Попробуйте снова позже."))
			log.Printf("Ошибка при поиске команды: %v", err)
		}
		return
	}

	err = h.DB.Delete(&team).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка при удалении команды. Попробуйте снова позже."))
		log.Printf("Ошибка при удалении команды: %v", err)
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Команда '%s' успешно удалена.", teamName)))
}

func (h *Handler) RemovePlayerFromTeam(bot *tgbotapi.BotAPI, chatID int64, playerName string) {
	var player models.Player

	err := h.DB.Where("name = ?", playerName).First(&player).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Игрок с именем '%s' не найден.", playerName)))
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка при поиске игрока. Попробуйте снова позже."))
			log.Printf("Ошибка при поиске игрока: %v", err)
		}
		return
	}

	if player.TeamID == nil {
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Игрок '%s' уже не состоит ни в одной команде.", playerName)))
		return
	}

	player.TeamID = nil
	err = h.DB.Save(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка при удалении игрока из команды. Попробуйте снова позже."))
		log.Printf("Ошибка при обновлении данных игрока: %v", err)
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Игрок '%s' успешно удалён из команды.", playerName)))
}
