#include <unordered_map>
#include <string>
#include <memory>
#include "models.h"

class AuthManager {
public:
    bool register_user(int64_t telegram_id, const std::string& name);
    std::shared_ptr<Player> get_player(int64_t telegram_id) const;
    bool authenticate(int64_t telegram_id) const;

private:
    std::unordered_map<int64_t, std::shared_ptr<Player>> players;
};

bool AuthManager::register_user(int64_t telegram_id, const std::string& name) {
    if (players.find(telegram_id) == players.end()) {
        auto player = std::make_shared<Player>(name, telegram_id);
        players[telegram_id] = player;
        return true;
    }
    return false;
}

std::shared_ptr<Player> AuthManager::get_player(int64_t telegram_id) const {
    auto it = players.find(telegram_id);
    return (it != players.end()) ? it->second : nullptr;
}

bool AuthManager::authenticate(int64_t telegram_id) const {
    return players.find(telegram_id) != players.end();
}