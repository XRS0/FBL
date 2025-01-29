package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"basketball-league/config"
	wsh "basketball-league/internal/WSH"
	dbpkg "basketball-league/internal/db"
	mtH "basketball-league/internal/matchHandlers"
	"basketball-league/internal/models"
	owH "basketball-league/internal/ownerHandlers"
	tmH "basketball-league/internal/teamHandlers"
	usH "basketball-league/internal/userHandlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlersConfig struct {
	mhHandler mtH.Handler
	tmHandler tmH.Handler
	usHandler usH.Handler
  owHandler owH.Handler
}

var cfg config.Config

var Handler HandlersConfig
var userStates = make(map[int64]string)
var temporaryData = make(map[int64]map[string]string)

func isAdmin(chatID int64, cfg config.Config) bool {
  for i := 0; i < len(cfg.Admins); i++ {
    if cfg.Admins[i] == chatID {
      return true
    }
  }
  return false
}

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Println(err)
    return;
	}

	var DB = dbpkg.InitDatabase(cfg)

  go wsh.StartWS(DB)

	bot, err := tgbotapi.NewBotAPI(cfg.TgApiToken)
	if err != nil {
		log.Fatalf("Не удалось инициализировать бота: %v", err)
	}

	bot.Debug = true
	log.Printf("Авторизация выполнена на аккаунте %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

  Handler.mhHandler.Handler = models.Handler{DB: DB, Bot: bot}
  Handler.tmHandler.Handler = models.Handler{DB: DB, Bot: bot}
  Handler.usHandler.Handler = models.Handler{DB: DB, Bot: bot}
  Handler.owHandler.Handler = models.Handler{DB: DB, Bot: bot}
 
	for update := range updates {
		if update.Message != nil {
			go handleMessage(bot, update.Message)
		} else if update.CallbackQuery != nil {
      go handleCallbackQuery(bot, update.CallbackQuery)
    }
	}
}

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	state, exists := userStates[userID]

	if !exists {
		processCommand(bot, msg, chatID, userID)
		return
	}

	// Обработка состояний
	switch state {
	case "register_name", "register_patronymic", "register_last_name", "register_height", "register_weight", "register_position", "register_contact":
		Handler.usHandler.RegisterPlayer(temporaryData, userStates, msg)
	case "update_name", "update_patronymic", "update_surname", "update_height", "update_weight", "update_position":
		Handler.usHandler.UpdatePlayer(msg, userStates, temporaryData)
	case "create_team_name":
		Handler.tmHandler.CreateTeamName(chatID, msg, int(userID), userStates)
	case "join_team":
    Handler.tmHandler.JoinTeam(chatID, msg.Text, userStates)
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
		Handler.usHandler.RegisterPlayer(temporaryData, userStates, msg)
	case "/profile":
		Handler.usHandler.ListProfile(chatID)
	case "/update_profile":
		Handler.usHandler.UpdatePlayer(msg, userStates, temporaryData)
	case "/teams":
		Handler.tmHandler.ListTeams(chatID)
	case "/create_team":
		Handler.tmHandler.CreateTeam(chatID, int(userID), userStates)
	case "/logout":
		Handler.usHandler.Logout(temporaryData, userStates, chatID, userID)
	case "/players":
		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте формат: /players ИМЯКОМАНДЫ,"+
				" либо /players_all	если хотите получить всех игроков без команды"))
		} else {
			teamName := strings.TrimSpace(commandParts[1])
			Handler.tmHandler.ListPlayersByTeam(chatID, teamName)
		}
	case "/players_all":
		Handler.tmHandler.ListPlayersWithoutTeam(chatID)
  case "/remove_player":
    if len(commandParts) < 2 {
        bot.Send(tgbotapi.NewMessage(chatID, "ℹ️ Используйте: /remove_player <номер>"))
        return
    }
    numStr := strings.TrimSpace(commandParts[1])
    num, err := strconv.ParseUint(numStr, 10, 8)
    if err != nil {
        bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный формат номера"))
        return
    }
    number := uint8(num)

    var owner models.Player
    if err := Handler.tmHandler.DB.Where("chat_id = ?", userID).First(&owner).Error; err != nil {
        bot.Send(tgbotapi.NewMessage(chatID, "❌ Вы не зарегистрированы как игрок"))
        return
    }

    var teams []models.Team
    if err := Handler.tmHandler.DB.Where("owner_id = ?", owner.ID).Find(&teams).Error; err != nil || len(teams) == 0 {
        bot.Send(tgbotapi.NewMessage(chatID, "❌ У вас нет команд"))
        return
    }

    if len(teams) > 1 {
        sendTeamSelectionMenu(bot, chatID, teams, number)
    } else {
        processPlayerRemoval(bot, chatID, userID, teams[0].ID, number)
    }
	case "/create_match":
		if !isAdmin(chatID, cfg) {
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
			fmt.Println("Матч не создан, че то пошло не так")
		}
		location := args[4]
		match, err := Handler.mhHandler.CreateMatch(uint(team1ID), uint(team2ID), date, location)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка создания матча: "+err.Error()))
			return
		}
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Матч создан: #%d", match.ID)))

	case "/get_match":
		if !isAdmin(chatID, cfg) {
			bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для выполнения этой команды."))
			return
		}
		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте: /get_match <ID>"))
			return
		}
		matchID, _ := strconv.Atoi(commandParts[1])
		match := Handler.mhHandler.GetMatchByID(matchID)
		if match == nil {
			bot.Send(tgbotapi.NewMessage(chatID, "матч не найден"))
			return
		}
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Матч #%d: %s vs %s в %s", match.ID, match.Team1.Name, match.Team2.Name, match.Location)))

	case "/delete_match":
		if !isAdmin(chatID, cfg) {
			bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для выполнения этой команды."))
			return
		}
		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте: /delete_match <ID>"))
			return
		}
		matchID, _ := strconv.Atoi(commandParts[1])
		err := Handler.mhHandler.DeleteMatch(matchID)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка удаления матча: "+err.Error()))
			return
		}
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Матч #%d успешно удален.", matchID)))

	case "/create_stat":
		if !isAdmin(chatID, cfg) {
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

		stat, err := Handler.mhHandler.CreateMatchStatistics(uint(matchID), uint(teamID1), uint(teamID2), team1Score, team2Score)
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

		stat, err := Handler.mhHandler.GetStatisticsByMatchID(uint(matchID))
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка: %v", err)))
			return
		}

		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"Счет команды %s - %v,\n"+
				"Счет команды %s - %v\n",
			Handler.tmHandler.GetTeamByID(int(stat.TeamID1)).Name, stat.Team1Score,
			Handler.tmHandler.GetTeamByID(int(stat.TeamID1)).Name, stat.Team2Score,
		)))

	case "/delete_stat":
		if !isAdmin(chatID, cfg) {
			bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для выполнения этой команды."))
			return
		}
		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте: /delete_stat <ID>"))
			return
		}
		response := Handler.mhHandler.DeleteMatchStatistic(commandParts[1])
		bot.Send(tgbotapi.NewMessage(chatID, response))

	default:
		if strings.HasPrefix(msg.Text, "/join_team") {
			Handler.tmHandler.JoinTeam(chatID, msg.Text, userStates)
		} else if strings.HasPrefix(userStates[userID], "register") {
			Handler.usHandler.RegisterPlayer(temporaryData, userStates, msg)
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

func sendTeamSelectionMenu(bot *tgbotapi.BotAPI, chatID int64, teams []models.Team, number uint8) {
	var buttons []tgbotapi.InlineKeyboardButton
	for _, team := range teams {
		callbackData := fmt.Sprintf("confirm_remove:%d:%d", team.ID, number)
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(team.Name, callbackData))
	}

	markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
	msg := tgbotapi.NewMessage(chatID, "Выберите команду для удаления игрока:")
	msg.ReplyMarkup = markup
	bot.Send(msg)
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	data := query.Data
	chatID := query.Message.Chat.ID
	userID := query.From.ID

	switch {
	case strings.HasPrefix(data, "confirm_remove:"):
		processConfirmation(bot, query, chatID, userID, data)
	case strings.HasPrefix(data, "execute_remove:"):
		executePlayerRemoval(bot, query, chatID, userID, data)
	case data == "cancel_remove":
		cancelRemoval(bot, query, chatID)
	}
}

