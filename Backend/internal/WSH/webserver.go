package wsh

import (
  "io"
  "fmt"
  "bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"basketball-league/config"
	mtH "basketball-league/internal/matchHandlers"
	"basketball-league/internal/models"
	tmH "basketball-league/internal/teamHandlers"

	"gorm.io/gorm"
)

// MatchResponse – структура ответа для матчей.
type MatchResponse struct {
	Time       string `json:"time"`
	Team1Score int    `json:"team1_score"`
	Team1Name  string `json:"team1_name"`
	Team2Name  string `json:"team2_name"`
	Team2Score int    `json:"team2_score"`
	Status     string `json:"status"`
	Location   string `json:"loc"`
}

// TeamResponse – структура ответа для команды в endpoint /teams_data.
type TeamResponse struct {
	Logo    string          `json:"logo"`    // Ссылка на логотип
	Name    string          `json:"name"`    // Название команды
	Games   int             `json:"games"`   // Количество игр
	Wins    int             `json:"wins"`    // Победы
	Loses   int             `json:"loses"`   // Поражения
	Points  int             `json:"points"`  // Очки
	Players []PlayerSummary `json:"players"` // Список игроков
	Captain string          `json:"captain"` // Имя капитана
}

// PlayerSummary – краткая информация об игроке.
type PlayerSummary struct {
	Name   string `json:"name"`   // Имя игрока (можно добавить фамилию)
	Avatar string `json:"avatar"` // Ссылка на аватар (если есть)
	Number uint8  `json:"number"` // Номер игрока
}

// TeamStatistics – структура для endpoint статистики (если требуется отдельно).
type TeamStatistics struct {
	Name   string `json:"name"`
	Games  int    `json:"games"`
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
	Points int    `json:"points"`
}

