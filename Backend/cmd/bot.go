package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"basketball-league/config"
	dbpkg "basketball-league/internal/db"
	"basketball-league/internal/models"
	mtH "basketball-league/internal/matchHandlers"
	owH "basketball-league/internal/ownerHandlers"
	tmH "basketball-league/internal/teamHandlers"
	usH "basketball-league/internal/userHandlers"
	wsh "basketball-league/internal/WSH"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// HandlersConfig группирует обработчики сущностей.
type HandlersConfig struct {
	MatchHandler mtH.Handler
	TeamHandler  tmH.Handler
	UserHandler  usH.Handler
	OwnerHandler owH.Handler
}

// Bot инкапсулирует работу Telegram-бота, его состояние и обработчики.
type Bot struct {
	API        *tgbotapi.BotAPI
	Config     *config.Config
	DB         *gorm.DB
	Handlers   HandlersConfig
	UserStates map[int64]string
	TempData   map[int64]map[string]string
}

// NewBot создаёт и инициализирует нового бота.
func NewBot(cfg *config.Config, db *gorm.DB) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TgApiToken)
	if err != nil {
		return nil, err
	}
	api.Debug = true

	bot := &Bot{
		API:        api,
		Config:     cfg,
		DB:         db,
		UserStates: make(map[int64]string),
		TempData:   make(map[int64]map[string]string),
	}

	bot.initHandlers()
	return bot, nil
}

// initHandlers инициализирует обработчики, передавая им зависимости.
func (b *Bot) initHandlers() {
	handler := models.Handler{DB: b.DB, Bot: b.API}
	b.Handlers.MatchHandler = mtH.Handler{Handler: handler}
	b.Handlers.TeamHandler = tmH.Handler{Handler: handler}
	b.Handlers.UserHandler = usH.Handler{Handler: handler}
	b.Handlers.OwnerHandler = owH.Handler{Handler: handler}
}

// Run запускает цикл получения и обработки обновлений.
func (b *Bot) Run() {
	log.Printf("Авторизация выполнена на аккаунте %s", b.API.Self.UserName)
	updates := b.API.GetUpdatesChan(tgbotapi.NewUpdate(0))

	for update := range updates {
		if update.Message != nil {
			go b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			go b.handleCallbackQuery(update.CallbackQuery)
		}
	}
}

// handleMessage обрабатывает входящие сообщения.
func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	userID := msg.From.ID
	if state, ok := b.UserStates[userID]; ok {
		b.handleStateMessage(msg, state)
	} else {
		b.processCommand(msg)
	}
}

// handleStateMessage маршрутизует сообщения в зависимости от состояния пользователя.
func (b *Bot) handleStateMessage(msg *tgbotapi.Message, state string) {
	switch state {
	case "register_name", "register_patronymic", "register_last_name",
		"register_height", "register_weight", "register_position", "register_contact":
		b.Handlers.UserHandler.RegisterPlayer(b.TempData, b.UserStates, msg)
	case "update_name", "update_patronymic", "update_surname",
		"update_height", "update_weight", "update_position":
		b.Handlers.UserHandler.UpdatePlayer(msg, b.UserStates, b.TempData)
	case "create_team_name":
		b.Handlers.TeamHandler.CreateTeamName(msg.Chat.ID, msg, int(msg.From.ID), b.UserStates)
	case "join_team":
		b.Handlers.TeamHandler.JoinTeam(msg.Chat.ID, msg.Text, b.UserStates)
	default:
		// Если состояние неизвестно, сбрасываем его и обрабатываем сообщение как команду.
		delete(b.UserStates, msg.From.ID)
		b.processCommand(msg)
	}
}

// processCommand обрабатывает команды, отправленные без активного состояния.
func (b *Bot) processCommand(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID
	parts := strings.SplitN(msg.Text, " ", 2)
	command := parts[0]
	// Переменная args убрана, так как не используется.

	switch command {
	case "/start":
		b.sendStartMessage(chatID)
	case "/register":
		b.UserStates[userID] = "register"
		b.Handlers.UserHandler.RegisterPlayer(b.TempData, b.UserStates, msg)
	case "/profile":
		b.Handlers.UserHandler.ListProfile(chatID)
	case "/update_profile":
		b.Handlers.UserHandler.UpdatePlayer(msg, b.UserStates, b.TempData)
	case "/teams":
		b.Handlers.TeamHandler.ListTeams(chatID)
	case "/create_team":
		b.Handlers.TeamHandler.CreateTeam(chatID, int(userID), b.UserStates)
	case "/logout":
		b.Handlers.UserHandler.Logout(b.TempData, b.UserStates, chatID, userID)
	// Добавьте сюда другие команды по необходимости.
	default:
		if strings.HasPrefix(msg.Text, "/join_team") {
			b.Handlers.TeamHandler.JoinTeam(chatID, msg.Text, b.UserStates)
		} else {
			b.sendMessage(chatID, "Неизвестная команда. Попробуйте /start.")
		}
	}
}