func processConfirmation(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, chatID int64, userID int64, data string) {
	parts := strings.Split(data, ":")
	if len(parts) != 3 {
		sendError(bot, chatID, "❌ Ошибка в данных запроса")
		return
	}

	teamID, err := strconv.Atoi(parts[1])
	if err != nil {
		sendError(bot, chatID, "❌ Неверный формат ID команды")
		return
	}

	number, err := strconv.ParseUint(parts[2], 10, 8)
	if err != nil {
		sendError(bot, chatID, "❌ Неверный формат номера игрока")
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
	bot.Send(editMsg)
}

func executePlayerRemoval(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, chatID int64, userID int64, data string) {
	parts := strings.Split(data, ":")
	teamID, _ := strconv.Atoi(parts[1])
	number, _ := strconv.ParseUint(parts[2], 10, 8)

	if err := Handler.tmHandler.RemovePlayerByNumber(userID, teamID, uint8(number)); err != nil {
		sendError(bot, chatID, "❌ Ошибка: "+err.Error())
	} else {
		sendSuccess(bot, chatID, fmt.Sprintf("✅ Игрок №%d успешно удалён", number))
	}
	deleteMessage(bot, query)
}

func cancelRemoval(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, chatID int64) {
	sendSuccess(bot, chatID, "❌ Удаление отменено")
	deleteMessage(bot, query)
}

func processPlayerRemoval(bot *tgbotapi.BotAPI, chatID int64, userID int64, teamID int, number uint8) {
	if err := Handler.tmHandler.RemovePlayerByNumber(userID, teamID, number); err != nil {
		sendError(bot, chatID, "❌ Ошибка: "+err.Error())
	} else {
		sendSuccess(bot, chatID, fmt.Sprintf("✅ Игрок №%d успешно удалён", number))
	}
}

func sendError(bot *tgbotapi.BotAPI, chatID int64, message string) {
	bot.Send(tgbotapi.NewMessage(chatID, message))
}

func sendSuccess(bot *tgbotapi.BotAPI, chatID int64, message string) {
	bot.Send(tgbotapi.NewMessage(chatID, message))
}


func deleteMessage(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
    delMsg := tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID)
    bot.Send(delMsg)
    
    // Используем AnswerCallbackQuery с CallbackConfig
    callbackCfg := tgbotapi.NewCallback(query.ID, "")
    if _, err := bot.Request(callbackCfg); err != nil {
        log.Println("Ошибка при ответе на callback:", err)
    }
}


func handleRemovePlayerCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, chatID, userID int64, commandParts []string) {
	var owner models.Player
	if err := Handler.tmHandler.DB.Where("chat_id = ?", userID).First(&owner).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Вы не зарегистрированы как игрок"))
		return
	}

	var teams []models.Team
	if err := Handler.tmHandler.DB.Where("owner_id = ?", owner.ID).Find(&teams).Error; err != nil || len(teams) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ У вас нет команд"))
		return
	}

	if len(commandParts) < 2 {
		bot.Send(tgbotapi.NewMessage(chatID, "ℹ️ Используйте: /remove_player <номер>"))
		return
	}

	num, err := strconv.ParseUint(strings.Split(commandParts[1], " ")[1], 10, 8)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный формат номера"))
		return
	}
	number := uint8(num)

	if len(teams) > 1 {
		sendTeamSelectionMenu(bot, chatID, teams, number)
	} else {
		processPlayerRemoval(bot, chatID, userID, teams[0].ID, number)
	}
}