// ServeMatchesHandler возвращает список матчей.
func ServeMatchesHandler(db *gorm.DB) http.HandlerFunc {
	// mainHandler используется для вызова методов из обработчика матчей/команд.
	var mainHandler = models.Handler{DB: db}
	if db == nil {
		log.Fatal("Объект базы данных не инициализирован!")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var matches []models.Match
		err := db.Preload("Team1").Preload("Team2").Find(&matches).Error
		if err != nil {
			http.Error(w, "Ошибка при получении матчей: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var results []MatchResponse
		for _, match := range matches {
			var stats models.MatchStatistics
			err := db.Where("match_id = ?", match.ID).First(&stats).Error
			if err != nil {
				log.Printf("Статистика для матча %d не найдена: %v", match.ID, err)
				continue
			}

			// Определяем статус матча.
			var status string
			if stats.Team1Score == 0 && stats.Team2Score == 0 {
				status = "Еще не прошел"
			} else {
				status = "Завершен"
			}

			// Используем обработчик команд для получения информации о командах.
			var ht = &tmH.Handler{Handler: mainHandler}
			team1 := ht.GetTeamByID(int(stats.TeamID1))
			team2 := ht.GetTeamByID(int(stats.TeamID2))
			team1Name, team2Name := "Неизвестная команда", "Неизвестная команда"
			if team1 != nil {
				team1Name = team1.Name
			} else {
				log.Printf("Команда с ID %d не найдена", stats.TeamID1)
			}
			if team2 != nil {
				team2Name = team2.Name
			} else {
				log.Printf("Команда с ID %d не найдена", stats.TeamID2)
			}

			// Получаем данные о матче через обработчик матчей.
			var hm = &mtH.Handler{Handler: mainHandler}
			var matchDetails = hm.GetMatchByID(int(stats.MatchID))
			location := "Неизвестная локация"
			if matchDetails != nil {
				location = matchDetails.Location
			} else {
				log.Printf("Матч с ID %d не найден", stats.MatchID)
			}

			results = append(results, MatchResponse{
				Time:       match.Date.Format(time.DateTime),
				Team1Score: stats.Team1Score,
				Team2Score: stats.Team2Score,
				Team1Name:  team1Name,
				Team2Name:  team2Name,
				Location:   location,
				Status:     status,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

// ServeTeamsDataHandler возвращает данные о командах в требуемом формате.
func ServeTeamsDataHandler(db *gorm.DB, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var teams []models.Team
		// Предзагружаем игроков и владельца (Owner) для каждой команды.
		if err := db.Preload("Players").Preload("Owner").Find(&teams).Error; err != nil {
			http.Error(w, "Ошибка при получении команд: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var results []TeamResponse
		for _, team := range teams {
			// Вычисляем статистику матчей для команды.
			var matches []models.Match
			if err := db.Where("team1_id = ? OR team2_id = ?", team.ID, team.ID).Find(&matches).Error; err != nil {
				http.Error(w, "Ошибка при получении матчей для команды: "+err.Error(), http.StatusInternalServerError)
				return
			}
			games := len(matches)
			wins, losses := 0, 0
			for _, match := range matches {
				var stats models.MatchStatistics
				if err := db.Where("match_id = ?", match.ID).First(&stats).Error; err != nil {
					continue
				}
				// Определяем победу или поражение.
				if (match.Team1ID == uint(team.ID) && stats.Team1Score > stats.Team2Score) ||
					(match.Team2ID == uint(team.ID) && stats.Team2Score > stats.Team1Score) {
					wins++
				} else if stats.Team1Score != stats.Team2Score {
					losses++
				}
			}
			points := wins // расчёт очков

			// Формируем список игроков.
			var players []PlayerSummary
			for _, player := range team.Players {
				players = append(players, PlayerSummary{
					Name:   player.Name,
					Avatar: "", // Здесь можно добавить ссылку на аватар, если она есть
					Number: player.Number,
				})
			}

			captain := ""
			if team.Owner != nil {
				captain = team.Owner.Name
			}

      var tr TeamResponse

      if team.PathToLogo != "" {
        tr = TeamResponse{
				  Logo:    team.PathToLogo, // Значение по умолчанию для логотипа
				  Name:    team.Name,
				  Games:   games,
				  Wins:    wins,
				  Loses:   losses,
				  Points:  points,
				  Players: players,
				  Captain: captain,
			  }
      } else {
        tr = TeamResponse{
				  Logo:    getTeamLogoLink(team.Name, cfg), // Значение по умолчанию для логотипа
				  Name:    team.Name,
				  Games:   games,
				  Wins:    wins,
				  Loses:   losses,
				  Points:  points,
				  Players: players,
				  Captain: captain,
			  }
      }

			results = append(results, tr)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}
  

// ServeStatisticsHandler возвращает статистику команд в формате TeamStatistics.
func ServeStatisticsHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var teams []models.Team
		if err := db.Find(&teams).Error; err != nil {
			http.Error(w, "Ошибка при получении статистики команд: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var results []TeamStatistics
		for _, team := range teams {
			var games, wins, losses int

			var matches []models.Match
			db.Where("team1_id = ? OR team2_id = ?", team.ID, team.ID).Find(&matches)
			for _, match := range matches {
				var stats models.MatchStatistics
				db.Where("match_id = ?", match.ID).First(&stats)
				games++
				if (match.Team1ID == uint(team.ID) && stats.Team1Score > stats.Team2Score) ||
					(match.Team2ID == uint(team.ID) && stats.Team2Score > stats.Team1Score) {
					wins++
				} else if stats.Team1Score != stats.Team2Score {
					losses++
				}
			}
			// расчёт очков.
			points := wins
			if points < 0 {
				points = 0
			}
			results = append(results, TeamStatistics{
				Name:   team.Name,
				Games:  games,
				Wins:   wins,
				Losses: losses,
				Points: points,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

// getTeamLogoLink отправляет запрос на создание ссылки на логотип команды
func getTeamLogoLink(teamName string, cfg config.Config) string {
	requestBody := map[string]string{
		"clientId": cfg.FSClientID,
		"filePath": fmt.Sprintf("%s/logo.png", teamName),
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("Ошибка при сериализации JSON: %v\n", err)
		return ""
	}

	url := fmt.Sprintf("%s/filelink", cfg.FSHost)
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

// withCORS добавляет заголовки CORS, разрешая все запросы.
func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		h(w, r)
	}
}


// StartWS регистрирует обработчики и запускает HTTP-сервер.
func StartWS(DB *gorm.DB, cfg config.Config) {
	http.HandleFunc("/matches", withCORS(ServeMatchesHandler(DB)))
	http.HandleFunc("/teams_data", withCORS(ServeTeamsDataHandler(DB, cfg)))
	http.HandleFunc("/statistics", withCORS(ServeStatisticsHandler(DB)))

	log.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

