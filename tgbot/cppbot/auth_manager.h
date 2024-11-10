#ifndef AUTH_MANAGER_H
#define AUTH_MANAGER_H

#include <unordered_map>
#include <string>
#include <memory>
#include "models.h"  // Подключаем модель Player

// Класс для управления аутентификацией пользователей
class AuthManager {
public:
    // Метод для регистрации нового пользователя
    bool register_user(int64_t telegram_id, const std::string& name);

    // Метод для получения игрока по его ID
    std::shared_ptr<Player> get_player(int64_t telegram_id) const;

    // Метод для аутентификации пользователя (проверка, зарегистрирован ли он)
    bool authenticate(int64_t telegram_id) const;

private:
    // Словарь для хранения пользователей по их telegram_id
    std::unordered_map<int64_t, std::shared_ptr<Player>> players;
};

#endif // AUTH_MANAGER_H