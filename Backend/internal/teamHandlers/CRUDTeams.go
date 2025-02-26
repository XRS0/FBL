package teamhandlers

import (
	"basketball-league/internal/models"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type Handler struct {
	models.Handler
}

func (h *Handler) GetAllTeams() []models.Team {
    var teams []models.Team
    h.DB.Find(&teams)
    return teams
}

// ----------------------------------------------------------------------------
// Вспомогательная функция для отправки текста
// ----------------------------------------------------------------------------
func (h *Handler) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	h.Bot.Send(msg)
}

// ----------------------------------------------------------------------------
// Пример методов
// ----------------------------------------------------------------------------

func (h *Handler) GetTeamByName(name string) (*models.Team, error) {
	var team models.Team
	if err := h.DB.Preload("Players").Where("name = ?", name).First(&team).Error; err != nil {
		return nil, errors.New("команда не найдена")
	}
	return &team, nil
}

func (h *Handler) ListTeams(chatID int64) {
	var teams []models.Team

	err := h.DB.Preload("Owner").Where("is_active = ?", true).Find(&teams).Error
	if err != nil {
		h.sendText(chatID, "Не удалось получить список команд. Попробуйте позже.")
		return
	}

	if len(teams) == 0 {
		h.sendText(chatID, "Доступных команд нет.")
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

	h.sendText(chatID, message)
}

func (h *Handler) CreateTeamName(chatID int64, msg *tgbotapi.Message, userID int, userStates map[int64]string) {
	teamName := strings.TrimSpace(msg.Text)

	// Проверка, существует ли уже команда с таким названием
	var existingTeam models.Team
	err := h.DB.Where("name = ?", teamName).First(&existingTeam).Error
	if err == nil {
		h.sendText(chatID, "Команда с таким названием уже существует. Попробуйте другое.")
		return
	}

	// Создание команды
	var owner models.Player
	if err := h.DB.Where("chat_id = ?", chatID).First(&owner).Error; err != nil {
		h.sendText(chatID, "Вы не зарегистрированы как игрок. Сначала используйте /register.")
		return
	}

	team := models.Team{
		Name:     teamName,
		OwnerID:  owner.ID,
		Owner:    &owner,
		IsActive: true,
	}

	if err := h.DB.Create(&team).Error; err != nil {
		h.sendText(chatID, "Ошибка создания команды. Попробуйте позже.")
		return
	}

	h.sendText(chatID, fmt.Sprintf("Команда '%s' успешно создана!", team.Name))
	userStates[int64(userID)] = ""
}

func (h *Handler) JoinTeam(chatID int64, text string, userStates map[int64]string) {
	// Если уже в процессе выбора номера
	if userStates[chatID] == "join_team" {
		if strings.TrimSpace(text) == "unk" {
			h.sendText(chatID, "Вы успешно вступили в команду! Номер игрока будет выбран позже.")
			delete(userStates, chatID)
			return
		}

		userNumber, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil || userNumber < 1 || userNumber > 100 {
			h.sendText(chatID, "Укажите корректный номер (от 1 до 100), или напишите 'unk' для того чтобы выбрать номер позже.")
			return
		}

		var player models.Player
		if err := h.DB.Where("chat_id = ?", chatID).First(&player).Error; err != nil {
			h.sendText(chatID, "Вы не зарегистрированы как игрок. Сначала используйте /register.")
			return
		}

		var existingPlayer models.Player
		err = h.DB.Where("team_id = ? AND number = ?", player.TeamID, userNumber).First(&existingPlayer).Error
		if err != nil {
			// Если игрок с таким номером не найден
			if errors.Is(err, gorm.ErrRecordNotFound) {
				player.Number = uint8(userNumber) // Устанавливаем номер игрока
				if err := h.DB.Save(&player).Error; err != nil {
					h.sendText(chatID, "Ошибка вступления в команду. Попробуйте позже.")
					return
				}

				h.sendText(chatID, fmt.Sprintf("Вы успешно вступили в команду с номером %d!", userNumber))
				delete(userStates, chatID)
				return
			}
			// Иная ошибка
			h.sendText(chatID, "Ошибка вступления в команду. Попробуйте позже.")
			return
		}

		// Если игрок с таким номером уже есть
		h.sendText(chatID, "Игрок с таким номером уже существует, попробуйте другой.")
		return
	}

	// Иначе мы только что получили команду /join_team "Имя команды"
	parts := strings.Split(text, " ")
	if len(parts) < 2 {
		h.sendText(chatID, "Используйте формат: /join_team \"Имя команды\"")
		return
	}

	teamName := strings.Join(parts[1:], " ")

	var team models.Team
	if err := h.DB.Where("name = ?", teamName).First(&team).Error; err != nil {
		h.sendText(chatID, "Команда с указанным именем не найдена. Проверьте имя команды.")
		return
	}

	var player models.Player
	if err := h.DB.Where("chat_id = ?", chatID).First(&player).Error; err != nil {
		h.sendText(chatID, "Вы не зарегистрированы как игрок. Сначала используйте /register.")
		return
	}

	team.Players = append(team.Players, player)
	if err := h.DB.Save(&team).Error; err != nil {
		h.sendText(chatID, "Ошибка вступления в команду. Попробуйте позже.")
		return
	}

	h.sendText(chatID, "Напишите желаемый номер игрока в команде (от 1 до 100), или 'unk' для выбора позже.")
	userStates[chatID] = "join_team"
}

func (h *Handler) CreateTeam(chatID int64, userID int, userStates map[int64]string) {
	var owner models.Player
	if err := h.DB.Where("chat_id = ?", chatID).First(&owner).Error; err != nil {
		h.sendText(chatID, "Вы не зарегистрированы как игрок. Сначала используйте /register.")
		return
	}

	h.sendText(chatID, "Введите название для команды:")
	userStates[int64(userID)] = "create_team_name"
}

func (h *Handler) GetTeamByID(teamID int) *models.Team {
	var team models.Team
	err := h.DB.Preload("Players").First(&team, teamID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	} else if err != nil {
		return nil
	}
	return &team
}

// RenameTeam позволяет владельцу команды переименовать команду
func (h *Handler) RenameTeam(msg *tgbotapi.Message, userStates map[int64]string) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	// Проверяем состояние
	if userStates[userID] != "rename_team" {
		userStates[userID] = "rename_team"
		h.sendText(chatID, "Введите новое название для вашей команды:")
		return
	}

	// Ищем команду, где пользователь владелец
	var team models.Team
	err := h.DB.Where("owner_id = ?", userID).First(&team).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.sendText(chatID, "Вы не являетесь владельцем команды или команда не найдена.")
		} else {
			h.sendText(chatID, "Ошибка при поиске вашей команды. Попробуйте позже.")
			log.Printf("Ошибка при поиске команды: %v", err)
		}
		delete(userStates, userID)
		return
	}

	// Проверка длины названия
	newTeamName := msg.Text
	if len(newTeamName) < 3 {
		h.sendText(chatID, "Название команды должно содержать хотя бы 3 символа. Попробуйте снова.")
		return
	}

	// Сохраняем в БД
	team.Name = newTeamName
	err = h.DB.Save(&team).Error
	if err != nil {
		h.sendText(chatID, "Ошибка при обновлении названия команды. Попробуйте снова позже.")
		log.Printf("Ошибка при обновлении названия команды: %v", err)
		return
	}

	delete(userStates, userID)
	h.sendText(chatID, fmt.Sprintf("Название вашей команды успешно обновлено на: %s", newTeamName))
}

