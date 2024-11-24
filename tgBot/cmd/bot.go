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
var userStates = make(map[int64]string) // Хранение состояния для каждого пользователя

func initDatabase() {
	dsn := "host=localhost user=admin password=password dbname=basketball_league port=5432 sslmode=disable TimeZone=UTC"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(&models.Team{}, &models.Player{}, &models.Match{}, &models.MatchStatistics{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}

func main() {
	initDatabase()

	bot, err := tgbotapi.NewBotAPI("7945815181:AAHAzN3QI5dUtq7iSmw9if2rrA5Rzi2j3bY")
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

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

	// Проверка текущего состояния пользователя
	state, exists := userStates[userID]
	if !exists {
		state = ""
	}

	if state == "register" {
		registerPlayer(bot, msg)
		return
	} else if state == "create_team_name" {
		createTeamName(bot, chatID, msg, int(userID))
		return
	}

	if msg.Text == "/start" {
		sendStartMessage(bot, chatID)
	} else if msg.Text == "/register" {
		userStates[userID] = "register"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите свои данные в формате: Имя, Рост (см), Вес (кг), Позиция, Контакт"))
	} else if msg.Text == "/teams" {
		listTeams(bot, chatID)
	} else if msg.Text == "/matches" {
		listMatches(bot, chatID)
	} else if strings.HasPrefix(msg.Text, "/join_team") {
		joinTeam(bot, chatID, msg.Text)
	} else if strings.HasPrefix(msg.Text, "/join_match") {
		joinMatch(bot, chatID, msg.Text)
	} else if msg.Text == "/create_team" {
		createTeam(bot, chatID, int(userID))
	} else {
		bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда. Попробуйте /start."))
	}
}

func sendStartMessage(bot *tgbotapi.BotAPI, chatID int64) {
	message := "Добро пожаловать! Вот список доступных команд:\n"
	message += "/register - Зарегистрироваться как игрок\n"
	message += "/teams - Просмотреть команды и вступить\n"
	message += "/matches - Записаться на матч\n"
	message += "/create_team - Создать свою команду\n"

	bot.Send(tgbotapi.NewMessage(chatID, message))
}

func registerPlayer(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	data := strings.Split(msg.Text, ",")
	if len(data) != 5 {
		bot.Send(tgbotapi.NewMessage(chatID, "Неправильный формат. Пожалуйста, используйте формат: Имя, Рост (см), Вес (кг), Позиция, Контакт"))
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
		log.Printf("Failed to register player: %v", err)
		return
	}

	userStates[userID] = ""
	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Игрок %s успешно зарегистрирован!", player.Name)))
}

func listTeams(bot *tgbotapi.BotAPI, chatID int64) {
	var teams []models.Team
	err := DB.Where("is_active = ?", true).Find(&teams).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Не удалось получить список команд. Попробуйте позже."))
		log.Printf("Failed to fetch teams: %v", err)
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

func listMatches(bot *tgbotapi.BotAPI, chatID int64) {
	var matches []models.Match
	err := DB.Find(&matches).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Не удалось получить список матчей. Попробуйте позже."))
		log.Printf("Failed to fetch matches: %v", err)
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

// Функция для вступления в команду
func joinTeam(bot *tgbotapi.BotAPI, chatID int64, text string) {
	// Извлекаем ID команды из текста сообщения
	parts := strings.Split(text, " ")
	if len(parts) < 2 {
		bot.Send(tgbotapi.NewMessage(chatID, "Укажите ID команды для вступления. Например: /join_team 1"))
		return
	}

	teamID, err := strconv.Atoi(parts[1])
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Некорректный ID команды."))
		return
	}

	var player models.Player
	err = DB.Where("chat_id = ?", chatID).First(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Игрок не найден. Зарегистрируйтесь с помощью /register."))
		return
	}

	var team models.Team
	err = DB.First(&team, teamID).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Команда с таким ID не найдена."))
		return
	}

	tid := uint(teamID)
	player.TeamID = &tid
	err = DB.Save(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Не удалось вступить в команду. Попробуйте позже."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы успешно вступили в команду %s!", team.Name)))
}

