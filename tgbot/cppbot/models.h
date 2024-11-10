#ifndef MODELS_H
#define MODELS_H

#include <iostream>
#include <string>
#include <vector>
#include <memory>
#include <ctime>

class Team;

class Player : public std::enable_shared_from_this<Player> {
public:
    Player(const std::string& name, int64_t telegram_id);
    const std::string& get_name() const;
    bool has_team() const;
    std::shared_ptr<Team> get_team() const;
    bool get_is_owner() const;
    int64_t get_telegram_id() const;
    void join_team(std::shared_ptr<Team> new_team);
    void create_team(const std::string& team_name);
    bool is_captain() const;
    void register_for_match(std::shared_ptr<class Match> match);

private:
    std::string name;              
    int64_t telegram_id;          
    std::shared_ptr<Team> team;     
    bool is_owner;
};

class Team {
public:
    Team(const std::string& name, std::shared_ptr<Player> owner);
    const std::string& get_name() const;
    std::shared_ptr<Player> get_owner() const;
    std::shared_ptr<Player> get_captain() const;
    void set_captain(std::shared_ptr<Player> player);
    void add_member(std::shared_ptr<Player> player);
    bool is_member(std::shared_ptr<Player> player) const;
    const std::vector<std::shared_ptr<Player>>& get_members() const;

private:
    std::string name;
    std::shared_ptr<Player> owner;
    std::shared_ptr<Player> captain;
    std::vector<std::shared_ptr<Player>> members;
};

class Match {
public:
    Match(const std::time_t& time);
    void add_team(std::shared_ptr<Team> team);
    const std::time_t& get_match_time() const;
    const std::vector<std::shared_ptr<Team>>& get_teams() const;

private:
    std::time_t match_time;
    std::vector<std::shared_ptr<Team>> teams;
};

#endif // MODELS_H