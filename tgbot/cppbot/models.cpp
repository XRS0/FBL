#include "models.h"
#include <iostream>
#include <algorithm>

Player::Player(const std::string& name, int64_t telegram_id) 
    : name(name), team(nullptr), telegram_id(telegram_id), is_owner(false) {}

const std::string& Player::get_name() const {
    return name;
}

bool Player::has_team() const {
    return team != nullptr;
}

std::shared_ptr<Team> Player::get_team() const {
    return team;
}

bool Player::get_is_owner() const {
    return is_owner;
}

int64_t Player::get_telegram_id() const {
    return telegram_id;
}

void Player::join_team(std::shared_ptr<Team> new_team) {
    if (!has_team()) {
        team = new_team;
        new_team->add_member(shared_from_this());
    } else {
        std::cerr << "Игрок уже состоит в команде!" << std::endl;
    }
}

void Player::create_team(const std::string& team_name) {
    if (!has_team()) {
        team = std::make_shared<Team>(team_name, shared_from_this());
        is_owner = true;
    } else {
        std::cerr << "Игрок уже состоит в команде!" << std::endl;
    }
}

bool Player::is_captain() const {
    return team && team->get_captain() == shared_from_this();
}

void Player::register_for_match(std::shared_ptr<class Match> match) {
    if (is_captain()) {
        match->add_team(team);
    } else {
        std::cerr << "Только капитан команды может записать команду на матч!" << std::endl;
    }
}

Team::Team(const std::string& name, std::shared_ptr<Player> owner) 
    : name(name), owner(owner), captain(owner) {
    members.push_back(owner);
}

const std::string& Team::get_name() const {
    return name;
}

std::shared_ptr<Player> Team::get_owner() const {
    return owner;
}

std::shared_ptr<Player> Team::get_captain() const {
    return captain;
}

void Team::set_captain(std::shared_ptr<Player> player) {
    if (is_member(player)) {
        captain = player;
    } else {
        std::cerr << "Игрок не является членом команды!" << std::endl;
    }
}

void Team::add_member(std::shared_ptr<Player> player) {
    if (!is_member(player)) {
        members.push_back(player);
    } else {
        std::cerr << "Игрок уже в команде!" << std::endl;
    }
}

bool Team::is_member(std::shared_ptr<Player> player) const {
    return std::find_if(members.begin(), members.end(), [&](const std::shared_ptr<Player>& member) {
        return member == player;
    }) != members.end();
}

const std::vector<std::shared_ptr<Player>>& Team::get_members() const {
    return members;
}

Match::Match(const std::time_t& time) : match_time(time) {}

void Match::add_team(std::shared_ptr<Team> team) {
    teams.push_back(team);
}

const std::time_t& Match::get_match_time() const {
    return match_time;
}

const std::vector<std::shared_ptr<Team>>& Match::get_teams() const {
    return teams;
}