func (h *Handler) ListPlayersWithoutTeam(chatID int64, isAdmin bool) {
	var players []models.Player

	err := h.DB.Where("team_id IS NULL").Find(&players).Error
	if err != nil {
		h.sendText(chatID, "Произошла ошибка при получении списка игроков. Попробуйте позже.")
		log.Printf("Ошибка при получении списка игроков без команды: %v", err)
		return
	}

	if len(players) == 0 {
		h.sendText(chatID, "Нет игроков без команды.")
		return
	}

	message := "Игроки без команды:\n"
	if isAdmin {
		for _, player := range players {
			message += fmt.Sprintf("- %s (Рост: %d см, Вес: %d кг, Позиция: %s, Контакт игрока: %s)\n",
				player.Name, player.Height, player.Weight, player.Position, player.Contact)
		}
	} else {
		for _, player := range players {
			message += fmt.Sprintf("- %s (Рост: %d см, Вес: %d кг, Позиция: %s)\n",
				player.Name, player.Height, player.Weight, player.Position)
		}
	}

	h.sendText(chatID, message)
}

func (h *Handler) ListPlayersByTeam(chatID int64, teamName string, isAdmin bool) {
	var team models.Team

	err := h.DB.Preload("Players").Where("name = ?", teamName).First(&team).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.sendText(chatID, fmt.Sprintf("Команда '%s' не найдена.", teamName))
		} else {
			h.sendText(chatID, "Произошла ошибка при получении команды. Попробуйте позже.")
			log.Printf("Ошибка при получении команды: %v", err)
		}
		return
	}

	if len(team.Players) == 0 {
		h.sendText(chatID, fmt.Sprintf("В команде '%s' нет зарегистрированных игроков.", teamName))
		return
	}

	message := fmt.Sprintf("Игроки команды '%s':\n\n", teamName)
	if isAdmin {
		for _, player := range team.Players {
			message += fmt.Sprintf(
				"- %s (Рост: %d см, Вес: %d кг, Позиция: %s, Номер игрока: %d, Контактные данные: %s)\n\n",
				player.Name, player.Height, player.Weight, player.Position, player.Number, player.Contact,
			)
		}
	} else {
		for _, player := range team.Players {
			message += fmt.Sprintf(
				"- %s (Рост: %d см, Вес: %d кг, Позиция: %s, Номер игрока: %d)\n\n",
				player.Name, player.Height, player.Weight, player.Position, player.Number,
			)
		}
	}

	h.sendText(chatID, message)
}

