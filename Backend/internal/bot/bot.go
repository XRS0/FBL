package bot

import (
	"fmt"
  "time"
	"log"
	"strconv"
	"strings"
  "net/http"
  "encoding/json"
  "io"
  "mime/multipart"
  "bytes"

	"basketball-league/config"
	mtH "basketball-league/internal/matchHandlers"
	"basketball-league/internal/models"
	owH "basketball-league/internal/ownerHandlers"
	tmH "basketball-league/internal/teamHandlers"
	usH "basketball-league/internal/userHandlers"

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
  chatID := msg.Chat.ID
  userID := msg.From.ID

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
  case "awaiting_team_photo":
	  b.processTeamPhoto(msg)
  case "create_match_date":
        const layout = "2006-01-02 15:04:05"
        _, err := time.Parse(layout, msg.Text)
        if err != nil {
            b.sendMessage(chatID, "Неверный формат даты. Попробуйте еще раз (YYYY-MM-DD HH:MM:SS):")
            return
        }
        b.TempData[userID]["create_match_date"] = msg.Text
        b.UserStates[userID] = "create_match_location"
        b.sendMessage(chatID, "Введите место проведения матча:")
  case "create_match_location":
        b.TempData[userID]["create_match_location"] = msg.Text
        team1ID, _ := strconv.Atoi(b.TempData[userID]["create_match_team1"])
        team2ID, _ := strconv.Atoi(b.TempData[userID]["create_match_team2"])
        date, _ := time.Parse("2006-01-02 15:04:05", b.TempData[userID]["create_match_date"])
        match, err := b.Handlers.MatchHandler.CreateMatch(uint(team1ID), uint(team2ID), date, msg.Text)
        if err != nil {
            b.sendMessage(chatID, "Ошибка создания матча: "+err.Error())
        } else {
            b.sendMessage(chatID, fmt.Sprintf("Матч создан: #%d", match.ID))
        }
        // Очистка состояния
        delete(b.UserStates, userID)
        delete(b.TempData[userID], "create_match_team1")
        delete(b.TempData[userID], "create_match_team2")
        delete(b.TempData[userID], "create_match_date")
        delete(b.TempData[userID], "create_match_location")
  case "create_stat_team1score":
        if _, err := strconv.Atoi(msg.Text); err != nil {
            b.sendMessage(chatID, "Неверный формат счета. Введите число:")
            return
        }
        b.TempData[userID]["create_stat_team1score"] = msg.Text
        // Получаем название второй команды из матча
        mID, _ := strconv.Atoi(b.TempData[userID]["create_stat_match"])
        match := b.Handlers.MatchHandler.GetMatchByID(mID)
        if match == nil {
            b.sendMessage(chatID, "Матч не найден")
            delete(b.UserStates, userID)
            return
        }
        b.UserStates[userID] = "create_stat_team2score"
        b.sendMessage(chatID, fmt.Sprintf("Введите счет для команды %s:", match.Team2.Name))
  case "create_stat_team2score":
        if _, err := strconv.Atoi(msg.Text); err != nil {
            b.sendMessage(chatID, "Неверный формат счета. Введите число:")
            return
        }
        b.TempData[userID]["create_stat_team2score"] = msg.Text
        matchID, _ := strconv.Atoi(b.TempData[userID]["create_stat_match"])
        team1ID, _ := strconv.Atoi(b.TempData[userID]["create_stat_team1"])
        team2ID, _ := strconv.Atoi(b.TempData[userID]["create_stat_team2"])
        team1Score, _ := strconv.Atoi(b.TempData[userID]["create_stat_team1score"])
        team2Score, _ := strconv.Atoi(b.TempData[userID]["create_stat_team2score"])
        stat, err := b.Handlers.MatchHandler.CreateMatchStatistics(uint(matchID), uint(team1ID), uint(team2ID), team1Score, team2Score)
        if err != nil {
            b.sendMessage(chatID, fmt.Sprintf("Ошибка создания статистики: %v", err))
        } else {
            b.sendMessage(chatID, fmt.Sprintf("Статистика успешно создана для матча #%d", stat.ID))
        }
        // Очистка состояния
        delete(b.UserStates, userID)
        delete(b.TempData[userID], "create_stat_match")
        delete(b.TempData[userID], "create_stat_team1")
        delete(b.TempData[userID], "create_stat_team2")
        delete(b.TempData[userID], "create_stat_team1score")
        delete(b.TempData[userID], "create_stat_team2score")
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
  isadmin := b.isAdmin(chatID)

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
	case "/players":
		// Ожидается формат: /players <имя команды>
		if len(parts) < 2 {
			b.sendMessage(chatID, "Используйте формат: /players ИМЯКОМАНДЫ, либо /players_all если хотите получить всех игроков без команды")
		} else {
			teamName := strings.TrimSpace(parts[1])
			b.Handlers.TeamHandler.ListPlayersByTeam(chatID, teamName, isadmin)
		}
	case "/players_all":
		b.Handlers.TeamHandler.ListPlayersWithoutTeam(chatID, isadmin)
	case "/remove_player":
		// Формат: /remove_player <номер>
		if len(parts) < 2 {
			b.sendMessage(chatID, "ℹ️ Используйте: /remove_player <номер>")
			return
		}
		numStr := strings.TrimSpace(parts[1])
		num, err := strconv.ParseUint(numStr, 10, 8)
		if err != nil {
			b.sendMessage(chatID, "❌ Неверный формат номера")
			return
		}
		number := uint8(num)

		// Получаем владельца игрока из БД.
		var owner models.Player
		if err := b.Handlers.TeamHandler.Handler.DB.Where("chat_id = ?", userID).First(&owner).Error; err != nil {
			b.sendMessage(chatID, "❌ Вы не зарегистрированы как игрок")
			return
		}

		// Получаем команды, где владелец является хозяином.
		var teams []models.Team
		if err := b.Handlers.TeamHandler.Handler.DB.Where("owner_id = ?", owner.ID).Find(&teams).Error; err != nil || len(teams) == 0 {
			b.sendMessage(chatID, "❌ У вас нет команд")
			return
		}
		if len(teams) > 1 {
			b.sendTeamSelectionMenu(chatID, teams, number)
		} else {
			b.processPlayerRemoval(chatID, userID, teams[0].ID, number)
		}
	case "/create_match":
    if !b.isAdmin(chatID) {
        b.sendMessage(chatID, "У вас нет прав для выполнения этой команды.")
        return
    }
    // Запускаем интерактивное создание матча:
    b.UserStates[userID] = "create_match_team1"
    b.sendTeamSelectionForMatchCreation(chatID, "Выберите первую команду:", "create_match:team1:")
	case "/get_match":
		if !b.isAdmin(chatID) {
			b.sendMessage(chatID, "У вас нет прав для выполнения этой команды.")
			return
		}
		if len(parts) < 2 {
			b.sendMessage(chatID, "Используйте: /get_match <ID>")
			return
		}
		matchID, err := strconv.Atoi(parts[1])
		if err != nil {
			b.sendMessage(chatID, "Неверный формат ID")
			return
		}
		match := b.Handlers.MatchHandler.GetMatchByID(matchID)
		if match == nil {
			b.sendMessage(chatID, "Матч не найден")
			return
		}
		b.sendMessage(chatID, fmt.Sprintf("Матч #%d: %s vs %s в %s", match.ID, match.Team1.Name, match.Team2.Name, match.Location))
	case "/delete_match":
		if !b.isAdmin(chatID) {
			b.sendMessage(chatID, "У вас нет прав для выполнения этой команды.")
			return
		}
		if len(parts) < 2 {
			b.sendMessage(chatID, "Используйте: /delete_match <ID>")
			return
		}
		matchID, err := strconv.Atoi(parts[1])
		if err != nil {
			b.sendMessage(chatID, "Неверный формат ID")
			return
		}
		if err := b.Handlers.MatchHandler.DeleteMatch(matchID); err != nil {
			b.sendMessage(chatID, "Ошибка удаления матча: "+err.Error())
			return
		}
		b.sendMessage(chatID, fmt.Sprintf("Матч #%d успешно удален.", matchID))
	case "/create_stat":
    if !b.isAdmin(chatID) {
        b.sendMessage(chatID, "У вас нет прав для выполнения этой команды.")
        return
    }
    // Запускаем интерактивное создание статистики:
    b.UserStates[userID] = "create_stat_match"
    b.sendMatchSelectionForStat(chatID, "Выберите матч для создания статистики:", "create_stat:match:")
	case "/get_stat":
		if len(parts) < 2 {
			b.sendMessage(chatID, "Используйте: /get_stat <MatchID>")
			return
		}
		matchID, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			b.sendMessage(chatID, "MatchID должен быть числом.")
			return
		}
		stat, err := b.Handlers.MatchHandler.GetStatisticsByMatchID(uint(matchID))
		if err != nil {
			b.sendMessage(chatID, fmt.Sprintf("Ошибка: %v", err))
			return
		}
		// Получаем данные о командах для отображения результата.
		team1 := b.Handlers.TeamHandler.GetTeamByID(int(stat.TeamID1))
		team2 := b.Handlers.TeamHandler.GetTeamByID(int(stat.TeamID2))
		b.sendMessage(chatID, fmt.Sprintf(
			"Счет команды %s - %v,\nСчет команды %s - %v\n",
			team1.Name, stat.Team1Score,
			team2.Name, stat.Team2Score,
		))
	case "/delete_stat":
		if !b.isAdmin(chatID) {
			b.sendMessage(chatID, "У вас нет прав для выполнения этой команды.")
			return
		}
		if len(parts) < 2 {
			b.sendMessage(chatID, "Используйте: /delete_stat <ID>")
			return
		}
		response := b.Handlers.MatchHandler.DeleteMatchStatistic(parts[1])
		b.sendMessage(chatID, response)
  case "/set_team_photo":
		// Доступно только владельцам (капитанам)
		var owner models.Player
		if err := b.Handlers.TeamHandler.Handler.DB.Where("chat_id = ?", userID).First(&owner).Error; err != nil {
			b.sendMessage(chatID, "❌ Вы не зарегистрированы как игрок.")
			return
		}

		var teams []models.Team
		if err := b.Handlers.TeamHandler.Handler.DB.Where("owner_id = ?", owner.ID).Find(&teams).Error; err != nil || len(teams) == 0 {
			b.sendMessage(chatID, "❌ У вас нет команд, которыми вы владеете.")
			return
		}

		// Если несколько команд – предлагаем выбор
		if len(teams) > 1 {
			var buttons []tgbotapi.InlineKeyboardButton
			for _, team := range teams {
				callbackData := fmt.Sprintf("set_team_photo:%d", team.ID)
				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(team.Name, callbackData))
			}
			markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
			msg := tgbotapi.NewMessage(chatID, "Выберите команду, для которой хотите установить фото:")
			msg.ReplyMarkup = markup
			b.API.Send(msg)
		} else {
			// Если только одна команда – сразу переходим к запросу фото
			if b.TempData[userID] == nil {
				b.TempData[userID] = make(map[string]string)
			}
			b.TempData[userID]["team_photo_team_id"] = fmt.Sprintf("%d", teams[0].ID)
			b.UserStates[userID] = "awaiting_team_photo"
			b.sendMessage(chatID, "Пожалуйста, отправьте фотографию для команды.")
		}
	default:
		// Если сообщение начинается с "/join_team", обрабатываем его отдельно.
		if strings.HasPrefix(msg.Text, "/join_team") {
			b.Handlers.TeamHandler.JoinTeam(chatID, msg.Text, b.UserStates)
		} else {
			b.sendMessage(chatID, "Неизвестная команда. Попробуйте /start.")
		}
	}
}

