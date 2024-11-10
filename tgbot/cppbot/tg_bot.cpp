#include "tg_bot.h"
#include "models.h"
#include "auth_manager.h"
#include <tgbot/tgbot.h>
#include <iostream>

TelegramBot::TelegramBot(const std::string& token) 
    : bot(token), auth_manager(std::make_shared<AuthManager>()) {  // Инициализируем AuthManager
    bot.getEvents().onCommand("start", [&](TgBot::Message::Ptr message) {
        authenticate_user(message);
    });
}

void TelegramBot::start() {
    bot.getEvents().onAnyMessage([&](TgBot::Message::Ptr message) {
        std::cout << "Получено сообщение: " << message->text << std::endl;
    });

    TgBot::TgLongPoll longPoll(bot);

    while (true) {
        std::cout << "Ожидание сообщений..." << std::endl;
        longPoll.start();
    }
}

void TelegramBot::send_registration_message(TgBot::Message::Ptr message) {
    int64_t telegram_id = message->chat->id;
    bot.getApi().sendMessage(telegram_id, "Вы не зарегистрированы! Пожалуйста, зарегистрируйтесь.");
}


void TelegramBot::authenticate_user(TgBot::Message::Ptr message) {
    int64_t telegram_id = message->chat->id;
    std::string name = message->from->firstName;

    // Проверка, зарегистрирован ли пользователь
    if (auth_manager->authenticate(telegram_id)) {
        send_welcome_message(message);
    } else {
        // Если пользователь не зарегистрирован, регистрируем его
        auth_manager->register_user(telegram_id, name);
        send_registration_message(message);
    }
}

void TelegramBot::send_welcome_message(TgBot::Message::Ptr message) {
    int64_t telegram_id = message->chat->id;
    std::shared_ptr<Player> player = auth_manager->get_player(telegram_id);
    
    if (player) {
        bot.getApi().sendMessage(telegram_id, "Привет, " + player->get_name() + "!");
    } else {
        bot.getApi().sendMessage(telegram_id, "Пожалуйста, зарегистрируйтесь.");
    }
}