func (h *Handler) DeleteTeamByName(chatID int64, teamName string) {
	var team models.Team

	err := h.DB.Where("name = ?", teamName).First(&team).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.sendText(chatID, fmt.Sprintf("Команда с именем '%s' не найдена.", teamName))
		} else {
			h.sendText(chatID, "Произошла ошибка при поиске команды. Попробуйте снова позже.")
			log.Printf("Ошибка при поиске команды: %v", err)
		}
		return
	}

	err = h.DB.Delete(&team).Error
	if err != nil {
		h.sendText(chatID, "Произошла ошибка при удалении команды. Попробуйте снова позже.")
		log.Printf("Ошибка при удалении команды: %v", err)
		return
	}

	h.sendText(chatID, fmt.Sprintf("Команда '%s' успешно удалена.", teamName))
}

func (h *Handler) RemovePlayerFromTeam(chatID int64, playerName string) {
	var player models.Player

	err := h.DB.Where("name = ?", playerName).First(&player).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.sendText(chatID, fmt.Sprintf("Игрок с именем '%s' не найден.", playerName))
		} else {
			h.sendText(chatID, "Произошла ошибка при поиске игрока. Попробуйте снова позже.")
			log.Printf("Ошибка при поиске игрока: %v", err)
		}
		return
	}

	if player.TeamID == nil {
		h.sendText(chatID, fmt.Sprintf("Игрок '%s' уже не состоит ни в одной команде.", playerName))
		return
	}

	player.TeamID = nil
	if err := h.DB.Save(&player).Error; err != nil {
		h.sendText(chatID, "Произошла ошибка при удалении игрока из команды. Попробуйте снова позже.")
		log.Printf("Ошибка при обновлении данных игрока: %v", err)
		return
	}

	h.sendText(chatID, fmt.Sprintf("Игрок '%s' успешно удалён из команды.", playerName))
}

func (h *Handler) IsOwner(chatId int64, teamName string) bool {
	var player models.Player
	if err := h.DB.Where("chat_id = ?", chatId).First(&player).Error; err != nil {
		h.sendText(chatId, "Произошла ошибка, попробуйте позже")
		return false
	}

	var team models.Team
	if err := h.DB.Where("name = ?", teamName).First(&team).Error; err != nil {
		h.sendText(chatId, "Произошла ошибка, попробуйте позже")
		return false
	}

	return team.OwnerID == player.ID
}

func (h *Handler) RemovePlayerByNumber(ownerChatID int64, teamID int, number uint8) error {
	var owner models.Player
	if err := h.DB.Where("chat_id = ?", ownerChatID).First(&owner).Error; err != nil {
		return fmt.Errorf("игрок не найден")
	}

	// Проверяем, что пользователь владеет выбранной командой
	var team models.Team
	if err := h.DB.Where("id = ? AND owner_id = ?", teamID, owner.ID).First(&team).Error; err != nil {
		return fmt.Errorf("команда не найдена или вы не владелец")
	}

	var player models.Player
	if err := h.DB.Where("team_id = ? AND number = ?", teamID, number).First(&player).Error; err != nil {
		return fmt.Errorf("игрок с номером %d не найден", number)
	}

	player.TeamID = nil
	if err := h.DB.Save(&player).Error; err != nil {
		return fmt.Errorf("ошибка сохранения изменений")
	}

	return nil
}

// HasLogo возвращает true/false в зависимости от того, установлен ли PathToLogo,
// и если установлен, возвращает сам путь.
func (h *Handler) HasLogo(teamID int) (bool, string, error) {
	var team models.Team
	if err := h.DB.Select("id", "path_to_logo").
		Where("id = ?", teamID).
		First(&team).Error; err != nil {
		// Если записи нет или произошла другая ошибка, возвращаем ошибку
		return false, "", err
	}

	// Если поле PathToLogo пустое, значит лого нет
	if team.PathToLogo == "" {
		return false, "", nil
	}

	// Иначе лого есть, возвращаем путь
	return true, team.PathToLogo, nil
}

// UpdateLogoPath записывает новый путь к логотипу в поле PathToLogo
func (h *Handler) UpdateLogoPath(teamID int, newPath string) error {
	return h.DB.Model(&models.Team{}).
		Where("id = ?", teamID).
		Update("path_to_logo", newPath).
		Error
}