// Функция для записи на матч
func joinMatch(bot *tgbotapi.BotAPI, chatID int64, text string) {
	// Извлекаем ID матча из текста
	parts := strings.Split(text, " ")
	if len(parts) < 2 {
		bot.Send(tgbotapi.NewMessage(chatID, "Укажите ID матча для записи. Например: /join_match 1"))
		return
	}

	matchID, err := strconv.Atoi(parts[1])
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Некорректный ID матча."))
		return
	}

	var player models.Player
	err = DB.Where("chat_id = ?", chatID).First(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Игрок не найден. Зарегистрируйтесь с помощью /register."))
		return
	}

	var match models.Match
	err = DB.First(&match, matchID).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Матч с таким ID не найден."))
		return
	}

	// Добавление игрока к матчу
	matchStats := models.MatchStatistics{
		MatchID: uint(match.ID),
		TeamID:  *player.TeamID, // Возможно, здесь потребуется дополнительная логика для записи команды
	}

	err = DB.Create(&matchStats).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Не удалось записаться на матч."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы успешно записались на матч #%d!", match.ID)))
}

// Функция для создания команды
func createTeam(bot *tgbotapi.BotAPI, chatID int64, userID int) {
	// Проверим, есть ли уже команда у игрока
	var player models.Player
	err := DB.Where("chat_id = ?", chatID).First(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Игрок не найден. Зарегистрируйтесь с помощью /register."))
		return
	}

	// Проверим, создал ли уже этот игрок команду
	var existingTeam models.Team
	err = DB.Where("owner_id = ?", player.ID).First(&existingTeam).Error
	if err == nil {
		bot.Send(tgbotapi.NewMessage(chatID, "У вас уже есть команда!"))
		return
	}

	// Запросим название команды у игрока
	bot.Send(tgbotapi.NewMessage(chatID, "Введите название команды:"))
	userStates[int64(userID)] = "create_team_name"
}

func createTeamName(bot *tgbotapi.BotAPI, chatID int64, msg *tgbotapi.Message, userID int) {
	teamName := msg.Text
	if len(teamName) < 3 {
		bot.Send(tgbotapi.NewMessage(chatID, "Название команды должно быть длиннее 3 символов. Попробуйте снова."))
		return
	}

	var player models.Player
	err := DB.Where("chat_id = ?", chatID).First(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при создании команды. Попробуйте снова."))
		return
	}

	// Создаём команду
	team := models.Team{
		Name:     teamName,
		OwnerID:  player.ID,
		Owner:    player,
		IsActive: true,
	}
	err = DB.Create(&team).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при создании команды. Попробуйте снова."))
		return
	}

	userStates[int64(userID)] = ""
	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Команда %s успешно создана!", team.Name)))
}

// Функция для добавления игрока в команду
func addPlayerToTeam(bot *tgbotapi.BotAPI, chatID int64, userID int, text string) {
	// Извлекаем ID команды из сообщения
	parts := strings.Split(text, " ")
	if len(parts) < 2 {
		bot.Send(tgbotapi.NewMessage(chatID, "Укажите ID команды для добавления игрока. Например: /add_player_to_team 1"))
		return
	}

	teamID, err := strconv.Atoi(parts[1])
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Некорректный ID команды."))
		return
	}

	// Проверим, существует ли команда с таким ID
	var team models.Team
	err = DB.First(&team, teamID).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Команда с таким ID не найдена."))
		return
	}

	// Проверим, является ли игрок владельцем этой команды
	var player models.Player
	err = DB.Where("chat_id = ?", chatID).First(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Игрок не найден. Зарегистрируйтесь с помощью /register."))
		return
	}

	if player.ID != team.OwnerID {
		bot.Send(tgbotapi.NewMessage(chatID, "Только владелец команды может добавлять игроков в команду."))
		return
	}

	// Запросим ID игрока для добавления в команду
	bot.Send(tgbotapi.NewMessage(chatID, "Введите Telegram ID игрока для добавления в команду:"))
	userStates[int64(userID)] = "add_player_to_team"
}

func addPlayerToTeamHandler(bot *tgbotapi.BotAPI, chatID int64, msg *tgbotapi.Message, userID int) {
	// Получаем Telegram ID игрока
	parts := strings.Split(msg.Text, " ")
	if len(parts) < 1 {
		bot.Send(tgbotapi.NewMessage(chatID, "Некорректный Telegram ID игрока."))
		return
	}

	playerID, err := strconv.Atoi(parts[0])
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Некорректный Telegram ID игрока."))
		return
	}

	var player models.Player
	err = DB.Where("chat_id = ?", playerID).First(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Игрок с таким Telegram ID не найден."))
		return
	}

	// Добавляем игрока в команду
	var team models.Team
	err = DB.Where("owner_id = ?", player.ID).First(&team).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при добавлении игрока в команду."))
		return
	}

	// Обновляем команду игрока
	player.TeamID = &team.ID
	err = DB.Save(&player).Error
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Не удалось добавить игрока в команду. Попробуйте позже."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Игрок с Telegram ID %d добавлен в команду %s!", playerID, team.Name)))
}