// handleCallbackQuery обрабатывает callback-запросы.
func (b *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	userID := query.From.ID
	data := query.Data

	switch {
	case strings.HasPrefix(data, "confirm_remove:"):
		b.processConfirmation(query, chatID, userID, data)
	case strings.HasPrefix(data, "execute_remove:"):
		b.executePlayerRemoval(query, chatID, userID, data)
	case data == "cancel_remove":
		b.cancelRemoval(query, chatID)
	}
}

// processConfirmation обрабатывает подтверждение удаления игрока.
func (b *Bot) processConfirmation(query *tgbotapi.CallbackQuery, chatID int64, userID int64, data string) {
	parts := strings.Split(data, ":")
	if len(parts) != 3 {
		b.sendMessage(chatID, "❌ Ошибка в данных запроса")
		return
	}

	teamID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный формат ID команды")
		return
	}
	number, err := strconv.ParseUint(parts[2], 10, 8)
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный формат номера игрока")
		return
	}

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Подтвердить", fmt.Sprintf("execute_remove:%d:%d", teamID, number)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel_remove"),
		),
	)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		chatID,
		query.Message.MessageID,
		"Вы уверены, что хотите удалить игрока?",
		markup,
	)
	b.API.Send(editMsg)
}

// executePlayerRemoval выполняет удаление игрока.
func (b *Bot) executePlayerRemoval(query *tgbotapi.CallbackQuery, chatID int64, userID int64, data string) {
	parts := strings.Split(data, ":")
	teamID, _ := strconv.Atoi(parts[1])
	number, _ := strconv.ParseUint(parts[2], 10, 8)

	if err := b.Handlers.TeamHandler.RemovePlayerByNumber(userID, teamID, uint8(number)); err != nil {
		b.sendMessage(chatID, "❌ Ошибка: "+err.Error())
	} else {
		b.sendMessage(chatID, fmt.Sprintf("✅ Игрок №%d успешно удалён", number))
	}
	b.deleteMessage(query)
}

// cancelRemoval обрабатывает отмену удаления.
func (b *Bot) cancelRemoval(query *tgbotapi.CallbackQuery, chatID int64) {
	b.sendMessage(chatID, "❌ Удаление отменено")
	b.deleteMessage(query)
}

// sendStartMessage отправляет приветственное сообщение.
func (b *Bot) sendStartMessage(chatID int64) {
	message := `Добро пожаловать!
Вот список доступных команд:
/profile - Просмотреть свой профиль
/update_profile - Обновить профиль
/register - Зарегистрироваться как игрок
/teams - Просмотреть команды и вступить
/players - Просмотреть игроков определенной команды
/matches - Просмотреть матчи
/create_team - Создать свою команду
/join_team - Вступить в команду
/join_match - Записаться на матч
/statistics - Просмотреть статистику матчей
/start - Получить справку по командам
/logout - Выйти из аккаунта

Для администраторов:
/create_stat - Создать статистику матча
/get_stat <ID> - Получить статистику матча по ID
/delete_stat <ID> - Удалить статистику матча
/create_match <Team1ID> <Team2ID> <Date> <Location> - Создать новый матч
/get_match <ID> - Получить информацию о матче по ID
/delete_match <ID> - Удалить матч
/update_match <ID> <Team1ID> <Team2ID> <Date> <Location> - Обновить матч
`
	b.sendMessage(chatID, message)
}

// sendMessage отправляет сообщение в чат.
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	b.API.Send(msg)
}

// deleteMessage удаляет сообщение и отвечает на callback-запрос.
func (b *Bot) deleteMessage(query *tgbotapi.CallbackQuery) {
	delMsg := tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID)
	b.API.Send(delMsg)
	callbackCfg := tgbotapi.NewCallback(query.ID, "")
	if _, err := b.API.Request(callbackCfg); err != nil {
		log.Println("Ошибка при ответе на callback:", err)
	}
}

func main() {
	// Инициализация конфигурации.
	cfg, err := config.InitConfig()
	if err != nil {
		log.Println("Ошибка инициализации конфигурации:", err)
		return
	}

	// Инициализация базы данных.
	DB := dbpkg.InitDatabase(cfg)

	// Запуск веб-сокета.
	go wsh.StartWS(DB)

	// Создаем и запускаем бота.
	bot, err := NewBot(cfg, DB)
	if err != nil {
		log.Fatalf("Не удалось инициализировать бота: %v", err)
	}
	bot.Run()
}

