#ifndef TG_BOT_H
#define TG_BOT_H

#include <tgbot/tgbot.h>
#include <memory>
#include "auth_manager.h"  // Подключаем auth_manager
#include <unordered_map>

class TelegramBot {
public:
    explicit TelegramBot(const std::string& token);

    void start();
    void authenticate_user(TgBot::Message::Ptr message);
    void send_welcome_message(TgBot::Message::Ptr message);

private:
    void register_user(int64_t telegram_id, const std::string& name);
    void send_registration_message(TgBot::Message::Ptr message);

    TgBot::Bot bot;
    std::shared_ptr<AuthManager> auth_manager;  // Добавляем поле для auth_manager
};

#endif // TG_BOT_H