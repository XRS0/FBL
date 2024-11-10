// auth_manager.go
package internal

import (
	"sync"
	Models "tgbot/pkg"
)

type AuthManager struct {
	players map[int64]*Models.Player
	mu      sync.RWMutex
}

func NewAuthManager() *AuthManager {
	return &AuthManager{
		players: make(map[int64]*Models.Player),
	}
}

func (am *AuthManager) RegisterUser(telegramID int64, name string) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.players[telegramID]; !exists {
		am.players[telegramID] = &Models.Player{Name: name, TelegramID: telegramID}
		return true
	}
	return false
}

func (am *AuthManager) GetPlayer(telegramID int64) *Models.Player {
	am.mu.RLock()
	defer am.mu.RUnlock()

	return am.players[telegramID]
}

func (am *AuthManager) Authenticate(telegramID int64) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()

	_, exists := am.players[telegramID]
	return exists
}
