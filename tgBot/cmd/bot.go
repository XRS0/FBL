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
		bot.Send(tgbotapi.NewMessage(chatID, "Введите свои данные в формате: Имя, Рост (см), Вес (кг), Позиция, Контакт"))
	case "/profile":
		listProfile(bot, chatID)
	case "/teams":
		listTeams(bot, chatID)
	case "/matches":
		listMatches(bot, chatID)
	case "/create_team":
		createTeam(bot, chatID, int(userID))
	default:
		if strings.HasPrefix(msg.Text, "/join_team") {
			joinTeam(bot, chatID, msg.Text)
		} else if strings.HasPrefix(msg.Text, "/join_match") {
			joinMatch(bot, chatID, msg.Text)
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
		"/join_match - Записаться на матч"
	bot.Send(tgbotapi.NewMessage(chatID, message))
}

// Регистрация игрока
func registerPlayer(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	data := strings.Split(msg.Text, ",")
	if len(data) != 5 {
		bot.Send(tgbotapi.NewMessage(chatID, "Неправильный формат. Используйте формат: Имя, Рост (см), Вес (кг), Позиция, Контакт"))
		return
	}

	height, err := strconv.Atoi(strings.TrimSpace(data[1]))
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Некорректный рост. Укажите число."))
		return
	}

	weight, err := strconv.Atoi(strings.TrimSpace(data[2]))
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Некорректный вес. Укажите число."))
		return
	}

	player := models.Player{
		Name:     strings.TrimSpace(data[0]),
		Height:   height,
		Weight:   weight,
		Position: strings.TrimSpace(data[3]),
		Contact:  strings.TrimSpace(data[4]),
		ChatID:   chatID,
	}

	err = DB.Create(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при регистрации. Попробуйте снова."))
		log.Printf("Ошибка при регистрации игрока: %v", err)
		return
	}
	userStates[userID] = ""
	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Игрок %s успешно зарегистрирован!", player.Name)))
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