// Отправка клавиатуры для выбора команды при создании матча.
func (b *Bot) sendTeamSelectionForMatchCreation(chatID int64, text, callbackPrefix string) {
    var teams []models.Team
    if err := b.DB.Find(&teams).Error; err != nil {
        b.sendMessage(chatID, "Ошибка получения списка команд")
        return
    }
    var buttons []tgbotapi.InlineKeyboardButton
    for _, team := range teams {
        data := fmt.Sprintf("%s%d", callbackPrefix, team.ID)
        buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(team.Name, data))
    }
    markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ReplyMarkup = markup
    b.API.Send(msg)
}

// Отправка клавиатуры для выбора матча при создании статистики.
func (b *Bot) sendMatchSelectionForStat(chatID int64, text, callbackPrefix string) {
    var matches []models.Match
    // Подгружаем данные о командах для отображения названий
    if err := b.DB.Preload("Team1").Preload("Team2").Find(&matches).Error; err != nil {
        b.sendMessage(chatID, "Ошибка получения списка матчей")
        return
    }
    var buttons []tgbotapi.InlineKeyboardButton
    for _, match := range matches {
        label := fmt.Sprintf("#%d: %s vs %s", match.ID, match.Team1.Name, match.Team2.Name)
        data := fmt.Sprintf("%s%d", callbackPrefix, match.ID)
        buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(label, data))
    }
    markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ReplyMarkup = markup
    b.API.Send(msg)
}

