#include "tg_bot.h"
#include <string>

int main() {
    std::string token = "7945815181:AAHAzN3QI5dUtq7iSmw9if2rrA5Rzi2j3bY";
    TelegramBot bot(token);
    bot.start();
    return 0;
}