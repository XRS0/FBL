// models.go
package pkg

import "time"

// Player представляет игрока в системе
type Player struct {
	Name       string
	TelegramID int64
	Team       *Team // Ссылка на команду, в которой состоит игрок (может быть nil)
	IsOwner    bool  // Флаг, указывающий, что игрок является владельцем команды
	IsCaptain  bool  // Флаг, указывающий, что игрок является капитаном команды
}

// Team представляет команду
type Team struct {
	Name      string
	Owner     *Player   // Владелец команды (создатель)
	Captain   *Player   // Капитан команды, который может записывать команду на матч
	Members   []*Player // Список членов команды
	MatchList []*Match  // Список матчей, в которых участвует команда
}

// Match представляет матч
type Match struct {
	Time  time.Time // Время, когда матч должен состояться
	TeamA *Team     // Команда A
	TeamB *Team     // Команда B
}

// NewPlayer создаёт нового игрока
func NewPlayer(name string, telegramID int64) *Player {
	return &Player{
		Name:       name,
		TelegramID: telegramID,
	}
}

// NewTeam создаёт новую команду и назначает игрока владельцем
func NewTeam(name string, owner *Player) *Team {
	team := &Team{
		Name:    name,
		Owner:   owner,
		Captain: owner, // Владелец команды становится капитаном
		Members: []*Player{owner},
	}
	owner.Team = team
	owner.IsOwner = true
	owner.IsCaptain = true
	return team
}

// AddPlayer добавляет игрока в команду
func (t *Team) AddPlayer(player *Player) bool {
	if player.Team != nil {
		return false // Игрок уже состоит в другой команде
	}
	player.Team = t
	t.Members = append(t.Members, player)
	return true
}

// SetCaptain назначает капитана команды
func (t *Team) SetCaptain(player *Player) bool {
	for _, member := range t.Members {
		if member == player {
			t.Captain = player
			player.IsCaptain = true
			return true
		}
	}
	return false // Игрок не является членом команды
}

// ScheduleMatch записывает команду на матч
func ScheduleMatch(time time.Time, teamA *Team, teamB *Team) *Match {
	match := &Match{
		Time:  time,
		TeamA: teamA,
		TeamB: teamB,
	}
	teamA.MatchList = append(teamA.MatchList, match)
	teamB.MatchList = append(teamB.MatchList, match)
	return match
}
