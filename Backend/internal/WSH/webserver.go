package wsh

import (
	mtH "basketball-league/internal/matchHandlers"
	"basketball-league/internal/models"
	tmH "basketball-league/internal/teamHandlers"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type MatchResponse struct {
	Time       string `json:"time"`
	Team1Score int    `json:"team1_score"`
	Team1Name  string `json:"team1_name"`
	Team2Name  string `json:"team2_name"`
	Team2Score int    `json:"team2_score"`
	Status     string `json:"status"`
	Location   string `json:"loc"`
}

type TeamStatistics struct {
	Name   string `json:"name"`
	Games  int    `json:"games"`
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
	Points int    `json:"points"`
}

func ServeMatchesHandler(db *gorm.DB) http.HandlerFunc {
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

			// Определяем статус матча
			var status string
			if stats.Team1Score == 0 && stats.Team2Score == 0 {
				status = "Идет регистрация"
			} else {
				status = "Завершен"
			}
			var ht = &tmH.Handler{
				mainHandler,
			}
			// Получение информации о командах

			team1 := ht.GetTeamByID(int(stats.TeamID1))
			team2 := ht.GetTeamByID(int(stats.TeamID2))

			fmt.Println("DKSJFLDJSFJSLDJFLKDSJFKDJFLDSJFLKDSJFLSDJFLDSJFLKDSJFDJFKDSJFLJDK")
			// Проверяем, что команды существуют
			var team1Name, team2Name string
			if team1 != nil {
				team1Name = team1.Name
			} else {
				team1Name = "Неизвестная команда"
				log.Printf("Команда с ID %d не найдена", stats.TeamID1)
			}

			if team2 != nil {
				team2Name = team2.Name
			} else {
				team2Name = "Неизвестная команда"
				log.Printf("Команда с ID %d не найдена", stats.TeamID2)
			}

			var hm = &mtH.Handler{
				mainHandler,
			}

			fmt.Println(hm)

			var matchDetails = hm.GetMatchByID(int(stats.MatchID))

			fmt.Println(matchDetails)
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

func ServeStatisticsHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var teams []models.Team
		err := db.Find(&teams).Error
		if err != nil {
			http.Error(w, "Ошибка при получении статистики команд: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var results []TeamStatistics
		for _, team := range teams {
			var games, wins, losses, points int

			var matches []models.Match
			db.Where("team1_id = ? OR team2_id = ?", team.ID, team.ID).Find(&matches)

			for _, match := range matches {
				var stats models.MatchStatistics
				db.Where("match_id = ?", match.ID).First(&stats)

				games++
				if (int(match.Team1ID) == team.ID && stats.Team1Score > stats.Team2Score) ||
					(int(match.Team2ID) == team.ID && stats.Team2Score > stats.Team1Score) {
					wins++
					points += 3
				} else if stats.Team1Score != stats.Team2Score {
					losses++
				}
			}

			if wins-losses < 0 {
				points = 0
			} else {
				points = wins - losses
			}

			results = append(results, TeamStatistics{
				Name:   team.Name,
				Games:  games,
				Wins:   wins,
				Losses: losses,
				Points: wins - points,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Обработка предварительных запросов CORS (OPTIONS)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		h(w, r)
	}
}

func StartWS(DB *gorm.DB) {
	http.HandleFunc("/matches", withCORS(ServeMatchesHandler(DB)))
	http.HandleFunc("/statistics", withCORS(ServeStatisticsHandler(DB)))

	log.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
