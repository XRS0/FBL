package db

import (
	"basketball-league/internal/models"
  "basketball-league/config"
	"log"
  "fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase(cfg *config.Config) *gorm.DB {
	var DB *gorm.DB
  dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=UTC",
    cfg.Host, cfg.User_DB, cfg.PasswordDB, cfg.DBName)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	err = DB.AutoMigrate(&models.Team{}, &models.Player{}, &models.Match{}, &models.MatchStatistics{})
	if err != nil {
		log.Fatalf("Ошибка миграции базы данных: %v", err)
	}
	return DB
}
