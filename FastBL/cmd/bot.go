package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	. "basketball-league/internal/WSH"
	. "basketball-league/internal/db"
	. "basketball-league/internal/matchHandlers"
	"basketball-league/internal/models"
	. "basketball-league/internal/teamHandlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

var DB *gorm.DB
var userStates = make(map[int64]string)

var admins = map[int64]bool{
	1324977667: true,
	984866387:  true,
}

// Проверка, является ли пользователь администратором
func isAdmin(chatID int64) bool {
	return admins[chatID]
}

func main() {
	DB = InitDatabase()
	go StartWS(DB)

	bot, err := tgbotapi.NewBotAPI("7945815181:AAHAzN3QI5dUtq7iSmw9if2rrA5Rzi2j3bY")
	if err != nil {
		log.Fatalf("Не удалось инициализировать бота: %v", err)
	}

	bot.Debug = true
	log.Printf("Авторизация выполнена на аккаунте %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update.Message)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	state := userStates[userID]

	switch state {
	case "register":
		registerPlayer(bot, msg, DB)
	case "create_team_name":
		CreateTeamName(bot, chatID, msg, int(userID), userStates, DB)
	default:
		processCommand(bot, msg, chatID, userID)
	}
}

// Обработка команд
func processCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, chatID int64, userID int64) {
	commandParts := strings.SplitN(msg.Text, " ", 2)
	command := commandParts[0]

	switch command {
	case "/start":
		sendStartMessage(bot, chatID)
	case "/register":
		userStates[userID] = "register"
		registerPlayer(bot, msg, DB)
	case "/profile":
		listProfile(bot, chatID, DB)
	case "/teams":
		ListTeams(bot, chatID, DB)
	case "/create_team":
		CreateTeam(bot, chatID, int(userID), userStates, DB)
	case "/logout":
		logout(bot, chatID, userID)
	case "/players":
		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте формат: /players ИМЯКОМАНДЫ"))
		} else {
			teamName := strings.TrimSpace(commandParts[1])
			ListPlayersByTeam(bot, chatID, teamName, DB)
		}
	case "/create_match":
		if !isAdmin(chatID) {
			bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для выполнения этой команды."))
			return
		}
		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте: /create_match <Team1ID> <Team2ID> <Date> <Location>"))
			return
		}
		args := strings.Fields(commandParts[1])
		if len(args) < 4 {
			bot.Send(tgbotapi.NewMessage(chatID, "Недостаточно данных. Используйте: /create_match <Team1ID> <Team2ID> <Date> <Location>"))
			return
		}
		team1ID, _ := strconv.Atoi(args[0])
		team2ID, _ := strconv.Atoi(args[1])
		date, _ := time.Parse("2006-01-02 15:04:05", args[2])
		location := args[3]
		match, err := CreateMatch(DB, uint(team1ID), uint(team2ID), date, location)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка создания матча: "+err.Error()))
			return
		}
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Матч создан: #%d", match.ID)))

	case "/get_match":
		if !isAdmin(chatID) {
			bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для выполнения этой команды."))
			return
		}
		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте: /get_match <ID>"))
			return
		}
		matchID, _ := strconv.Atoi(commandParts[1])
		match, err := GetMatchByID(DB, matchID)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка: "+err.Error()))
			return
		}
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Матч #%d: %s vs %s в %s", match.ID, match.Team1.Name, match.Team2.Name, match.Location)))

	case "/delete_match":
		if !isAdmin(chatID) {
			bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для выполнения этой команды."))
			return
		}
		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте: /delete_match <ID>"))
			return
		}
		matchID, _ := strconv.Atoi(commandParts[1])
		err := DeleteMatch(DB, matchID)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка удаления матча: "+err.Error()))
			return
		}
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Матч #%d успешно удален.", matchID)))

	case "/create_stat":
		if !isAdmin(chatID) {
			bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для выполнения этой команды."))
			return
		}

		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте: /create_stat <MatchID> <TeamID1> <TeamID2> <Team1Score> <Team2Score>"))
			return
		}

		data := strings.Fields(commandParts[1])
		if len(data) != 5 {
			bot.Send(tgbotapi.NewMessage(chatID, "Неверное количество параметров. Используйте: /create_stat <MatchID> <TeamID1> <TeamID2> <Team1Score> <Team2Score>"))
			return
		}

		matchID, err := strconv.ParseUint(data[0], 10, 32)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "MatchID должен быть числом."))
			return
		}

		teamID1, err := strconv.ParseUint(data[1], 10, 32)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "TeamID1 должен быть числом."))
			return
		}

		teamID2, err := strconv.ParseUint(data[2], 10, 32)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "TeamID2 должен быть числом."))
			return
		}

		team1Score, err := strconv.Atoi(data[3])
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Team1Score должен быть числом."))
			return
		}

		team2Score, err := strconv.Atoi(data[4])
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Team2Score должен быть числом."))
			return
		}

		stat, err := CreateMatchStatistics(DB, uint(matchID), uint(teamID1), uint(teamID2), team1Score, team2Score)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка создания статистики: %v", err)))
			return
		}

		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Статистика успешно создана для матча #%d", stat.ID)))

	case "/get_stat":
		if !isAdmin(chatID) {
			bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для выполнения этой команды."))
			return
		}

		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте: /get_stat <MatchID>"))
			return
		}

		matchID, err := strconv.ParseUint(commandParts[1], 10, 32)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "MatchID должен быть числом."))
			return
		}

		stat, err := GetStatisticsByMatchID(DB, uint(matchID))
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка: %v", err)))
			return
		}

		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Статистика матча #%d: %v", matchID, stat)))

	case "/delete_stat":
		if !isAdmin(chatID) {
			bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для выполнения этой команды."))
			return
		}
		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте: /delete_stat <ID>"))
			return
		}
		response := DeleteMatchStatistic(DB, commandParts[1])
		bot.Send(tgbotapi.NewMessage(chatID, response))

	default:
		if strings.HasPrefix(msg.Text, "/join_team") {
			JoinTeam(bot, chatID, msg.Text, DB)
		} else if strings.HasPrefix(userStates[userID], "register") {
			registerPlayer(bot, msg, DB)
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда. Попробуйте /start."))
		}
	}
}

