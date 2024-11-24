// package api

// import (
// 	"basketball-league/internal/models"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"gorm.io/gorm"
// )

// func SetupRoutes(router *gin.Engine, db *gorm.DB) {
// 	// Регистрация игрока
// 	router.POST("/players", func(c *gin.Context) {
// 		var player models.Player
// 		if err := c.ShouldBindJSON(&player); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}
// 		if err := db.Create(&player).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register player"})
// 			return
// 		}
// 		c.JSON(http.StatusOK, player)
// 	})

// 	// Вступление в команду
// 	router.POST("/teams/:id/join", func(c *gin.Context) {
// 		var player models.Player
// 		teamID := c.Param("id")

// 		if err := c.ShouldBindJSON(&player); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		// Обновление команды игрока
// 		if err := db.Model(&models.Player{}).Where("id = ?", player.ID).Update("team_id", teamID).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join team"})
// 			return
// 		}
// 		c.JSON(http.StatusOK, gin.H{"message": "Successfully joined the team"})
// 	})

// 	// Запись на матч
// 	router.POST("/matches/:id/join", func(c *gin.Context) {
// 		var stats models.MatchStatistics
// 		matchID := c.Param("id")

// 		if err := c.ShouldBindJSON(&stats); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		stats.MatchID = uint(matchID)
// 		if err := db.Create(&stats).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join match"})
// 			return
// 		}
// 		c.JSON(http.StatusOK, gin.H{"message": "Successfully joined the match"})
// 	})
// }

// func RunServer(db *gorm.DB) {
// 	router := gin.Default()
// 	SetupRoutes(router, db)

// 	router.GET("/ping", func(c *gin.Context) {
// 		c.JSON(http.StatusOK, gin.H{"message": "pong"})
// 	})

// 	router.Run(":8080")
// }
