// package main

// import (
// 	"fmt"
// 	"log"

// 	"basketball-league/internal/api"
// 	"basketball-league/internal/models"

// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// var DB *gorm.DB

// func initDatabase() {
// 	dsn := "host=localhost user=admin password=password dbname=basketball_league port=5432 sslmode=disable TimeZone=UTC"
// 	var err error
// 	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatalf("Failed to connect to database: %v", err)
// 	}

// 	package api

// import (
// 	"log"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"gorm.io/gorm"
// )

// func RunServer(db *gorm.DB) {
// 	router := gin.Default()

// 	// Пример маршрута
// 	router.GET("/ping", func(c *gin.Context) {
// 		c.JSON(http.StatusOK, gin.H{"message": "pong"})
// 	})

// 	log.Println("API server running on port 8080")
// 	err := router.Run(":8080")
// 	if err != nil {
// 		log.Fatalf("Failed to start server: %v", err)
// 	}
// }