// Отправка приветственного сообщения
func sendStartMessage(bot *tgbotapi.BotAPI, chatID int64) {
	message := "Добро пожаловать!\n\n" +
		"Вот список доступных команд:\n" +
		"- /profile - Просмотреть свой профиль\n" +
		"- /register - Зарегистрироваться как игрок\n" +
		"- /teams - Просмотреть команды и вступить\n" +
		"- /players - Просмотреть игроков определенной команды\n" +
		"- /matches - Просмотреть матчи\n" +
		"- /create_team - Создать свою команду\n" +
		"- /join_team - Вступить в команду\n" +
		"- /join_match - Записаться на матч\n" +
		"- /statistics - Просмотреть статистику матчей\n" +
		"- /start - Получить справку по командам\n" +
		"- /logout - Выйти из аккаунта\n\n" +
		"Для администраторов:\n" +
		"- /create_stat - Создать статистику матча\n" +
		"- /get_stat <ID> - Получить статистику матча по ID\n" +
		"- /delete_stat <ID> - Удалить статистику матча\n" +
		"- /create_match <Team1ID> <Team2ID> <Date> <Location> - Создать новый матч\n" +
		"- /get_match <ID> - Получить информацию о матче по ID\n" +
		"- /delete_match <ID> - Удалить матч\n" +
		"- /update_match <ID> <Team1ID> <Team2ID> <Date> <Location> - Обновить матч\n\n" +
		"Выберите команду и начните взаимодействовать с ботом."

	msg := tgbotapi.NewMessage(chatID, message)
	bot.Send(msg)
}

func registerPlayer(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, DB *gorm.DB) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	// Инициализация данных пользователя, если ранее не была начата регистрация
	if _, exists := userStates[userID]; !exists {
		userStates[userID] = "register_name"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваше имя:"))
		return
	}

	var existingPlayer models.Player
	err := DB.Where("chat_id = ?", chatID).First(&existingPlayer).Error
	if err == nil {
		// Если игрок найден в базе данных, сообщаем о том, что он уже зарегистрирован
		bot.Send(tgbotapi.NewMessage(chatID, "Вы уже зарегистрированы в системе!"))
		userStates[userID] = ""
		return
	}

	// Состояние регистрации
	state := userStates[userID]
	switch state {
	case "register_name":
		if len(msg.Text) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Имя должно быть длиннее 1 символа. Попробуйте снова."))
			return
		}
		setTemporaryData(userID, "name", msg.Text)
		userStates[userID] = "register_height"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваш рост (см):"))

	case "register_height":
		height, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || height < 100 || height > 250 {
			bot.Send(tgbotapi.NewMessage(chatID, "Укажите корректный рост в сантиметрах (от 100 до 250)."))
			return
		}
		setTemporaryData(userID, "height", strconv.Itoa(height))
		userStates[userID] = "register_weight"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваш вес (кг):"))

	case "register_weight":
		weight, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || weight < 30 || weight > 200 {
			bot.Send(tgbotapi.NewMessage(chatID, "Укажите корректный вес в килограммах (от 30 до 200)."))
			return
		}
		setTemporaryData(userID, "weight", strconv.Itoa(weight))
		userStates[userID] = "register_position"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите вашу игровую позицию (например, Центровой, Разыгрывающий):"))

	case "register_position":
		if len(msg.Text) < 3 {
			bot.Send(tgbotapi.NewMessage(chatID, "Позиция должна содержать хотя бы 3 символа. Попробуйте снова."))
			return
		}
		setTemporaryData(userID, "position", msg.Text)
		userStates[userID] = "register_contact"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваш контакт (например, номер телефона или email):"))

	case "register_contact":
		if len(msg.Text) < 5 {
			bot.Send(tgbotapi.NewMessage(chatID, "Контактная информация должна содержать хотя бы 5 символов. Попробуйте снова."))
			return
		}
		setTemporaryData(userID, "contact", msg.Text)

		// Создание игрока в базе данных
		tempData := getTemporaryData(userID)
		height, _ := strconv.Atoi(tempData["height"])
		weight, _ := strconv.Atoi(tempData["weight"])

		player := models.Player{
			Name:     tempData["name"],
			Height:   height,
			Weight:   weight,
			Position: tempData["position"],
			Contact:  tempData["contact"],
			ChatID:   chatID,
		}

		err := DB.Create(&player).Error
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при регистрации. Попробуйте снова."))
			log.Printf("Ошибка при регистрации игрока: %v", err)
			return
		}

		// Сброс состояния пользователя
		deleteTemporaryData(userID)
		delete(userStates, userID)

		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Игрок %s успешно зарегистрирован!", player.Name)))
	default:
		userStates[userID] = "register_name"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваше имя:"))
	}
}

