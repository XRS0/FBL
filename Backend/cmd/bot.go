package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"basketball-league/config"
	wsh "basketball-league/internal/WSH"
	. "basketball-league/internal/db"
	mtH "basketball-league/internal/matchHandlers"
	"basketball-league/internal/models"
	tmH "basketball-league/internal/teamHandlers"
	usH "basketball-league/internal/userHandlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlersConfig struct {
	mhHandler mtH.Handler
	tmHandler tmH.Handler
	usHandler usH.Handler
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

	var DB = InitDatabase(cfg)
	Handler.mhHandler.Handler, Handler.tmHandler.Handler, Handler.usHandler.Handler = models.Handler{DB: DB}, models.Handler{DB: DB}, models.Handler{DB: DB}
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

	for update := range updates {
		if update.Message != nil {
			go handleMessage(bot, update.Message)
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
		Handler.usHandler.RegisterPlayer(bot, temporaryData, userStates, msg)
	case "update_name", "update_patronymic", "update_surname", "update_height", "update_weight", "update_position":
		Handler.usHandler.UpdatePlayer(bot, msg, userStates, temporaryData)
	case "create_team_name":
		Handler.tmHandler.CreateTeamName(bot, chatID, msg, int(userID), userStates)
	case "rename_team":
		Handler.tmHandler.RenameTeam(bot, msg, userStates)
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
		Handler.usHandler.RegisterPlayer(bot, temporaryData, userStates, msg)
	case "/profile":
		Handler.usHandler.ListProfile(bot, chatID)
	case "/update_profile":
		Handler.usHandler.UpdatePlayer(bot, msg, userStates, temporaryData)
	case "/teams":
		Handler.tmHandler.ListTeams(bot, chatID)
	//case "/rename_team":
	//	RenameTeam(bot, msg, DB, userStates)
	case "/create_team":
		Handler.tmHandler.CreateTeam(bot, chatID, int(userID), userStates)
	case "/logout":
		Handler.usHandler.Logout(bot, temporaryData, userStates, chatID, userID)
	case "/players":
		if len(commandParts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Используйте формат: /players ИМЯКОМАНДЫ,"+
				" либо /players_all	если хотите получить всех игроков без команды"))
		} else {
			teamName := strings.TrimSpace(commandParts[1])
			Handler.tmHandler.ListPlayersByTeam(bot, chatID, teamName)
		}
	case "/players_all":
		Handler.tmHandler.ListPlayersWithoutTeam(bot, chatID)
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
			Handler.tmHandler.JoinTeam(bot, chatID, msg.Text)
		} else if strings.HasPrefix(userStates[userID], "register") {
			Handler.usHandler.RegisterPlayer(bot, temporaryData, userStates, msg)
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