// sendTeamSelectionMenu выводит меню выбора команды для удаления игрока.
func (b *Bot) sendTeamSelectionMenu(chatID int64, teams []models.Team, number uint8) {
	var buttons []tgbotapi.InlineKeyboardButton
	for _, team := range teams {
		callbackData := fmt.Sprintf("confirm_remove:%d:%d", team.ID, number)
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(team.Name, callbackData))
	}
	markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
	msg := tgbotapi.NewMessage(chatID, "Выберите команду для удаления игрока:")
	msg.ReplyMarkup = markup
	b.API.Send(msg)
}

// processPlayerRemoval удаляет игрока из команды.
func (b *Bot) processPlayerRemoval(chatID int64, userID int64, teamID int, number uint8) {
	if err := b.Handlers.TeamHandler.RemovePlayerByNumber(userID, teamID, number); err != nil {
		b.sendMessage(chatID, "❌ Ошибка: "+err.Error())
	} else {
		b.sendMessage(chatID, fmt.Sprintf("✅ Игрок №%d успешно удалён", number))
	}
}

// isAdmin проверяет, является ли пользователь администратором.
func (b *Bot) isAdmin(chatID int64) bool {
	for _, admin := range b.Config.Admins {
		if admin == chatID {
			return true
		}
	}
	return false
}

