
package db

import (
	"basketball-league/config"
	"basketball-league/internal/models"
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=UTC",
		cfg.Host, cfg.User_DB, cfg.PasswordDB, cfg.DBName)
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	// Попытка удалить ограничение, если оно существует.
	if DB.Migrator().HasConstraint("match_statistics", "uni_match_statistics_match_id") {
		err = DB.Migrator().DropConstraint("match_statistics", "uni_match_statistics_match_id")
		if err != nil {
			log.Fatalf("Ошибка удаления ограничения: %v", err)
		}
	} else {
		log.Println("Ограничение uni_match_statistics_match_id не найдено, пропуск удаления")
	}

	// Выполняем миграцию. Если возникает ошибка, связанную с отсутствием ограничения,
	// игнорируем её.
	err = DB.AutoMigrate(&models.Team{}, &models.Player{}, &models.Match{}, &models.MatchStatistics{})
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "42704") {
			log.Printf("Предупреждение миграции (ограничение отсутствует): %v", err)
		} else {
			log.Fatalf("Ошибка миграции базы данных: %v", err)
		}
	}
	return DB
}

