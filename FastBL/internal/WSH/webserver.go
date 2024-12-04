package wsh

import (
	"basketball-league/internal/models"
	"encoding/json"
	"log"
	"net/http"

	"gorm.io/gorm"
)

// MatchResponse описывает формат данных для матча.
type MatchResponse struct {
	Time       string `json:"time"`
	Team1Score int    `json:"team1_score"`
	Team2Score int    `json:"team2_score"`
	Status     string `json:"status"`
}

// TeamStatistics описывает формат данных для статистики команд.
type TeamStatistics struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	Games        int    `json:"games"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	Points       int    `json:"points"`
}

// ServeMatchesHandler отправляет данные о матчах.
func ServeMatchesHandler(db *gorm.DB) http.HandlerFunc {
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
			db.Where("match_id = ?", match.ID).First(&stats)

			// Определяем статус матча
			var status string
			if stats.Team1Score == 0 && stats.Team2Score == 0 {
				status = "Идет регистрация"
			} else {
				status = "Завершен"
			}

			results = append(results, MatchResponse{
				Time:       match.Date.Format("15:04"),
				Team1Score: stats.Team1Score,
				Team2Score: stats.Team2Score,
				Status:     status,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

// ServeStatisticsHandler отправляет данные о статистике команд.
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
			// Подсчет статистики для команды
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

			results = append(results, TeamStatistics{
				Name:         team.Name,
				Abbreviation: "TB",
				Games:        games,
				Wins:         wins,
				Losses:       losses,
				Points:       points,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

func StartWS(DB *gorm.DB) {
	http.HandleFunc("/matches", ServeMatchesHandler(DB))
	http.HandleFunc("/statistics", ServeStatisticsHandler(DB))

	// Запуск сервера
	log.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