func (b *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	userID := query.From.ID
	data := query.Data

	switch {
	case strings.HasPrefix(data, "set_team_photo:"):
		b.handleSetTeamPhotoCallback(query)
  case strings.HasPrefix(data, "create_match:"):
    b.handleCreateMatchCallback(query)
  case strings.HasPrefix(data, "create_stat:"):
    b.handleCreateStatCallback(query)
	case strings.HasPrefix(data, "confirm_remove:"):
		b.processConfirmation(query, chatID, userID, data)
	case strings.HasPrefix(data, "execute_remove:"):
		b.executePlayerRemoval(query, chatID, userID, data)
	case data == "cancel_remove":
		b.cancelRemoval(query, chatID)
	}
}

// Обработка выбора команды при создании матча
func (b *Bot) handleCreateMatchCallback(query *tgbotapi.CallbackQuery) {
    chatID := query.Message.Chat.ID
    userID := query.From.ID
    parts := strings.Split(query.Data, ":") // Формат: "create_match:teamX:<ID>"
    if len(parts) < 3 {
        b.sendMessage(chatID, "Неверные данные для создания матча")
        return
    }
    step := parts[1] // "team1" или "team2"
    teamID := parts[2]
    if step == "team1" {
        if b.TempData[userID] == nil {
            b.TempData[userID] = make(map[string]string)
        }
        b.TempData[userID]["create_match_team1"] = teamID
        b.UserStates[userID] = "create_match_team2"
        b.sendTeamSelectionForMatchCreation(chatID, "Выберите вторую команду:", "create_match:team2:")
    } else if step == "team2" {
        if b.TempData[userID] == nil {
            b.TempData[userID] = make(map[string]string)
        }
        b.TempData[userID]["create_match_team2"] = teamID
        b.UserStates[userID] = "create_match_date"
        b.sendMessage(chatID, "Введите дату матча в формате YYYY-MM-DD HH:MM:SS:")
    }
    callbackCfg := tgbotapi.NewCallback(query.ID, "Команда выбрана")
    b.API.Request(callbackCfg)
}