// Хранилище временных данных для пользователя
var temporaryData = make(map[int64]map[string]string)

// Установка временных данных
func setTemporaryData(userID int64, key, value string) {
	if _, exists := temporaryData[userID]; !exists {
		temporaryData[userID] = make(map[string]string)
	}
	temporaryData[userID][key] = value
}

// Получение временных данных
func getTemporaryData(userID int64) map[string]string {
	if data, exists := temporaryData[userID]; exists {
		return data
	}
	return make(map[string]string)
}

// Удаление временных данных
func deleteTemporaryData(userID int64) {
	delete(temporaryData, userID)
}

// Выход из аккаунта
func logout(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	// Удаление игрока из базы данных
	err := DB.Where("chat_id = ?", chatID).Delete(&models.Player{}).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при выходе из аккаунта. Попробуйте снова."))
		log.Printf("Ошибка при удалении аккаунта: %v", err)
		return
	}

	// Очистка временных данных и состояния пользователя
	delete(userStates, userID)
	deleteTemporaryData(userID)

	bot.Send(tgbotapi.NewMessage(chatID, "Вы успешно вышли из аккаунта. Для повторной регистрации используйте команду /register."))
}

// Просмотр профиля
func listProfile(bot *tgbotapi.BotAPI, chatID int64, DB *gorm.DB) {
	var player models.Player

	err := DB.Where("chat_id = ?", chatID).First(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы не зарегистрированы. Используйте команду /register."))
		return
	}

	message := fmt.Sprintf("Имя: %s\nРост: %d см\nВес: %d кг\nПозиция: %s\nКонтакты: %s",
		player.Name, player.Height, player.Weight, player.Position, player.Contact)
	bot.Send(tgbotapi.NewMessage(chatID, message))
}

// func viewStatistics(bot *tgbotapi.BotAPI, chatID int64) {
// 	var player models.Player
// 	err := DB.Where("chat_id = ?", chatID).Preload("Team").First(&player).Error
// 	if err != nil {
// 		bot.Send(tgbotapi.NewMessage(chatID, "Вы еще не зарегистрированы. Используйте /register."))
// 		return
// 	}

// 	var stats []models.MatchStatistics
// 	err = DB.Where("player_id = ?", player.ID).Find(&stats).Error
// 	if err != nil {
// 		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при получении статистики. Попробуйте позже."))
// 		log.Printf("Ошибка получения статистики: %v", err)
// 		return
// 	}

// 	totalPoints, totalAssists, totalRebounds := 0, 0, 0
// 	for _, stat := range stats {
// 		totalPoints += stat.Points
// 		totalAssists += stat.Assists
// 		totalRebounds += stat.Rebounds
// 	}

// 	message := fmt.Sprintf("Статистика игрока %s:\nОчки: %d\nПередачи: %d\nПодборы: %d",
// 		player.Name, totalPoints, totalAssists, totalRebounds)

// 	if player.Team != nil {
// 		message += fmt.Sprintf("\nКоманда: %s", player.Team.Name)
// 	}

// 	bot.Send(tgbotapi.NewMessage(chatID, message))
// }

func updateProfile(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	var player models.Player
	err := DB.Where("chat_id = ?", chatID).First(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы еще не зарегистрированы. Используйте /register."))
		return
	}

	userStates[userID] = "update_profile"
	bot.Send(tgbotapi.NewMessage(chatID, "Введите новые данные профиля в формате:\nРост (см), Вес (кг), Позиция"))
}

func processUpdateProfile(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	input := strings.Split(msg.Text, ",")
	if len(input) != 3 {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Неверный формат. Попробуйте снова: Рост (см), Вес (кг), Позиция"))
		return
	}

	height, err1 := strconv.Atoi(strings.TrimSpace(input[0]))
	weight, err2 := strconv.Atoi(strings.TrimSpace(input[1]))
	position := strings.TrimSpace(input[2])

	if err1 != nil || err2 != nil || len(position) < 3 {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Неверные данные. Убедитесь, что указаны корректные значения."))
		return
	}

	var player models.Player
	err := DB.Where("chat_id = ?", msg.Chat.ID).First(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка: Профиль не найден."))
		return
	}

	player.Height = height
	player.Weight = weight
	player.Position = position

	DB.Save(&player)
	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Профиль обновлен!"))
}
