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
	. "basketball-league/internal/tempDataHandlers"
	. "basketball-league/internal/userHandlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

var DB *gorm.DB
var userStates = make(map[int64]string)
var temporaryData = make(map[int64]map[string]string)

var admins = map[int64]bool{
	1324977667: true,
	984866387:  true,
	1655151699: true,
}

// Проверка, является ли пользователь администратором
func isAdmin(chatID int64) bool {
	return admins[chatID]
}

func main() {
	DB = InitDatabase()
	go StartWS(DB)

	//bot, err := tgbotapi.NewBotAPI("7945815181:AAHAzN3QI5dUtq7iSmw9if2rrA5Rzi2j3bY")
	bot, err := tgbotapi.NewBotAPI("6942168243:AAGtBiMeTWDtHJNxeCkqT2SnA1qSHMQTimI")
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

	// Получаем текущее состояние пользователя
	state, exists := userStates[userID]

	if !exists {
		// Если состояния нет, проверяем команду
		processCommand(bot, msg, chatID, userID)
		return
	}

	// Обработка состояний
	switch state {
	case "register_name", "register_patronymic", "register_last_name", "register_height", "register_weight", "register_position", "register_contact":
		registerPlayer(bot, msg, DB)
	case "update_name", "update_patronymic", "update_surname", "update_height", "update_weight", "update_position":
		UpdatePlayer(bot, msg, DB, userStates, temporaryData)
	case "create_team_name":
		CreateTeamName(bot, chatID, msg, int(userID), userStates, DB)
	case "rename_team":
		RenameTeam(bot, msg, DB, userStates)
	default:
		// Если состояние неизвестно, сбрасываем его
		delete(userStates, userID)
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
		ListProfile(bot, chatID, DB)
	case "/update_profile":
		UpdatePlayer(bot, msg, DB, userStates, temporaryData)
	case "/teams":
		ListTeams(bot, chatID, DB)
	//case "/rename_team":
	//	RenameTeam(bot, msg, DB, userStates)
	case "/create_team":
		CreateTeam(bot, chatID, int(userID), userStates, DB)
	case "/logout":
		logout(bot, chatID, userID)
	case "/players":
		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте формат: /players ИМЯКОМАНДЫ,"+
				" либо /players_all	если хотите получить всех игроков без команды"))
		} else {
			teamName := strings.TrimSpace(commandParts[1])
			ListPlayersByTeam(bot, chatID, teamName, DB)
		}
	case "/players_all":
		ListPlayersWithoutTeam(bot, chatID, DB)
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
		date, err := time.Parse(time.DateTime, fmt.Sprint(args[2]+" "+args[3]))
		if err != nil {
			fmt.Println("Пошел ты нахер козел")
		}
		location := args[4]
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
		match := GetMatchByID(DB, matchID)
		if match == nil {
			bot.Send(tgbotapi.NewMessage(chatID, "матч не найден"))
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

		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"Счет команды %s - %v,\n"+
				"Счет команды %s - %v\n",
			GetTeamByID(DB, int(stat.TeamID1)).Name, stat.Team1Score,
			GetTeamByID(DB, int(stat.TeamID1)).Name, stat.Team2Score,
		)))

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
		"- /update_profile - Обновить профиль\n" +
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
		SetTemporaryData(userID, "first_name", msg.Text, temporaryData)
		userStates[userID] = "register_patronymic"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваше отчество:"))

	case "register_patronymic":
		if len(msg.Text) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Отчество должно быть длиннее 1 символа. Попробуйте снова."))
			return
		}
		SetTemporaryData(userID, "patronymic", msg.Text, temporaryData)
		userStates[userID] = "register_last_name"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите вашу фамилию:"))

	case "register_last_name":
		if len(msg.Text) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Фамилия должна быть длиннее 1 символа. Попробуйте снова."))
			return
		}
		SetTemporaryData(userID, "last_name", msg.Text, temporaryData)

		tempData := GetTemporaryData(userID, temporaryData)
		fullName := fmt.Sprintf("%s %s %s", tempData["first_name"], tempData["patronymic"], tempData["last_name"])
		SetTemporaryData(userID, "name", fullName, temporaryData)

		userStates[userID] = "register_height"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваш рост (см):"))

	case "register_height":
		height, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || height < 100 || height > 250 {
			bot.Send(tgbotapi.NewMessage(chatID, "Укажите корректный рост в сантиметрах (от 100 до 250)."))
			return
		}
		SetTemporaryData(userID, "height", strconv.Itoa(height), temporaryData)
		userStates[userID] = "register_weight"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваш вес (кг):"))

	case "register_weight":
		weight, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || weight < 30 || weight > 200 {
			bot.Send(tgbotapi.NewMessage(chatID, "Укажите корректный вес в килограммах (от 30 до 200)."))
			return
		}
		SetTemporaryData(userID, "weight", strconv.Itoa(weight), temporaryData)
		userStates[userID] = "register_position"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите вашу игровую позицию (например, Центровой, Разыгрывающий):"))

	case "register_position":
		if len(msg.Text) < 3 {
			bot.Send(tgbotapi.NewMessage(chatID, "Позиция должна содержать хотя бы 3 символа. Попробуйте снова."))
			return
		}
		SetTemporaryData(userID, "position", msg.Text, temporaryData)
		userStates[userID] = "register_contact"

		// Клавиатура для отправки контакта
		contactButton := tgbotapi.NewKeyboardButtonContact("Поделиться контактом")
		replyMarkup := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(contactButton),
		)
		message := tgbotapi.NewMessage(chatID, "Поделитесь вашим контактом через кнопку ниже:")
		message.ReplyMarkup = replyMarkup
		bot.Send(message)

	case "register_contact":
		// Проверка на наличие контакта в сообщении
		if msg.Contact == nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, используйте кнопку 'Поделиться контактом'."))
			return
		}

		SetTemporaryData(userID, "contact", fmt.Sprintf("Номер телефона - %s\nTgID - @%s", msg.Contact.PhoneNumber, msg.From.UserName), temporaryData)

		// Создание игрока в базе данных
		tempData := GetTemporaryData(userID, temporaryData)
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
		DeleteTemporaryData(userID, temporaryData)
		delete(userStates, userID)

		// Удаляем клавиатуру после завершения
		msgClearKeyboard := tgbotapi.NewMessage(chatID, fmt.Sprintf("Игрок %s успешно зарегистрирован!", player.Name))
		msgClearKeyboard.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		bot.Send(msgClearKeyboard)

	default:
		userStates[userID] = "register_name"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите ваше имя:"))
	}
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
	DeleteTemporaryData(userID, temporaryData)

	bot.Send(tgbotapi.NewMessage(chatID, "Вы успешно вышли из аккаунта. Для повторной регистрации используйте команду /register."))
}

// Просмотр профиля
func ListProfile(bot *tgbotapi.BotAPI, chatID int64, DB *gorm.DB) {
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