// Обработка выбора матча при создании статистики
func (b *Bot) handleCreateStatCallback(query *tgbotapi.CallbackQuery) {
    chatID := query.Message.Chat.ID
    userID := query.From.ID
    parts := strings.Split(query.Data, ":") // Формат: "create_stat:match:<ID>"
    if len(parts) < 3 {
        b.sendMessage(chatID, "Неверные данные для создания статистики")
        return
    }
    matchID := parts[2]
    if b.TempData[userID] == nil {
        b.TempData[userID] = make(map[string]string)
    }
    b.TempData[userID]["create_stat_match"] = matchID

    // Получаем данные матча, чтобы подставить команды
    mID, err := strconv.Atoi(matchID)
    if err != nil {
        b.sendMessage(chatID, "Неверный формат matchID")
        return
    }
    match := b.Handlers.MatchHandler.GetMatchByID(mID)
    if match == nil {
        b.sendMessage(chatID, "Матч не найден")
        return
    }
    // Сохраняем ID команд из матча
    b.TempData[userID]["create_stat_team1"] = fmt.Sprintf("%d", match.Team1.ID)
    b.TempData[userID]["create_stat_team2"] = fmt.Sprintf("%d", match.Team2.ID)
    // Переходим к вводу счета для первой команды
    b.UserStates[userID] = "create_stat_team1score"
    b.sendMessage(chatID, fmt.Sprintf("Введите счет для команды %s:", match.Team1.Name))
    
    callbackCfg := tgbotapi.NewCallback(query.ID, "Матч выбран")
    b.API.Request(callbackCfg)
}

// handleSetTeamPhotoCallback обрабатывает выбор команды для установки фото.
func (b *Bot) handleSetTeamPhotoCallback(query *tgbotapi.CallbackQuery) {
    chatID := query.Message.Chat.ID
    userID := query.From.ID
    parts := strings.Split(query.Data, ":")
    if len(parts) != 2 {
        b.sendMessage(chatID, "Ошибка выбора команды.")
        return
    }
    teamID := parts[1]
    if b.TempData[userID] == nil {
        b.TempData[userID] = make(map[string]string)
    }
    b.TempData[userID]["team_photo_team_id"] = teamID
    b.UserStates[userID] = "awaiting_team_photo"

    b.sendMessage(chatID, "Отправьте фотографию для выбранной команды.")

    // Ответим на нажатие кнопки, чтобы убрать "часики" у inline-кнопки:
    callbackCfg := tgbotapi.NewCallback(query.ID, "Команда выбрана")
    b.API.Request(callbackCfg)
}


