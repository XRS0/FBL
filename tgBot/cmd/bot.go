package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"basketball-league/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var userStates = make(map[int64]string) // Хранение состояния пользователей

// Инициализация базы данных
func initDatabase() {
	dsn := "host=localhost user=admin password=password dbname=basketball_league port=5432 sslmode=disable TimeZone=UTC"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	err = DB.AutoMigrate(&models.Team{}, &models.Player{}, &models.Match{}, &models.MatchStatistics{})
	if err != nil {
		log.Fatalf("Ошибка миграции базы данных: %v", err)
	}
}

// Главная функция
func main() {
	initDatabase()

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

// Обработка сообщений
func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	// Получение текущего состояния пользователя
	state := userStates[userID]

	switch state {
	case "register":
		registerPlayer(bot, msg)
	case "create_team_name":
		createTeamName(bot, chatID, msg, int(userID))
	default:
		processCommand(bot, msg, chatID, userID)
	}
}

// Обработка команд
func processCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, chatID int64, userID int64) {
	switch msg.Text {
	case "/start":
		sendStartMessage(bot, chatID)
	case "/register":
		userStates[userID] = "register"
		registerPlayer(bot, msg)
	case "/profile":
		listProfile(bot, chatID)
	case "/teams":
		listTeams(bot, chatID)
	case "/matches":
		listMatches(bot, chatID)
	case "/create_team":
		createTeam(bot, chatID, int(userID))
	case "/logout":
		logout(bot, chatID, userID)
	default:
		if strings.HasPrefix(msg.Text, "/join_team") {
			joinTeam(bot, chatID, msg.Text)
		} else if strings.HasPrefix(msg.Text, "/join_match") {
			joinMatch(bot, chatID, msg.Text)
		} else if strings.HasPrefix(userStates[userID], "register") {
			registerPlayer(bot, msg)
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда. Попробуйте /start."))
		}
	}
}

// Отправка приветственного сообщения
func sendStartMessage(bot *tgbotapi.BotAPI, chatID int64) {
	message := "Добро пожаловать! Вот список доступных команд:\n" +
		"/profile - Просмотреть свой профиль\n" +
		"/register - Зарегистрироваться как игрок\n" +
		"/teams - Просмотреть команды и вступить\n" +
		"/matches - Просмотреть матчи\n" +
		"/create_team - Создать свою команду\n" +
		"/join_team - Вступить в команду\n" +
		"/join_match - Записаться на матч" +
		"/logout - Выйти из аккаунта"
	bot.Send(tgbotapi.NewMessage(chatID, message))
}

func registerPlayer(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
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
func listProfile(bot *tgbotapi.BotAPI, chatID int64) {
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

// Просмотр списка команд
func listTeams(bot *tgbotapi.BotAPI, chatID int64) {
	var teams []models.Team

	err := DB.Where("is_active = ?", true).Find(&teams).Error
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
		message += fmt.Sprintf("- %s (Владелец: %s)\n", team.Name, team.Owner.Name)
	}
	bot.Send(tgbotapi.NewMessage(chatID, message))
}

// Просмотр списка матчей
func listMatches(bot *tgbotapi.BotAPI, chatID int64) {
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
		message += fmt.Sprintf("- Матч #%d: %s vs %s (%s)\n", match.ID, match.Team1ID, match.Team2ID, match.Location)
	}
	bot.Send(tgbotapi.NewMessage(chatID, message))
}

func createTeamName(bot *tgbotapi.BotAPI, chatID int64, msg *tgbotapi.Message, userID int) {
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
		IsActive: true,
	}

	if err := DB.Create(&team).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка создания команды. Попробуйте позже."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Команда '%s' успешно создана!", team.Name)))
	userStates[int64(userID)] = ""
}

// Вступление в команду
func joinTeam(bot *tgbotapi.BotAPI, chatID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		bot.Send(tgbotapi.NewMessage(chatID, "Используйте формат: /join_team ID_команды"))
		return
	}
	teamID, err := strconv.Atoi(parts[1])
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "ID команды должно быть числом."))
		return
	}

	var team models.Team
	if err := DB.First(&team, teamID).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Команда не найдена. Проверьте ID команды."))
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

// Запись на матч
func joinMatch(bot *tgbotapi.BotAPI, chatID int64, text string) {
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

// Функция для создания команды
func createTeam(bot *tgbotapi.BotAPI, chatID int64, userID int) {

	var owner models.Player
	if err := DB.Where("chat_id = ?", chatID).First(&owner).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы не зарегистрированы как игрок. Сначала используйте /register."))
		return
	}

	userStates[int64(userID)] = "create_team_name"
}
