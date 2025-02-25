package main

import (
	"log"

	"basketball-league/config"
	wsh "basketball-league/internal/WSH"
	dbpkg "basketball-league/internal/db"
	"basketball-league/internal/bot"
)

func main() {
	// Инициализация конфигурации
	cfg, err := config.InitConfig()
	if err != nil {
		log.Println("Ошибка инициализации конфигурации:", err)
		return
	}

	// Инициализация базы данных
	DB := dbpkg.InitDatabase(cfg)

	// Запуск веб-сервера 
	go wsh.StartWS(DB, *cfg)

	// Создаем и запускаем бота
	tgBot, err := bot.NewBot(cfg, DB)
	if err != nil {
		log.Fatalf("Не удалось инициализировать бота: %v", err)
	}
	tgBot.Run()
}