// processTeamPhoto теперь отправляет файл на внешний файловый сервер.
func (b *Bot) processTeamPhoto(msg *tgbotapi.Message) {
    userID := msg.From.ID
    chatID := msg.Chat.ID

    teamIDStr, ok := b.TempData[userID]["team_photo_team_id"]
    if !ok {
        b.sendMessage(chatID, "Не удалось определить команду. Попробуйте снова.")
        delete(b.UserStates, userID)
        return
    }

    // Преобразуем строку teamID в int
    teamIDInt, err := strconv.Atoi(teamIDStr)
    if err != nil {
        b.sendMessage(chatID, "Ошибка определения ID команды.")
        delete(b.UserStates, userID)
        return
    }

    // Предположим, что есть метод GetTeamByID, возвращающий *models.Team
    team := b.Handlers.TeamHandler.GetTeamByID(teamIDInt)
    if team == nil {
        b.sendMessage(chatID, "Команда не найдена.")
        delete(b.UserStates, userID)
        return
    }

    if len(msg.Photo) == 0 {
        b.sendMessage(chatID, "Пожалуйста, отправьте фотографию.")
        return
    }

    // Выбираем фото с наибольшим разрешением
    photo := msg.Photo[len(msg.Photo)-1]

    // Получаем файл через BotAPI
    fileCfg := tgbotapi.FileConfig{FileID: photo.FileID}
    tgFile, err := b.API.GetFile(fileCfg)
    if err != nil {
        b.sendMessage(chatID, "Ошибка получения файла: "+err.Error())
        return
    }

    // Ссылка для скачивания
    fileURL := tgFile.Link(b.API.Token)

    // Скачиваем файл в resp.Body
    resp, err := http.Get(fileURL)
    if err != nil {
        b.sendMessage(chatID, "Ошибка скачивания файла: "+err.Error())
        return
    }
    defer resp.Body.Close()

    // Готовим multipart-запрос к файловому серверу
    var buf bytes.Buffer
    writer := multipart.NewWriter(&buf)

    // 1) clientId
    if err := writer.WriteField("clientId", b.Config.FSClientID); err != nil {
        b.sendMessage(chatID, "Ошибка при формировании запроса (clientId).")
        return
    }

    // 2) filePath: "ИмяКоманды/лого.png"
    //    Если в team.Name пробелы или спец.символы, возможно, стоит их заменять или экранировать.
    filePath := fmt.Sprintf("%s/logo.png", team.Name)
    if err := writer.WriteField("filePath", filePath); err != nil {
        b.sendMessage(chatID, "Ошибка при формировании запроса (filePath).")
        return
    }

    // 3) Файл в поле "file"
    fileFormField, err := writer.CreateFormFile("file", "logo.png")
    if err != nil {
        b.sendMessage(chatID, "Ошибка при формировании multipart: "+err.Error())
        return
    }

    // Копируем содержимое скачанного файла из resp.Body
    if _, err = io.Copy(fileFormField, resp.Body); err != nil {
        b.sendMessage(chatID, "Ошибка копирования файла в буфер: "+err.Error())
        return
    }

    // Закрываем multipart-форму
    if err = writer.Close(); err != nil {
        b.sendMessage(chatID, "Ошибка при закрытии multipart-формы.")
        return
    }

    // Адрес файлового сервера, например http://localhost:8080/upload
    fsURL := fmt.Sprintf("%s/upload", b.Config.FSHost)

    // Формируем POST-запрос
    req, err := http.NewRequest(http.MethodPost, fsURL, &buf)
    if err != nil {
        b.sendMessage(chatID, "Ошибка формирования POST-запроса: "+err.Error())
        return
    }

    // Устанавливаем заголовок Content-Type
    req.Header.Set("Content-Type", writer.FormDataContentType())

    // Выполняем запрос
    httpClient := &http.Client{}
    uploadResp, err := httpClient.Do(req)
    if err != nil {
        b.sendMessage(chatID, "Ошибка при отправке файла на сервер: "+err.Error())
        return
    }
    defer uploadResp.Body.Close()

    if uploadResp.StatusCode != http.StatusOK {
        // Если код не 200 OK — читаем тело, чтобы узнать, что за ошибка
        bodyBytes, _ := io.ReadAll(uploadResp.Body)
        log.Printf("Ошибка при загрузке файла: %s", string(bodyBytes))
        b.sendMessage(chatID, fmt.Sprintf("Ошибка загрузки файла на сервер (код %d).", uploadResp.StatusCode))
        return
    }

    // Успешно
    if err := b.Handlers.TeamHandler.UpdateLogoPath(team.ID, b.getTeamLogoLink(team.Name)); err != nil {
      b.sendMessage(chatID, "Произошла ошибка, попробуйте позже")
    } else {
      b.sendMessage(chatID, fmt.Sprintf("Фото команды «%s» успешно загружено!", team.Name))
    }

    // Сбрасываем состояние
    delete(b.UserStates, userID)
    delete(b.TempData[userID], "team_photo_team_id")
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

// getTeamLogoLink отправляет запрос на создание ссылки на логотип команды
func (b *Bot) getTeamLogoLink(teamName string) string {
	requestBody := map[string]string{
		"clientId": b.Config.FSClientID,
		"filePath": fmt.Sprintf("%s/logo.png", teamName),
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("Ошибка при сериализации JSON: %v\n", err)
		return ""
	}

	url := fmt.Sprintf("%s/filelink", b.Config.FSHost)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка при отправке запроса на %s: %v\n", url, err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Сервер вернул ошибку: %s\n", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Ответ сервера: %s\n", string(body))
		return ""
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("Ошибка при декодировании ответа: %v\n", err)
		return ""
	}

	if success, ok := response["success"].(bool); ok && success {
		if url, ok := response["url"].(string); ok {
			fmt.Println("Ссылка на файл успешно создана!")
			fmt.Printf("URL: %s\n", url)
			return url
		}
		fmt.Println("URL не найден в ответе сервера.")
	} else {
		fmt.Println("Не удалось создать ссылку на файл.")
		fmt.Printf("Ответ сервера: %v\n", response)
	}

	return ""
}

