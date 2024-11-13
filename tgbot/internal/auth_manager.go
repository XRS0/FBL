// auth_manager.go
package internal

import (
	"sync"
	. "tgbot/pkg"
)

type AuthManager struct {
	players map[int64]*Player
	mu      sync.Mutex
}

func NewAuthManager() *AuthManager {
	return &AuthManager{players: make(map[int64]*Player)}
}

// RegisterUser добавляет нового пользователя с паролем
func (am *AuthManager) RegisterUser(telegramID int64, name, password string) *Player {
	am.mu.Lock()
	defer am.mu.Unlock()

	player := NewPlayer(name, telegramID, password)
	am.players[telegramID] = player
	return player
}

// Authenticate проверяет, зарегистрирован ли пользователь и совпадает ли пароль
func (am *AuthManager) Authenticate(telegramID int64, name, password string) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	player, exists := am.players[telegramID]
	return exists && player.Password == password
}

// GetPlayer возвращает игрока по его Telegram ID
func (am *AuthManager) GetPlayer(telegramID int64) *Player {
	am.mu.Lock()
	defer am.mu.Unlock()

	return am.players[telegramID]
}

// IsRegistered проверяет, зарегистрирован ли пользователь
func (am *AuthManager) IsRegistered(telegramID int64) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	_, exists := am.players[telegramID]
	return exists
}
